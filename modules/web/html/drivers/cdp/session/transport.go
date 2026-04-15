package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/target"
	"github.com/mafredri/cdp/rpcc"
)

const defaultDetachTimeout = 5 * time.Second

type transport struct {
	recvC    chan []byte
	send     func([]byte) error
	init     chan struct{}
	conn     *rpcc.Conn
	id       target.SessionID
	targetID target.ID
	detached atomic.Bool
}

func newTransport(
	ctx context.Context,
	browserClient *cdp.Client,
	targetID target.ID,
	sessionID target.SessionID,
) (*transport, error) {
	s := &transport{
		id:       sessionID,
		targetID: targetID,
		recvC:    make(chan []byte, 1),
		init:     make(chan struct{}),
	}

	s.send = func(data []byte) error {
		<-s.init

		//lint:ignore SA1019 CDP session transport still tunnels messages through Target.SendMessageToTarget for attached sessions.
		return browserClient.Target.SendMessageToTarget(
			s.conn.Context(),
			target.NewSendMessageToTargetArgs(string(data)).SetSessionID(s.id),
		)
	}

	detach := func() error {
		if !s.detached.CompareAndSwap(false, true) {
			return nil
		}

		detachCtx, cancel := context.WithTimeout(context.Background(), defaultDetachTimeout)
		defer cancel()

		err := browserClient.Target.DetachFromTarget(
			detachCtx,
			target.NewDetachFromTargetArgs().SetSessionID(s.id),
		)
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("session: detach timed out for session %s", s.id)
		}
		if isMissingSessionError(err) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("session: detach failed for session %s: %w", s.id, err)
		}

		return nil
	}

	conn, err := rpcc.DialContext(
		ctx,
		"",
		rpcc.WithDialer(func(_ context.Context, _ string) (io.ReadWriteCloser, error) {
			return &closeConn{close: detach}, nil
		}),
		rpcc.WithCodec(func(_ io.ReadWriter) rpcc.Codec {
			return s
		}),
	)
	if err != nil {
		return nil, err
	}

	s.conn = conn
	close(s.init)

	return s, nil
}

func (s *transport) WriteRequest(req *rpcc.Request) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.send(data)
}

func (s *transport) ReadResponse(resp *rpcc.Response) error {
	<-s.init

	select {
	case message := <-s.recvC:
		return json.Unmarshal(message, resp)
	case <-s.conn.Context().Done():
		return s.conn.Context().Err()
	}
}

func (s *transport) Write(data []byte) error {
	select {
	case s.recvC <- data:
		return nil
	case <-s.conn.Context().Done():
		return s.conn.Context().Err()
	}
}

func (s *transport) Close() error {
	if s.conn == nil {
		return nil
	}

	return s.conn.Close()
}

func (s *transport) MarkDetached() {
	s.detached.Store(true)
}

func (s *transport) Conn() *rpcc.Conn {
	return s.conn
}
