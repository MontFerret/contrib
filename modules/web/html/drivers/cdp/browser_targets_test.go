package cdp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/page"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"
	"github.com/mafredri/cdp/protocol/target"
	"github.com/mafredri/cdp/rpcc"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	cdpnet "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/network"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/session"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestBrowserFramesAndWorkers(t *testing.T) {
	ctx, address, _ := policyBrowser(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/worker.js":
			w.Header().Set("Content-Type", "application/javascript")
			fmt.Fprint(w, "self.onmessage=e=>postMessage(e.data)")
		case "/sw.js":
			w.Header().Set("Content-Type", "application/javascript")
			fmt.Fprint(w, "self.addEventListener('install',()=>self.skipWaiting()); self.addEventListener('activate',e=>e.waitUntil(self.clients.claim()));")
		default:
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, "<title>%s</title><body>fixture</body>", r.URL.Path)
		}
	}))
	defer server.Close()
	p := policyPage(t, ctx, New(WithAddress(address)), server.URL)
	remote := strings.Replace(server.URL, "127.0.0.1", "localhost", 1)
	policyEvaluate(t, ctx, p, fmt.Sprintf(`for(const src of [%q,%q]){const f=document.createElement('iframe');f.src=src;document.body.append(f)};true`, server.URL+"/local", remote+"/remote"))
	policyWait(t, func() bool {
		frames, _ := p.sessions.FrameSessions()

		return len(frames) == 1
	})
	assertContents := func(expected string) drivers.HTMLDocument {
		var found drivers.HTMLDocument
		t.Helper()
		policyWait(t, func() bool {
			frames, err := p.GetFrames(ctx)
			if err != nil {
				t.Fatal(err)
			}

			n, err := frames.Length(ctx)
			if err != nil {
				t.Fatal(err)
			}

			for i := runtime.Int(0); i < n; i++ {
				item, err := frames.At(ctx, i)
				if err != nil {
					t.Fatal(err)
				}

				doc := item.(drivers.HTMLDocument)
				if strings.HasSuffix(doc.GetURL().String(), expected) {
					found = doc

					return doc.GetTitle().String() == expected
				}
			}

			return false
		})

		return found
	}

	assertContents("/local")
	retained := assertContents("/remote")
	policyEvaluate(t, ctx, p, fmt.Sprintf("document.querySelectorAll('iframe')[1].src=%q", server.URL+"/swapped"))
	policyWait(t, func() bool { frames, _ := p.sessions.FrameSessions(); return len(frames) == 0 })
	policyWait(t, func() bool { return retained.GetTitle().String() == "/swapped" })
	assertContents("/swapped")
	policyEvaluate(t, ctx, p, fmt.Sprintf("document.querySelectorAll('iframe')[1].src=%q", remote+"/again"))
	assertContents("/again")
	policyEvaluate(t, ctx, p, `window.worker = new Worker('/worker.js'); navigator.serviceWorker.register('/sw.js').then(()=>navigator.serviceWorker.ready).then(()=>true)`)
	for _, kind := range []string{"worker", "service_worker"} {
		var worker *session.Client
		policyWait(t, func() bool {
			for _, c := range p.sessions.Snapshot() {
				if c.Info.Type == kind {
					worker = c

					return true
				}
			}

			return false
		})

		reply, err := worker.CDP.Runtime.Evaluate(ctx, cdpruntime.NewEvaluateArgs("self.location.pathname"))
		if err != nil || !strings.Contains(string(reply.Result.Value), ".js") {
			t.Fatalf("%s evaluate: %v, %v", kind, reply, err)
		}
	}
}

