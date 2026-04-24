package network

import (
	"context"
	"errors"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/rs/zerolog"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type sessionNavigationEventStream struct {
	logger  zerolog.Logger
	client  *cdp.Client
	onFrame page.FrameNavigatedClient
	onDoc   page.NavigatedWithinDocumentClient
}

func newSessionNavigationEventStream(
	logger zerolog.Logger,
	client *cdp.Client,
	onFrame page.FrameNavigatedClient,
	onDoc page.NavigatedWithinDocumentClient,
) runtime.Stream {
	return &sessionNavigationEventStream{
		logger:  logger,
		client:  client,
		onFrame: onFrame,
		onDoc:   onDoc,
	}
}

func (s *sessionNavigationEventStream) Close() error {
	return errors.Join(
		s.onFrame.Close(),
		s.onDoc.Close(),
	)
}

func (s *sessionNavigationEventStream) Read(ctx context.Context) <-chan runtime.Message {
	ch := make(chan runtime.Message)

	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				return
			case <-s.onDoc.Ready():
				if ctx.Err() != nil {
					return
				}

				repl, err := s.onDoc.Recv()
				if err != nil {
					select {
					case <-ctx.Done():
					case ch <- runtime.NewErrorMessage(err):
					}
					s.logger.Trace().Err(err).Msg("failed to read data from within document navigation event stream")
					return
				}

				evt := NavigationEvent{
					URL:          repl.URL,
					FrameID:      repl.FrameID,
					sourceClient: s.client,
				}

				s.logger.Trace().
					Str("url", evt.URL).
					Str("frame_id", string(evt.FrameID)).
					Str("type", evt.MimeType).
					Msg("received within document navigation event")

				select {
				case <-ctx.Done():
					return
				case ch <- runtime.NewValueMessage(&evt):
				}
			case <-s.onFrame.Ready():
				if ctx.Err() != nil {
					return
				}

				repl, err := s.onFrame.Recv()
				if err != nil {
					select {
					case <-ctx.Done():
					case ch <- runtime.NewErrorMessage(err):
					}
					s.logger.Trace().Err(err).Msg("failed to read data from frame navigation event stream")
					return
				}

				evt := NavigationEvent{
					URL:          repl.Frame.URL,
					FrameID:      repl.Frame.ID,
					MimeType:     repl.Frame.MimeType,
					sourceClient: s.client,
				}

				s.logger.Trace().
					Str("url", evt.URL).
					Str("frame_id", string(evt.FrameID)).
					Str("type", evt.MimeType).
					Msg("received frame navigation event")

				select {
				case <-ctx.Done():
					return
				case ch <- runtime.NewValueMessage(&evt):
				}
			}
		}
	}()

	return ch
}
