package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/MontFerret/barn/pkg/publish"
)

func main() {
	options, err := parseOptions(os.Args[1:], os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "configure publisher: %v\n", err)
		os.Exit(2)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := execute(ctx, options, publish.Prepare, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "publish module: %v\n", err)
		os.Exit(1)
	}
}
