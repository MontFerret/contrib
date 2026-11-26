package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/MontFerret/api"
	"github.com/MontFerret/ferret/v2"
	"github.com/MontFerret/ferret/v2/uapi"
	"github.com/MontFerret/wire/client"
)

func TestActiveExecutionCleanup(t *testing.T) {
	for _, reason := range []string{"transport lost", "host stopped"} {
		t.Run(reason, func(t *testing.T) {
			ctx, timeout := context.WithTimeout(t.Context(), 15*time.Second)
			defer timeout()
			entered := make(chan struct{})
			events := make(chan string, 3)
			planErr := errors.New("plan close error")
			closeErr := errors.New("engine close error")

			hosted, err := uapi.New("test",
				ferret.WithBeforeRunHook(func(ctx context.Context) (context.Context, error) {
					close(entered)
					<-ctx.Done()

					return ctx, ctx.Err()
				}),
				ferret.WithSessionCloseHook(func() error {
					events <- "session"

					return nil
				}),
				ferret.WithPlanCloseHook(func() error {
					events <- "plan"

					return planErr
				}),
				ferret.WithEngineCloseHook(func() error {
					events <- "engine"

					return closeErr
				}),
			)
			if err != nil {
				t.Fatal(err)
			}

			addr, stop := startTestHost(t, ctx, hosted)
			conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				t.Fatal(err)
			}

			defer conn.Close()

			remote, err := client.From(ctx, conn)
			if err != nil {
				t.Fatal(err)
			}

			defer remote.Close()

			plan, err := remote.Compile(ctx, api.NewAnonymousSource("RETURN 42"))
			if err != nil {
				t.Fatal(err)
			}

			defer plan.Close()

			session, err := plan.NewSession(ctx)
			if err != nil {
				t.Fatal(err)
			}

			defer session.Close()

			finished := make(chan error, 1)
			go func() {
				output, err := session.Run(ctx)
				if err == nil {
					_, err = output.Collect(ctx)
					err = errors.Join(err, output.Close())
				}

				finished <- err
			}()

			select {
			case <-entered:
			case <-ctx.Done():
				t.Fatal("execution did not start")
			}

			var serveErr error
			if reason == "transport lost" {
				if err := conn.Close(); err != nil {
					t.Fatal(err)
				}
			} else {
				serveErr = stop()
			}

			for _, want := range []string{"session", "plan"} {
				select {
				case got := <-events:
					if got != want {
						t.Fatalf("cleanup = %q, want %q", got, want)
					}
				case <-ctx.Done():
					t.Fatalf("timed out waiting for %s cleanup", want)
				}
			}

			select {
			case err := <-finished:
				if err == nil {
					t.Fatal("interrupted execution succeeded")
				}
			case <-ctx.Done():
				t.Fatal("client did not observe interruption")
			}

			if reason == "transport lost" {
				next, err := client.New(ctx, addr, client.WithInsecure())
				if err != nil {
					t.Fatalf("reconnect: %v", err)
				}

				if err := next.Close(); err != nil {
					t.Fatal(err)
				}

				serveErr = stop()
			}

			if reason == "host stopped" && !errors.Is(serveErr, planErr) {
				t.Fatalf("host lost Wire cleanup failure: %v", serveErr)
			}

			select {
			case got := <-events:
				if got != "engine" || !errors.Is(serveErr, closeErr) {
					t.Fatalf("final cleanup = %q, host error = %v", got, serveErr)
				}
			default:
				t.Fatal("host returned without closing the engine")
			}
		})
	}
}

func startTestHost(t *testing.T, ctx context.Context, hosted *uapi.Runtime) (string, func() error) {
	t.Helper()
	ctx, cancel := context.WithCancel(ctx)
	reader, writer := io.Pipe()
	defer reader.Close()

	done := make(chan struct{})
	var serveErr error
	go func() {
		defer close(done)
		defer writer.Close()
		serveErr = serve(ctx, hosted, "127.0.0.1:0", writer)
	}()

	stop := func() error {
		t.Helper()
		cancel()

		select {
		case <-done:
			return serveErr
		case <-time.After(5 * time.Second):
			t.Fatal("host did not stop")

			return nil
		}
	}
	t.Cleanup(func() { _ = stop() })

	line, err := bufio.NewReader(reader).ReadString('\n')
	if err != nil {
		t.Fatalf("read endpoint: %v, host error: %v", err, stop())
	}

	return endpointAddress(t, line), stop
}

func TestAnnouncementFailure(t *testing.T) {
	closeErr := errors.New("engine close error")
	hosted, err := uapi.New("test", ferret.WithEngineCloseHook(func() error { return closeErr }))
	if err != nil {
		t.Fatal(err)
	}

	var announcement bytes.Buffer
	reader, writer := io.Pipe()
	_ = reader.CloseWithError(io.ErrShortWrite)
	defer writer.Close()

	err = serve(t.Context(), hosted, "127.0.0.1:0", io.MultiWriter(&announcement, writer))
	if !errors.Is(err, io.ErrShortWrite) || !errors.Is(err, closeErr) {
		t.Fatalf("lost announcement or cleanup failure: %v", err)
	}

	addr := endpointAddress(t, announcement.String())
	rebound, err := net.Listen("tcp4", addr)
	if err != nil {
		t.Fatalf("listener still bound after failed announcement: %v", err)
	}

	rebound.Close()
}
