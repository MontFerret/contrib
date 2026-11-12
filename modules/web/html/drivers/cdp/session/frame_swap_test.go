package session_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/protocol/target"
	"github.com/mafredri/cdp/rpcc"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	driver "github.com/MontFerret/contrib/modules/web/html/drivers/cdp"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/dom"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/network"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestRetainedFrameDuringReplacement(t *testing.T) {
	for _, fail := range []bool{false, true} {
		name := "recover and remove"
		if fail {
			name = "setup failure"
		}

		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			peer := newFrameProtocol(t, fail)
			manager := peer.Manager
			send := peer.send

			d, err := dom.New(zerolog.Nop(), manager.Root().CDP, nil, nil)
			if err != nil {
				t.Fatal(err)
			}

			d.SetFrameSessions(manager)

			root, err := d.LoadRootDocument(ctx)
			if err != nil {
				t.Fatal(err)
			}

			d.SetMainFrame(root)

			doc, err := d.GetFrameNode(ctx, "frame")
			if err != nil {
				t.Fatal(err)
			}

			if url, err := doc.GetCurrentURL(ctx); err != nil || url != "root" {
				t.Fatalf("original: %s %v", url, err)
			}

			netManager, err := network.New(zerolog.Nop(), manager.Root().CDP, manager, network.Options{})
			if err != nil {
				t.Fatal(err)
			}

			defer netManager.Close()
			p := driver.NewHTMLPage(zerolog.Nop(), manager.Root().CDP, manager, netManager, d)
			subscribe := func(frame drivers.HTMLDocument) runtime.Stream {
				opts := runtime.NewObject()
				if frame != nil {
					opts = runtime.NewObjectWith(map[string]runtime.Value{"frame": frame})
				}

				stream, err := p.Subscribe(ctx, runtime.Subscription{EventName: drivers.NavigationEvent, Options: opts})
				if err != nil {
					t.Fatal(err)
				}

				t.Cleanup(func() { _ = stream.Close() })

				return stream
			}

			var pending, sibling, unfiltered runtime.Stream
			if fail {
				other, err := d.GetFrameNode(ctx, "sibling")
				if err != nil {
					t.Fatal(err)
				}

				pending, sibling, unfiltered = subscribe(doc), subscribe(other), subscribe(nil)
			}

			peer.swapping.Store(true)
			send(map[string]any{"sessionId": "root", "method": "Page.frameDetached", "params": page.FrameDetachedReply{FrameID: "frame", Reason: "swap"}})
			send(map[string]any{"sessionId": "root", "method": "Target.attachedToTarget", "params": target.AttachedToTargetReply{SessionID: "replacement", TargetInfo: target.Info{TargetID: "frame", Type: "iframe"}}})
			var id uint64
			select {
			case id = <-peer.held:
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}

			done := make(chan error, 1)
			go func() {
				url, err := doc.GetCurrentURL(ctx)
				if err == nil && url != "replacement" {
					err = errors.New("read used the old renderer")
				}

				done <- err
			}()
			select {
			case <-peer.checked:
			case err := <-done:
				t.Fatalf("read returned during setup: %v", err)
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}

			short, stop := context.WithTimeout(ctx, 20*time.Millisecond)
			_, err = doc.GetCurrentURL(short)
			stop()
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("concurrent read exceeded its deadline: %v", err)
			}

			// A root command can finish while the retained read is waiting.
			if err := manager.Root().CDP.Runtime.Enable(ctx); err != nil {
				t.Fatal(err)
			}

			waiting := make(chan error, 1)
			if fail {
				go func() { waiting <- p.WaitForFrameNavigation(ctx, doc, "") }()
			}

			reply := map[string]any{"id": id, "sessionId": "replacement"}
			if fail {
				reply["error"] = rpcc.ResponseError{Code: -32000, Message: "frame setup rejected"}
			} else {
				reply["result"] = struct{}{}
			}

			send(reply)
			select {
			case err := <-done:
				if fail {
					cause := manager.FrameError("frame")
					var protocol *rpcc.ResponseError
					if !errors.As(cause, &protocol) || protocol.Message != "frame setup rejected" || !errors.Is(err, cause) {
						t.Fatalf("frame failure: %v, read: %v", cause, err)
					}

					if err := <-waiting; !errors.Is(err, cause) {
						t.Fatalf("pending navigation wait: %v", err)
					}

					if err := p.WaitForFrameNavigation(ctx, doc, ""); !errors.Is(err, cause) {
						t.Fatalf("wait after failure: %v", err)
					}

					if msg := <-pending.Read(ctx); !errors.Is(msg.Err(), cause) {
						t.Fatalf("pending subscription: %v", msg.Err())
					}

					send(map[string]any{"sessionId": "root", "method": "Page.navigatedWithinDocument", "params": page.NavigatedWithinDocumentReply{FrameID: "sibling", URL: "about:blank#next"}})
					for _, stream := range []runtime.Stream{sibling, unfiltered} {
						select {
						case msg, ok := <-stream.Read(ctx):
							if !ok || msg.Err() != nil || msg.Value().(*network.NavigationEvent).FrameID != "sibling" {
								t.Fatalf("healthy navigation: %v", msg)
							}
						case <-ctx.Done():
							t.Fatal(ctx.Err())
						}
					}

					return
				}

				if err != nil {
					t.Fatal(err)
				}
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}

			peer.missingTree.Store(true)
			if err := d.ReloadRoot(ctx); err != nil {
				t.Fatal(err)
			}

			short, stop = context.WithTimeout(ctx, 40*time.Millisecond)
			_, err = doc.GetCurrentURL(short)
			stop()

			if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("missing renderer reply retired the document: %v", err)
			}

			peer.missingTree.Store(false)

			if url, err := doc.GetCurrentURL(ctx); err != nil || url != "replacement" {
				t.Fatalf("retained document after renderer recovery: %s, %v", url, err)
			}

			send(map[string]any{"sessionId": "root", "method": "Page.frameDetached", "params": page.FrameDetachedReply{FrameID: "frame", Reason: "remove"}})
			send(map[string]any{"sessionId": "root", "method": "Target.detachedFromTarget", "params": target.DetachedFromTargetReply{SessionID: "replacement"}})
			if err := d.RefreshFrames(ctx); err != nil {
				t.Fatal(err)
			}

			if _, err := doc.GetCurrentURL(ctx); !errors.Is(err, drivers.ErrDetached) {
				t.Fatalf("removed document: %v", err)
			}
		})
	}
}