func TestBrowserPopupOwnership(t *testing.T) {
	ctx, address, browser := policyBrowser(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "<title>popup fixture</title>") }))
	defer srv.Close()
	drv := New(WithAddress(address), WithKeepCookies())
	p := policyPage(t, ctx, drv, srv.URL)
	sibling := policyPage(t, ctx, drv, srv.URL)
	rootID := p.sessions.Root().TargetID

	policyEvaluate(t, ctx, p, `window.a=window.open('/first');true`)
	policyEvaluate(t, ctx, p, `window.b=window.open('/second');true`)
	policyEvaluate(t, ctx, p, `window.open('/brief').close();true`)
	policyEvaluate(t, ctx, sibling, `window.other=window.open('/other');true`)
	policyWait(t, func() bool { return len(p.sessions.Snapshot()) == 3 && len(sibling.sessions.Snapshot()) == 2 })
	for _, c := range p.sessions.Snapshot() {
		if c.BelongsToPopup() {
			reply, err := c.CDP.Runtime.Evaluate(ctx, cdpruntime.NewEvaluateArgs("!!window.open('/nested')").SetUserGesture(true).SetReturnByValue(true))
			if err != nil || string(reply.Result.Value) != "true" {
				t.Fatalf("open nested popup: %v, %v", reply, err)
			}

			break
		}
	}

	policyWait(t, func() bool { return len(p.sessions.Snapshot()) == 4 })

	nav, err := p.network.OnNavigation(ctx)
	if err != nil {
		t.Fatal(err)
	}

	defer nav.Close()
	events := nav.Read(ctx)
	policyEvaluate(t, ctx, p, `window.a.location.hash='popup';true`)
	policyEvaluate(t, ctx, p, `location.hash='opener';true`)
	select {
	case msg := <-events:
		if msg.Err() != nil || !strings.HasSuffix(msg.Value().(*cdpnet.NavigationEvent).URL, "#opener") {
			t.Fatalf("opener navigation = %v", msg)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	frames, err := p.GetFrames(ctx)
	if err != nil {
		t.Fatal(err)
	}

	count, err := frames.Length(ctx)
	if err != nil || count != 1 {
		t.Fatalf("opener frame count = %v (%v)", count, err)
	}

	owned := map[target.ID]bool{}
	for _, c := range p.sessions.Snapshot() {
		if c.BelongsToPopup() {
			owned[c.TargetID] = true
		}
	}

	var streams []runtime.Stream
	for _, event := range []string{drivers.RequestEvent, drivers.ResponseEvent, drivers.NetworkRequestStartedEvent, drivers.NetworkIdleEvent} {
		s, err := p.Subscribe(ctx, runtime.Subscription{EventName: runtime.NewString(event)})
		if err != nil {
			t.Fatal(err)
		}

		defer s.Close()
		streams = append(streams, s)
	}

	wait, err := p.navigationStream(ctx, cdpnet.WaitEventOptions{})
	if err != nil {
		t.Fatal(err)
	}

	defer wait.Close()
	waiting := make(chan error, 1)
	go func() { waiting <- p.readNavigation(ctx, wait) }()

	if _, err := browser.Target.CloseTarget(ctx, target.NewCloseTargetArgs(rootID)); err != nil {
		t.Fatal(err)
	}

	policyWait(t, func() bool { return p.sessions.Root() == nil })
	select {
	case err := <-waiting:
		if !errors.Is(err, rpcc.ErrConnClosing) {
			t.Fatalf("pending wait: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("navigation wait survived its root")
	}

	if err := p.WaitForNavigation(ctx, "never-matches"); !errors.Is(err, rpcc.ErrConnClosing) {
		t.Fatalf("new wait after root loss: %v", err)
	}

	for _, s := range streams {
		messages := s.Read(ctx)
	read:
		for {
			select {
			case msg, ok := <-messages:
				if !ok {
					t.Fatal("stream ended without a root-closed error")
				}

				if errors.Is(msg.Err(), rpcc.ErrConnClosing) {
					break read
				}
			case <-time.After(time.Second):
				t.Fatal("network subscription survived root loss")
			}
		}
	}

	for _, c := range p.sessions.Snapshot() {
		reply, err := c.CDP.Runtime.Evaluate(ctx, cdpruntime.NewEvaluateArgs("document.title"))
		if err != nil || string(reply.Result.Value) != `"popup fixture"` {
			t.Fatalf("owned popup after root loss: %v, %v", reply, err)
		}
	}

	current := func() map[target.ID]bool {
		reply, err := browser.Target.GetTargets(ctx, target.NewGetTargetsArgs())
		if err != nil {
			t.Fatal(err)
		}

		ids := map[target.ID]bool{}
		for _, info := range reply.TargetInfos {
			ids[info.TargetID] = true
		}

		return ids
	}

	live := current()
	for id := range owned {
		if !live[id] {
			t.Fatal("opener destruction closed popup", id)
		}
	}

	if err := p.Close(); err != nil {
		t.Fatal(err)
	}

	live = current()
	for id := range owned {
		if live[id] {
			t.Fatal("page close leaked popup", id)
		}
	}

	survivors := sibling.sessions.Snapshot()
	if len(survivors) != 2 {
		t.Fatalf("sibling retained %d sessions, want its page and popup", len(survivors))
	}

	for _, c := range survivors {
		if !live[c.TargetID] {
			t.Fatal("page close removed sibling target", c.TargetID)
		}

		reply, err := c.CDP.Runtime.Evaluate(ctx, cdpruntime.NewEvaluateArgs("document.title"))
		if err != nil || string(reply.Result.Value) != `"popup fixture"` {
			t.Fatalf("sibling target %s is unusable: %v, %v", c.TargetID, reply, err)
		}
	}
}

func TestBrowserConnectionLoss(t *testing.T) {
	ctx, address, browser := policyBrowser(t)
	p := policyPage(t, ctx, New(WithAddress(address), WithKeepCookies()), "about:blank")
	targetID := p.sessions.Root().TargetID

	navigation, err := p.network.OnNavigation(ctx)
	if err != nil {
		t.Fatal(err)
	}

	defer navigation.Close()
	messages := navigation.Read(ctx)

	requests, err := p.network.OnEvent(ctx, runtime.NewString(drivers.NetworkRequestStartedEvent), nil)
	if err != nil {
		t.Fatal(err)
	}

	defer requests.Close()
	networkEvents := requests.Read(ctx)

	done := make(chan error, 1)
	go func() {
		_, err := p.client.Runtime.Evaluate(ctx, cdpruntime.NewEvaluateArgs("new Promise(()=>{})").SetAwaitPromise(true))
		done <- err
	}()

	if err := p.policySession.conn.Close(); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("pending evaluation succeeded")
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	for _, stream := range []<-chan runtime.Message{messages, networkEvents} {
	read:
		for {
			select {
			case _, ok := <-stream:
				if !ok {
					break read
				}
			case <-ctx.Done():
				t.Fatal("stream survived connection loss")
			}
		}
	}

	if err := p.Close(); err != nil {
		t.Fatal(err)
	}

	targets, err := browser.Target.GetTargets(ctx, target.NewGetTargetsArgs())
	if err != nil {
		t.Fatal(err)
	}

	for _, info := range targets.TargetInfos {
		if info.TargetID == targetID {
			t.Fatal("disconnected page survived cleanup")
		}
	}
}

func newExistingPageManager(t *testing.T, ctx context.Context, address string, browser *cdp.Client) *session.Manager {
	t.Helper()

	root, err := browser.Target.CreateTarget(ctx, target.NewCreateTargetArgs("about:blank"))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _, _ = browser.Target.CloseTarget(ctx, target.NewCloseTargetArgs(root.TargetID)) })

	version, err := devtool.New(address).Version(ctx)
	if err != nil {
		t.Fatal(err)
	}

	conn, err := session.DialFlat(ctx, version.WebSocketDebuggerURL)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = conn.Conn.Close() })

	manager, err := session.NewFlat(ctx, conn, root.TargetID)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = manager.Close() })

	return manager
}

