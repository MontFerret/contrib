package session

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/protocol/target"
	"github.com/mafredri/cdp/rpcc"
)

type router struct {
	decoder  *json.Decoder
	encoder  *json.Encoder
	writeMu  sync.Mutex
	mu       sync.Mutex
	sessions map[target.SessionID]*targetSession
	pending  map[uint64]pendingCall
	nextID   uint64
	targets  *targetTracker
	err      error
}

type pendingCall struct {
	session target.SessionID
	id      uint64
}

func (r *router) WriteRequest(req *rpcc.Request) error {
	return r.write("", req)
}

func (r *router) write(id target.SessionID, req *rpcc.Request) error {
	r.writeMu.Lock()
	defer r.writeMu.Unlock()

	r.mu.Lock()

	if r.err != nil {
		err := r.err
		r.mu.Unlock()

		return err
	}

	if id != "" && r.sessions[id] == nil {
		r.mu.Unlock()

		return rpcc.ErrConnClosing
	}

	r.nextID++
	wireID := r.nextID
	r.pending[wireID] = pendingCall{session: id, id: req.ID}
	r.mu.Unlock()

	err := r.encoder.Encode(struct {
		ID        uint64           `json:"id"`
		Method    string           `json:"method"`
		Args      any              `json:"params,omitempty"`
		SessionID target.SessionID `json:"sessionId,omitempty"`
	}{wireID, req.Method, req.Args, id})
	if err != nil {
		r.mu.Lock()
		delete(r.pending, wireID)
		r.mu.Unlock()
	}

	return err
}

func (r *router) ReadResponse(out *rpcc.Response) error {
	for {
		var packet struct {
			rpcc.Response
			SessionID target.SessionID `json:"sessionId"`
		}

		if err := r.decoder.Decode(&packet); err != nil {
			r.shutdown(err)

			return err
		}

		r.mu.Lock()
		id := packet.SessionID

		if packet.ID != 0 {
			call, ok := r.pending[packet.ID]
			if !ok {
				r.mu.Unlock()

				continue
			}

			// Chromium can omit sessionId on a browser-generated protocol error.
			if id != call.session && !(id == "" && packet.Error != nil) {
				r.mu.Unlock()
				err := errors.New("CDP response has a conflicting session ID")
				r.shutdown(err)

				return err
			}

			delete(r.pending, packet.ID)
			id, packet.ID = call.session, call.id
		}

		if r.targets != nil && r.targets.manager.ctx.Err() == nil {
			if err := r.targetEvent(id, &packet.Response); err != nil {
				r.mu.Unlock()
				r.shutdown(err)

				return err
			}
		}

		s := r.sessions[id]
		if s != nil {
			s.receive(packet.Response)
		}

		r.mu.Unlock()

		if id == "" {
			*out = packet.Response

			return nil
		}
	}
}

// Called by the sole reader with mu held, before reading any traffic from
// the newly attached session. Target initialization runs off the reader.
func (r *router) targetEvent(parent target.SessionID, msg *rpcc.Response) error {
	switch msg.Method {
	case "Page.frameDetached", "Page.frameNavigated":
		if s := r.sessions[parent]; s != nil && !s.holding {
			return r.frameEvent(parent, msg)
		}
	case "Target.attachedToTarget":
		var event target.AttachedToTargetReply

		if err := json.Unmarshal(msg.Args, &event); err != nil {
			return err
		}

		if event.SessionID == "" || event.TargetInfo.TargetID == "" {
			return errors.New("CDP attachment is missing its session or target ID")
		}

		if r.sessions[event.SessionID] == nil {
			s := &targetSession{id: event.SessionID, parent: parent, router: r, inbox: newMailbox[rpcc.Response](), holding: true}
			if event.TargetInfo.TargetID != r.targets.manager.rootTargetID {
				s.setup, s.abort = context.WithCancelCause(r.targets.manager.ctx)
				s.timer = time.AfterFunc(r.targets.setupTimeout, func() {
					r.mu.Lock()
					defer r.mu.Unlock()
					s.failSetup(context.DeadlineExceeded)
				})
			}

			r.sessions[s.id] = s
			r.targets.events.put(targetChange{session: s, info: event.TargetInfo, waiting: event.WaitingForDebugger})
		}
	case "Target.detachedFromTarget":
		var event target.DetachedFromTargetReply

		if err := json.Unmarshal(msg.Args, &event); err != nil {
			return err
		}

		if s := r.sessions[event.SessionID]; s != nil {
			r.remove(s)
			r.targets.events.put(targetChange{detached: s.id})
		}
	case "Target.targetCreated", "Target.targetInfoChanged":
		var event target.CreatedReply

		if err := json.Unmarshal(msg.Args, &event); err != nil {
			return err
		}

		if event.TargetInfo.Type == "page" && event.TargetInfo.OpenerID != nil {
			r.targets.events.put(targetChange{info: event.TargetInfo})
		}
	case "Target.targetDestroyed":
		var event target.DestroyedReply

		if err := json.Unmarshal(msg.Args, &event); err != nil {
			return err
		}

		r.targets.events.put(targetChange{destroyed: event.TargetID})
	}

	return nil
}

func (r *router) remove(s *targetSession) {
	delete(r.sessions, s.id)
	for wireID, call := range r.pending {
		if call.session == s.id {
			delete(r.pending, wireID)
		}
	}

	s.stopSetup(rpcc.ErrConnClosing)
	s.inbox.stop(rpcc.ErrConnClosing)
	s.buffer = nil
}

func (r *router) shutdown(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.err = err
	clear(r.pending)
	for _, s := range r.sessions {
		s.stopSetup(err)
		s.inbox.stop(err)
		s.buffer = nil
	}

	clear(r.sessions)

	if r.targets != nil {
		r.targets.events.stop(err)
	}
}

func (r *router) parentClient(s *targetSession) *cdp.Client {
	if s.parent == "" {
		return r.targets.manager.browserClient
	}

	return r.targets.parentClient(s.parent)
}

// Early Page events stay inside the child's bounded queue until its listeners exist.
func (r *router) frameEvent(parent target.SessionID, msg *rpcc.Response) error {
	switch msg.Method {
	case "Page.frameDetached":
		var ev page.FrameDetachedReply
		if err := json.Unmarshal(msg.Args, &ev); err != nil {
			return err
		}

		r.targets.events.put(targetChange{source: parent, frameID: target.ID(ev.FrameID), frameReason: ev.Reason})
	case "Page.frameNavigated":
		var ev page.FrameNavigatedReply
		if err := json.Unmarshal(msg.Args, &ev); err != nil {
			return err
		}

		r.targets.events.put(targetChange{source: parent, frameID: target.ID(ev.Frame.ID), frameReason: "navigated"})
	}

	return nil
}
