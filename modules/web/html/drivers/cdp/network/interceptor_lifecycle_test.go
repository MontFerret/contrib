package network_test

import (
	"context"
	"errors"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/fetch"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/network"
)

func TestInterceptorStartupAndShutdown(t *testing.T) {
	for _, fail := range []string{"", "subscribe", "enable", "stream"} {
		t.Run(fail, func(t *testing.T) {
			var subscribed atomic.Bool
			var disables atomic.Int32
			stream := NewRequestPausedClient()
			failure := errors.New("interception failed")
			api := &FetchAPI{
				requestPaused: func(context.Context) (fetch.RequestPausedClient, error) {
					if fail == "subscribe" {
						return nil, failure
					}

					subscribed.Store(true)

					return stream, nil
				},
				enable: func(context.Context, *fetch.EnableArgs) error {
					if !subscribed.Load() {
						t.Error("Fetch enabled before subscribing")
					}

					if fail == "enable" {
						return failure
					}

					return nil
				},
				disable: func(context.Context) error {
					disables.Add(1)

					return nil
				},
			}
			interceptor := network.NewInterceptor(zerolog.Nop(), &cdp.Client{Fetch: api})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := interceptor.Run(ctx)
			if fail == "subscribe" || fail == "enable" {
				if !errors.Is(err, failure) {
					t.Fatalf("startup error = %v", err)
				}

				_ = interceptor.Close()

				return
			}

			if err != nil {
				t.Fatal(err)
			}

			if fail == "stream" {
				_ = stream.Close()
			} else {
				cancel()
			}

			select {
			case <-interceptor.Done():
			case <-time.After(time.Second):
				t.Fatal("interception did not stop")
			}

			if interceptor.IsRunning() {
				t.Fatal("stopped interceptor reports running")
			}

			if fail == "stream" {
				if !errors.Is(interceptor.Err(), io.EOF) || disables.Load() != 0 {
					t.Fatalf("stream failure must leave Fetch armed: err=%v disables=%d", interceptor.Err(), disables.Load())
				}
			} else if disables.Load() != 1 {
				t.Fatalf("Fetch disabled %d times", disables.Load())
			}

			_ = interceptor.Close()
			_ = interceptor.Close()
		})
	}
}
