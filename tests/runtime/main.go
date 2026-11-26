package main

import (
	"context"
	"flag"
	"fmt"
	"net/netip"
	"os"
	"os/signal"
	"syscall"

	"github.com/MontFerret/ferret/v2"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:0", "Wire listener address (127.0.0.1:port)")
	browser := flag.String("browser-address", "http://127.0.0.1:9222", "CDP address")
	level := flag.String("log-level", ferret.LogError.String(), "engine log level")
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Fprintln(os.Stderr, "unexpected positional arguments: submit queries through Wire")
		os.Exit(2)
	}

	if err := run(*listen, *browser, *level); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(listen, browser, level string) error {
	address, err := netip.ParseAddrPort(listen)
	if err != nil || address.Addr().String() != "127.0.0.1" {
		return fmt.Errorf("invalid listen address %q: use 127.0.0.1 with a port from 0 to 65535", listen)
	}

	logLevel, err := ferret.ParseLogLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level %q: %w", level, err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	hosted, err := newRuntime(browser, logLevel, os.Stderr)
	if err != nil {
		return fmt.Errorf("initialize Ferret runtime: %w", err)
	}

	return serve(ctx, hosted, address.String(), os.Stdout)
}
