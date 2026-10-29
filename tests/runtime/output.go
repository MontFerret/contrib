package main

import (
	"context"
	"errors"
	"io"

	"github.com/MontFerret/ferret/v2"
)

func writeResult(ctx context.Context, dst io.Writer, output ferret.Output) error {
	content, err := collectOutput(ctx, output)
	if err != nil {
		return err
	}

	if content != nil {
		n, err := dst.Write(content.Data)
		if err != nil {
			return err
		}

		if n != len(content.Data) {
			return io.ErrShortWrite
		}
	}

	n, err := io.WriteString(dst, "\n")
	if err != nil {
		return err
	}

	if n != 1 {
		return io.ErrShortWrite
	}

	return nil
}

func collectOutput(ctx context.Context, output ferret.Output) (content *ferret.Content, err error) {
	defer func() { err = joinErrors(err, output.Close()) }()

	return output.Collect(ctx)
}

// Preserve native diagnostic formatting when only one operation failed.
func joinErrors(first, second error) error {
	if first == nil {
		return second
	}

	if second == nil {
		return first
	}

	return errors.Join(first, second)
}