func TestExistingPageCloseCancelsNavigationWait(t *testing.T) {
	ctx, address, browser := policyBrowser(t)
	manager := newExistingPageManager(t, ctx, address, browser)

	p, err := LoadHTMLPage(ctx, manager, drivers.Params{})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = p.Close() })

	wait, err := p.navigationStream(ctx, cdpnet.WaitEventOptions{})
	if err != nil {
		t.Fatal(err)
	}

	defer wait.Close()
	waiting := make(chan error, 1)
	go func() { waiting <- p.readNavigation(ctx, wait) }()

	if err := p.Close(); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-waiting:
		if err == nil {
			t.Fatal("page close completed a navigation wait successfully")
		}
	case <-ctx.Done():
		t.Fatal("navigation wait survived page close")
	}

	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
}

type rejectingContentPage struct {
	cdp.Page
	err error
}

func (p *rejectingContentPage) SetDocumentContent(context.Context, *page.SetDocumentContentArgs) error {
	return p.err
}

func TestExistingPageContentFailureClosesTarget(t *testing.T) {
	ctx, address, browser := policyBrowser(t)
	manager := newExistingPageManager(t, ctx, address, browser)
	root := manager.Root()
	cause := errors.New("document content rejected")
	root.CDP.Page = &rejectingContentPage{Page: root.CDP.Page, err: cause}

	_, err := LoadHTMLPageWithContent(ctx, manager, drivers.Params{}, []byte("<html></html>"))
	if !errors.Is(err, cause) {
		t.Fatalf("content load returned %v", err)
	}

	if manager.Context().Err() == nil {
		t.Fatal("content failure left the manager running")
	}

	targets, err := browser.Target.GetTargets(ctx, target.NewGetTargetsArgs())
	if err != nil {
		t.Fatal(err)
	}

	for _, info := range targets.TargetInfos {
		if info.TargetID == root.TargetID {
			t.Fatal("content failure left the target open")
		}
	}
}

