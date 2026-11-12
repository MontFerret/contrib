package session

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"testing"
	"time"

	"github.com/mafredri/cdp/protocol/target"
	"github.com/mafredri/cdp/rpcc"
)

type protocolRequest struct {
	ID        uint64
	Method    string
	SessionID target.SessionID
	Params    struct{ TargetID target.ID }
}

type protocolPeer struct {
	t       *testing.T
	encoder *json.Encoder
	decoder *json.Decoder
	remote  net.Conn
	manager *Manager
}

func newProtocolPeer(t *testing.T) *protocolPeer {
	t.Helper()
	local, remote := net.Pipe()
	_ = remote.SetDeadline(time.Now().Add(8 * time.Second))
	t.Cleanup(func() { _ = remote.Close() })
	p := &protocolPeer{t: t, encoder: json.NewEncoder(remote), decoder: json.NewDecoder(remote), remote: remote}

	conn, err := DialFlat(t.Context(), "", rpcc.WithDialer(func(context.Context, string) (io.ReadWriteCloser, error) { return local, nil }))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = conn.Conn.Close() })
	ready := make(chan error, 1)
	go func() {
		var err error
		p.manager, err = NewFlat(t.Context(), conn, "page")
		ready <- err
	}()
	p.reply(p.request("Target.setDiscoverTargets"), struct{}{})
	attach := p.request("Target.autoAttachRelated")
	p.attach("root", "page", "page")
	p.reply(attach, struct{}{})

	if err := <-ready; err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = p.manager.Close() })

	return p
}

func (p *protocolPeer) send(msg any) {
	p.t.Helper()

	if err := p.encoder.Encode(msg); err != nil {
		p.t.Fatal(err)
	}
}

func (p *protocolPeer) request(method string) protocolRequest {
	p.t.Helper()
	var r protocolRequest

	if err := p.decoder.Decode(&r); err != nil {
		p.t.Fatal(err)
	}

	if method != "" && method != r.Method {
		p.t.Fatalf("got %s, want %s", r.Method, method)
	}

	return r
}

func (p *protocolPeer) reply(r protocolRequest, result any) {
	p.t.Helper()
	p.send(map[string]any{"id": r.ID, "sessionId": r.SessionID, "result": result})
}

func (p *protocolPeer) event(session, method string, params any) {
	p.t.Helper()
	p.send(map[string]any{"sessionId": session, "method": method, "params": params})
}

func (p *protocolPeer) attach(id, name, kind string) {
	p.t.Helper()
	p.event("", "Target.attachedToTarget", target.AttachedToTargetReply{SessionID: target.SessionID(id), TargetInfo: target.Info{TargetID: target.ID(name), Type: kind}})
}

// NewFrameSwapPeer lets external tests exercise retained DOM handles through the same protocol fixture.
func NewFrameSwapPeer(t *testing.T) (*Manager, net.Conn) {
	p := newProtocolPeer(t)

	return p.manager, p.remote
}
