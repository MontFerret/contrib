package session

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/target"
	"github.com/mafredri/cdp/rpcc"
)

const (
	maxSetupEvents = 1024
	maxSetupBytes  = 4 << 20
)

var errSetupOverflow = errors.New("CDP child initialization event buffer exceeded")

type targetSession struct {
	router     *router
	id, parent target.SessionID
	inbox      *mailbox[rpcc.Response]
	// Guarded by router.mu. Domain enable may emit events before it replies.
	setup       context.Context
	abort       context.CancelCauseFunc
	timer       *time.Timer
	bufferBytes int
	holding     bool
	buffer      []rpcc.Response
}

func (s *targetSession) receive(msg rpcc.Response) {
	if s.holding && msg.Method != "" {
		if s.timer != nil {
			size := len(msg.Method) + len(msg.Args)
			if s.setup.Err() != nil {
				return
			}

			if len(s.buffer) >= maxSetupEvents || size > maxSetupBytes-s.bufferBytes {
				s.failSetup(errSetupOverflow)

				return
			}

			s.bufferBytes += size
		}

		s.buffer = append(s.buffer, msg)

		return
	}

	s.inbox.put(msg)
}

func (s *targetSession) releaseEvents() error {
	s.router.mu.Lock()
	defer s.router.mu.Unlock()

	if err := s.finishSetup(); err != nil {
		return err
	}

	for _, msg := range s.buffer {
		if err := s.router.frameEvent(s.id, &msg); err != nil {
			return err
		}

		s.inbox.put(msg)
	}

	s.buffer = nil
	s.holding = false

	return nil
}

func (s *targetSession) client(ctx context.Context, info target.Info, popup bool) (*Client, error) {
	conn, err := rpcc.DialContext(ctx, "", rpcc.WithDialer(func(context.Context, string) (io.ReadWriteCloser, error) {
		return &closeConn{close: s.detach}, nil
	}), rpcc.WithCodec(func(io.ReadWriter) rpcc.Codec { return s }))
	if err != nil {
		return nil, err
	}

	return &Client{
		ID:       s.id,
		TargetID: info.TargetID,
		Info:     info,
		Conn:     conn,
		CDP:      cdp.NewClient(conn),
		popup:    popup,
		closeFn: sync.OnceValue(func() error {
			if conn.Context().Err() != nil {
				return nil
			}

			return conn.Close()
		}),
	}, nil
}

func (s *targetSession) detach() error {
	r := s.router
	r.mu.Lock()

	active := r.sessions[s.id] == s
	if active {
		r.remove(s)
		r.targets.events.put(targetChange{detached: s.id})
	}

	r.mu.Unlock()

	if !active {
		return nil
	}

	parent := r.parentClient(s)
	if parent == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultDetachTimeout)
	defer cancel()

	err := parent.Target.DetachFromTarget(ctx, target.NewDetachFromTargetArgs().SetSessionID(s.id))
	if isMissingSessionError(err) || errors.Is(err, rpcc.ErrConnClosing) {
		return nil
	}

	return err
}

func (s *targetSession) WriteRequest(r *rpcc.Request) error {
	return s.router.write(s.id, r)
}

func (s *targetSession) ReadResponse(out *rpcc.Response) error {
	msg, err := s.inbox.next()
	*out = msg

	return err
}

// Called under the router lock. Replies bypass the early-event buffer.
func (s *targetSession) stopSetup(cause error) {
	if s.timer != nil {
		s.timer.Stop()
		s.abort(cause)
	}
}

// Called with the router locked so an expiry, detach, or successful release wins once.
func (s *targetSession) failSetup(cause error) {
	if s.timer == nil || s.setup.Err() != nil {
		return
	}

	s.stopSetup(cause)
	s.buffer = nil
	s.router.targets.events.put(targetChange{session: s, err: cause})
}

func (s *targetSession) finishSetup() error {
	if s.timer == nil {
		return nil
	}

	if !s.timer.Stop() && s.setup.Err() == nil {
		s.abort(context.DeadlineExceeded)
	}

	err := context.Cause(s.setup)
	s.abort(context.Canceled)
	if err == nil {
		s.timer = nil
	}

	return err
}
