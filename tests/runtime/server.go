package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/MontFerret/ferret/v2/uapi"
	"github.com/MontFerret/wire/server"
)

// serve owns hosted. Wire borrows it until Shutdown finishes settling its connections.
func serve(ctx context.Context, hosted *uapi.Runtime, address string, stdout io.Writer) (err error) {
	defer func() { err = errors.Join(err, hosted.Close()) }()

	wire, err := server.New(hosted)
	if err != nil {
		return err
	}

	defer func() { err = errors.Join(err, shutdown(wire)) }()

	listener, err := new(net.ListenConfig).Listen(ctx, "tcp4", address)
	if err != nil {
		return fmt.Errorf("bind Wire listener: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return errors.Join(err, listener.Close())
	}

	if _, err := fmt.Fprintf(stdout, "wire+tcp://%s\n", listener.Addr()); err != nil {
		return errors.Join(fmt.Errorf("announce Wire endpoint: %w", err), listener.Close())
	}

	// Close the listener first so that no new transports enter while Wire drains.
	closed := make(chan error, 1)
	stop := context.AfterFunc(ctx, func() { closed <- listener.Close() })
	err = wire.Serve(context.WithoutCancel(ctx), listener)
	if !stop() {
		closeErr := <-closed
		if errors.Is(err, net.ErrClosed) {
			err = nil
		}

		if !errors.Is(closeErr, net.ErrClosed) {
			err = errors.Join(err, closeErr)
		}
	}

	return err
}

func shutdown(wire *server.Server) error {
	drain, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	shutdownErr := wire.Shutdown(drain)
	if errors.Is(shutdownErr, context.DeadlineExceeded) {
		// A deadline forces the transport to stop, but hosted cleanup can outlive it.
		shutdownErr = errors.Join(shutdownErr, wire.Shutdown(context.Background()))
	}

	return shutdownErr
}
