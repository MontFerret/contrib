package session

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"testing"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/rpcc"
)

// Both transports run the same typed call over an in-process CDP peer.
func BenchmarkTargetRoundTrip(b *testing.B) {
	for _, mode := range []string{"legacy", "flat"} {
		b.Run(mode, func(b *testing.B) { benchmarkRoundTrip(b, mode) })
	}
}

func benchmarkRoundTrip(b *testing.B, mode string) {
	local, remote := net.Pipe()
	defer remote.Close()

	dial := rpcc.WithDialer(func(context.Context, string) (io.ReadWriteCloser, error) { return local, nil })
	go serveBenchmarkTarget(remote)

	var manager *Manager
	var err error

	if mode == "flat" {
		conn, dialErr := DialFlat(b.Context(), "", dial)
		if dialErr != nil {
			b.Fatal(dialErr)
		}

		manager, err = NewFlat(b.Context(), conn, "page")
	} else {
		conn, dialErr := rpcc.DialContext(b.Context(), "", dial)
		if dialErr != nil {
			b.Fatal(dialErr)
		}

		manager, err = New(b.Context(), conn, cdp.NewClient(conn), "page")
	}

	if err != nil {
		b.Fatal(err)
	}

	defer manager.Close()
	root := manager.Root().CDP
	b.ReportAllocs()
	for b.Loop() {
		if err := root.Runtime.Enable(b.Context()); err != nil {
			b.Fatal(err)
		}
	}
}

func serveBenchmarkTarget(peer net.Conn) {
	decoder, encoder := json.NewDecoder(peer), json.NewEncoder(peer)
	for {
		var request struct {
			ID        uint64
			Method    string
			SessionID string
			Params    struct{ Message string }
		}

		if err := decoder.Decode(&request); err != nil {
			return
		}

		reply := map[string]any{"id": request.ID, "result": struct{}{}}
		switch request.Method {
		case "Target.attachToTarget":
			reply["result"] = map[string]string{"sessionId": "root"}
		case "Target.autoAttachRelated":
			_ = encoder.Encode(map[string]any{"method": "Target.attachedToTarget", "params": map[string]any{"sessionId": "root", "targetInfo": map[string]string{"targetId": "page", "type": "page"}}})
		case "Target.sendMessageToTarget":
			var inner struct{ ID uint64 }
			_ = json.Unmarshal([]byte(request.Params.Message), &inner)
			result, _ := json.Marshal(map[string]any{"id": inner.ID, "result": struct{}{}})
			_ = encoder.Encode(map[string]any{"method": "Target.receivedMessageFromTarget", "params": map[string]string{"sessionId": "root", "message": string(result)}})
		default:
			if request.SessionID != "" {
				reply["sessionId"] = request.SessionID
			}
		}

		if err := encoder.Encode(reply); err != nil {
			return
		}
	}
}
