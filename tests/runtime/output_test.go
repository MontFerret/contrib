package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/MontFerret/ferret/v2"
)

func TestWriteResult(t *testing.T) {
	terminalErr := errors.New("encoding failed")
	cleanupErr := errors.New("cleanup failed")

	for _, test := range []struct {
		name       string
		content    *ferret.Content
		collectErr error
		closeErr   error
		want       string
	}{
		{name: "JSON", content: &ferret.Content{Data: []byte(`{"ok":true}`)}, want: "{\"ok\":true}\n"},
		{name: "absent", want: "\n"},
		{name: "empty", content: &ferret.Content{}, want: "\n"},
		{name: "partial encoding failure", content: &ferret.Content{Data: []byte(`{"ok":`)}, collectErr: terminalErr},
		{name: "cleanup failure", content: &ferret.Content{Data: []byte(`true`)}, closeErr: cleanupErr},
		{name: "combined failure", collectErr: terminalErr, closeErr: cleanupErr},
	} {
		t.Run(test.name, func(t *testing.T) {
			output := &outputStub{content: test.content, collectErr: test.collectErr, closeErr: test.closeErr}
			var dst bytes.Buffer

			err := writeResult(t.Context(), &dst, output)
			if test.collectErr == nil && test.closeErr == nil && err != nil {
				t.Fatal(err)
			}

			for _, cause := range []error{test.collectErr, test.closeErr} {
				if cause != nil && !errors.Is(err, cause) {
					t.Fatalf("error %v does not preserve %v", err, cause)
				}
			}

			if dst.String() != test.want || !output.closed {
				t.Fatalf("output = %q, closed = %v; want %q and closed", dst.String(), output.closed, test.want)
			}
		})
	}
}

func TestWriteResultCanceledContextClosesOutput(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	output := &outputStub{content: &ferret.Content{Data: []byte("true")}}
	var dst bytes.Buffer

	err := writeResult(ctx, &dst, output)
	if !errors.Is(err, context.Canceled) || !output.closed || dst.Len() != 0 {
		t.Fatalf("error = %v, closed = %v, output = %q", err, output.closed, dst.String())
	}
}

func TestWriteResultPropagatesWriterFailures(t *testing.T) {
	writeErr := errors.New("writer failed")

	for _, test := range []struct {
		name string
		call int
		err  error
		want error
	}{
		{name: "payload", call: 1, err: writeErr, want: writeErr},
		{name: "newline", call: 2, err: writeErr, want: writeErr},
		{name: "short payload", call: 1, want: io.ErrShortWrite},
		{name: "short newline", call: 2, want: io.ErrShortWrite},
	} {
		t.Run(test.name, func(t *testing.T) {
			output := &outputStub{content: &ferret.Content{Data: []byte("true")}}
			dst := &resultWriter{err: test.err, failCall: test.call}

			err := writeResult(t.Context(), dst, output)
			if !errors.Is(err, test.want) || !output.closed {
				t.Fatalf("error = %v, closed = %v", err, output.closed)
			}
		})
	}
}
