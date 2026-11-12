package session_test

import (
	"encoding/json"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/rpcc"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/session"
)

type frameProtocol struct {
	*session.Manager
	swapping    *atomic.Bool
	missingTree *atomic.Bool
	held        <-chan uint64
	checked     <-chan struct{}
	send        func(map[string]any)
}

func newFrameProtocol(t *testing.T, sibling bool) *frameProtocol {
	t.Helper()
	manager, conn := session.NewFrameSwapPeer(t)
	encoder := json.NewEncoder(conn)
	var writeMu sync.Mutex
	send := func(packet map[string]any) {
		writeMu.Lock()
		defer writeMu.Unlock()
		if err := encoder.Encode(packet); err != nil && t.Context().Err() == nil {
			t.Error(err)
		}
	}

	rootID := page.FrameID("page")
	child := page.FrameTree{Frame: page.Frame{ID: "frame", ParentID: &rootID}}
	swapping := new(atomic.Bool)
	missingTree := new(atomic.Bool)
	held := make(chan uint64, 1)
	checked := make(chan struct{}, 1)
	go func() {
		decoder := json.NewDecoder(conn)
		for {
			var r struct {
				ID                uint64
				SessionID, Method string
				Params            struct{ FunctionDeclaration string }
			}

			if err := decoder.Decode(&r); err != nil {
				return
			}

			reply := map[string]any{"id": r.ID, "sessionId": r.SessionID, "result": struct{}{}}
			switch r.Method {
			case "Page.enable":
				held <- r.ID

				continue
			case "Page.getFrameTree":
				tree := page.FrameTree{Frame: page.Frame{ID: rootID}}
				if r.SessionID == "replacement" {
					if missingTree.Load() {
						delete(reply, "result")
						reply["error"] = rpcc.ResponseError{Code: -32000, Message: "renderer unavailable"}
						break
					}

					tree = child
				} else if !swapping.Load() {
					tree.ChildFrames = []page.FrameTree{child}
				} else {
					select {
					case checked <- struct{}{}:
					default:
					}
				}

				if sibling && r.SessionID == "root" {
					tree.ChildFrames = append(tree.ChildFrames, page.FrameTree{Frame: page.Frame{ID: "sibling", ParentID: &rootID}})
				}

				reply["result"] = page.GetFrameTreeReply{FrameTree: tree}
			case "Page.createIsolatedWorld":
				reply["result"] = map[string]int{"executionContextId": 1}
			case "Runtime.callFunctionOn":
				object := map[string]any{"type": "object", "objectId": "document"}
				if strings.Contains(r.Params.FunctionDeclaration, "document.URL") {
					object = map[string]any{"type": "string", "value": r.SessionID}
				}

				if strings.Contains(r.Params.FunctionDeclaration, "document.readyState") {
					object = map[string]any{"type": "boolean", "value": true}
				}

				reply["result"] = map[string]any{"result": object}
			case "DOM.getFrameOwner":
				delete(reply, "result")
				reply["error"] = rpcc.ResponseError{Code: -32000, Message: "Frame with the given id was not found"}
			}

			send(reply)
		}
	}()

	return &frameProtocol{
		Manager: manager, swapping: swapping, missingTree: missingTree,
		held: held, checked: checked, send: send,
	}
}
