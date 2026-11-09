package cdp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gorilla/websocket"
	cdplib "github.com/mafredri/cdp"
	"github.com/mafredri/cdp/rpcc"
)

// policyTestEndpoint supplies CDP replies without a browser. Returning nil leaves
// a command pending, while the reader remains available to observe cancellation.
type policyTestEndpoint struct {
	server *httptest.Server
	handle func(int32, string, json.RawMessage) (any, error)
	opened atomic.Int32
	active atomic.Int32
	closed chan int32
}

func newPolicyTestEndpoint(t *testing.T, handle func(int32, string, json.RawMessage) (any, error)) *policyTestEndpoint {
	t.Helper()
	f := &policyTestEndpoint{handle: handle, closed: make(chan int32, 32)}
	f.server = httptest.NewServer(http.HandlerFunc(f.serveHTTP))
	t.Cleanup(f.server.Close)

	return f
}

func (f *policyTestEndpoint) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/json/version" {
		_ = json.NewEncoder(w).Encode(map[string]string{"webSocketDebuggerUrl": strings.Replace(f.server.URL, "http:", "ws:", 1)})

		return
	}

	id := f.opened.Add(1)
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	f.active.Add(1)
	defer func() {
		_ = conn.Close()
		f.active.Add(-1)
		f.closed <- id
	}()

	for {
		var request struct {
			ID     uint64          `json:"id"`
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}

		if err := conn.ReadJSON(&request); err != nil {
			return
		}

		result, callErr := f.handle(id, request.Method, request.Params)
		if result == nil && callErr == nil {
			continue
		}

		response := map[string]any{"id": request.ID, "result": result}
		if callErr != nil {
			delete(response, "result")
			response["error"] = callErr
		}

		if err := conn.WriteJSON(response); err != nil {
			return
		}
	}
}

func (f *policyTestEndpoint) client(t *testing.T) (*rpcc.Conn, *cdplib.Client) {
	t.Helper()
	conn, err := rpcc.DialContext(t.Context(), strings.Replace(f.server.URL, "http:", "ws:", 1))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = conn.Close() })

	return conn, cdplib.NewClient(conn)
}
