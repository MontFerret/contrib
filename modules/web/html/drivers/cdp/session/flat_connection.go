package session

import (
	"context"
	"encoding/json"
	"io"

	"github.com/mafredri/cdp/protocol/target"
	"github.com/mafredri/cdp/rpcc"
)

// FlatConnection owns a browser WebSocket and routes each flat target session
// to a separate typed client. Browser commands use Conn without a session ID.
type FlatConnection struct {
	Conn   *rpcc.Conn
	router *router
}

// DialFlat installs the session-aware codec at dial time. The caller owns Conn
// until NewFlat succeeds, and can issue browser commands before attaching.
func DialFlat(ctx context.Context, url string, options ...rpcc.DialOption) (*FlatConnection, error) {
	r := &router{sessions: make(map[target.SessionID]*targetSession), pending: make(map[uint64]pendingCall)}
	options = append(options, rpcc.WithCodec(func(rw io.ReadWriter) rpcc.Codec {
		r.decoder, r.encoder = json.NewDecoder(rw), json.NewEncoder(rw)

		return r
	}))

	conn, err := rpcc.DialContext(ctx, url, options...)
	if err != nil {
		return nil, err
	}

	return &FlatConnection{Conn: conn, router: r}, nil
}