func TestBrowserFrameInputAndOrder(t *testing.T) {
	ctx, address, _ := policyBrowser(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<title>%s</title><body>", r.URL.Path)
		switch r.URL.Path {
		case "/":
			remote := "http://" + strings.Replace(r.Host, "127.0.0.1", "localhost", 1)
			fmt.Fprintf(w, `<button onmouseover="document.title='wrong'" onclick="document.title='wrong'">root</button>
				<iframe style="position:absolute;left:350px;top:170px" src="%s/first"></iframe>
				<iframe src="/local"></iframe><iframe src="%s/last"></iframe>`, remote, remote)
		case "/first":
			fmt.Fprint(w, `<button id="hit" onmouseover="document.title='hover:'+this.id" onclick="document.title='click:'+this.id">frame button</button>
				<iframe src="/nested"></iframe>`)
		}
	}))
	defer server.Close()

	for range 3 {
		p := policyPage(t, ctx, New(WithAddress(address)), server.URL)
		policyWait(t, func() bool {
			frames, _ := p.sessions.FrameSessions()

			return len(frames) == 2
		})

		list, err := p.GetFrames(ctx)
		if err != nil {
			t.Fatal(err)
		}

		count, err := list.Length(ctx)
		if err != nil || count != 5 {
			t.Fatalf("frame count %d: %v", count, err)
		}

		var remote drivers.HTMLDocument
		for i, want := range []string{"/", "/first", "/nested", "/local", "/last"} {
			item, err := list.At(ctx, runtime.Int(i))
			if err != nil {
				t.Fatal(err)
			}

			doc := item.(drivers.HTMLDocument)
			if title := doc.GetTitle().String(); title != want {
				t.Fatalf("frame[%d] = %s, want %s", i, title, want)
			}

			if i == 1 {
				remote = doc
			}
		}

		input, err := drivers.ToInteractionTarget(remote.GetElement())
		if err != nil {
			t.Fatal(err)
		}

		button := drivers.NewCSSSelector("#hit")
		if err := input.HoverBySelector(ctx, button); err != nil {
			t.Fatal(err)
		}

		if got := remote.GetTitle(); got != "hover:hit" {
			t.Fatalf("hover recipient: %s", got)
		}

		if err := input.ClickBySelector(ctx, button, 1); err != nil {
			t.Fatal(err)
		}

		if got := remote.GetTitle(); got != "click:hit" {
			t.Fatalf("click recipient: %s", got)
		}

		if err := p.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
