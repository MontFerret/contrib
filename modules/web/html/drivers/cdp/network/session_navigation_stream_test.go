package network

import (
	"context"
	"errors"
	"testing"

	"github.com/mafredri/cdp/protocol/page"
	"github.com/rs/zerolog"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type nilNavigationStream struct {
	ready chan struct{}
}

func newNilNavigationStream(trigger bool) *nilNavigationStream {
	ready := make(chan struct{})
	if trigger {
		close(ready)
	}

	return &nilNavigationStream{ready: ready}
}

func (stream *nilNavigationStream) Ready() <-chan struct{} {
	return stream.ready
}

func (stream *nilNavigationStream) RecvMsg(any) error {
	return nil
}

func (stream *nilNavigationStream) Close() error {
	return nil
}

type nilFrameNavigationStream struct {
	*nilNavigationStream
}

func (stream *nilFrameNavigationStream) Recv() (*page.FrameNavigatedReply, error) {
	return nil, nil
}

type nilWithinDocumentNavigationStream struct {
	*nilNavigationStream
}

func (stream *nilWithinDocumentNavigationStream) Recv() (*page.NavigatedWithinDocumentReply, error) {
	return nil, nil
}

func TestSessionNavigationStreamRejectsNilReplies(t *testing.T) {
	for _, tt := range []struct {
		name         string
		triggerFrame bool
	}{
		{name: "frame", triggerFrame: true},
		{name: "within document"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			frame := &nilFrameNavigationStream{newNilNavigationStream(tt.triggerFrame)}
			within := &nilWithinDocumentNavigationStream{newNilNavigationStream(!tt.triggerFrame)}
			stream := newSessionNavigationEventStream(zerolog.Nop(), nil, frame, within)

			messages := stream.Read(context.Background())
			message, ok := <-messages
			if !ok {
				t.Fatal("stream terminated without an error message")
			}
			if !errors.Is(message.Err(), runtime.ErrUnexpected) {
				t.Fatalf("message error = %v, want unexpected", message.Err())
			}
			if _, ok := <-messages; ok {
				t.Fatal("stream did not terminate after the nil reply")
			}
		})
	}
}
