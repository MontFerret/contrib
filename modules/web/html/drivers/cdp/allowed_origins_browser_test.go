package cdp

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/fetch"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"
	"github.com/mafredri/cdp/protocol/security"
	"github.com/mafredri/cdp/protocol/target"
	"github.com/mafredri/cdp/rpcc"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func policyBrowser(t *testing.T) (context.Context, string, *cdp.Client) {
	t.Helper()
	address := os.Getenv("FERRET_CDP_TEST_ADDRESS")
	if address == "" {
		t.Skip("set FERRET_CDP_TEST_ADDRESS to a dedicated local Chromium endpoint")
	}

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	t.Cleanup(cancel)
	version, err := devtool.New(address).Version(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("browser: %s", version.Browser)
	conn, err := rpcc.DialContext(ctx, version.WebSocketDebuggerURL)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = conn.Close() })
	client := cdp.NewClient(conn)
	if err := client.Security.SetIgnoreCertificateErrors(ctx, security.NewSetIgnoreCertificateErrorsArgs(true)); err != nil {
		t.Fatal(err)
	}

	return ctx, address, client
}

func policyPage(t *testing.T, ctx context.Context, driver *Driver, url string) *HTMLPage {
	t.Helper()
	result, err := driver.Open(ctx, drivers.Params{URL: url})
	if err != nil {
		t.Fatal(err)
	}

	page := result.(*HTMLPage)
	t.Cleanup(func() { _ = page.Close() })

	return page
}

func policyEvaluate(t *testing.T, ctx context.Context, p *HTMLPage, expression string) string {
	t.Helper()
	reply, err := p.client.Runtime.Evaluate(ctx, cdpruntime.NewEvaluateArgs(expression).
		SetAwaitPromise(true).SetReturnByValue(true).SetUserGesture(true))
	if err != nil {
		t.Fatal(err)
	}

	if reply.ExceptionDetails != nil {
		t.Fatalf("JavaScript exception: %+v", reply.ExceptionDetails)
	}

	return string(reply.Result.Value)
}

func policyWait(t *testing.T, check func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for !check() {
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for policy lifecycle")
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func TestBrowserAllowedOriginsRequests(t *testing.T) {
	ctx, address, _ := policyBrowser(t)
	var forbidden atomic.Int32
	denied := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forbidden.Add(1)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "X-Probe")
		_, _ = w.Write([]byte("forbidden"))
	}))
	defer denied.Close()

	allowed := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, denied.URL+"/redirect-target", http.StatusTemporaryRedirect)

			return
		}

		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><title>allowed</title><body>fixture</body></html>"))
	}))
	defer allowed.Close()

	driver := New(WithAddress(address), WithAllowedOrigins(allowed.URL))
	defer driver.Close()
	p := policyPage(t, ctx, driver, allowed.URL)

	for _, path := range []string{"/fetch", "/preflight"} {
		expression := fmt.Sprintf(`fetch(%q, {headers: {'X-Probe':'yes'}}).then(()=>'allowed', ()=>'blocked')`, denied.URL+path)
		if actual := policyEvaluate(t, ctx, p, expression); actual != `"blocked"` {
			t.Fatalf("fetch result = %s", actual)
		}
	}

	if actual := policyEvaluate(t, ctx, p, fmt.Sprintf(`new Promise(resolve=>{const x=new XMLHttpRequest(); x.open('GET',%q); x.onload=()=>resolve('allowed'); x.onerror=()=>resolve('blocked'); x.send();})`, denied.URL+"/xhr")); actual != `"blocked"` {
		t.Fatalf("XHR result = %s", actual)
	}

	for _, kind := range []string{"script", "img", "link", "iframe"} {
		expression := fmt.Sprintf(`new Promise(resolve=>{const e=document.createElement(%q); if(e.tagName==='LINK'){e.rel='stylesheet';e.href=%q;}else{e.src=%q;}e.onload=()=>resolve('loaded');e.onerror=()=>resolve('blocked');document.body.appendChild(e);setTimeout(()=>resolve('settled'),250);})`, kind, denied.URL+"/"+kind, denied.URL+"/"+kind)
		policyEvaluate(t, ctx, p, expression)
	}

	if _, err := driver.Open(ctx, drivers.Params{URL: denied.URL + "/navigation"}); err == nil {
		t.Fatal("unlisted HTTPS navigation succeeded")
	}

	if _, err := driver.Open(ctx, drivers.Params{URL: allowed.URL + "/redirect"}); err == nil {
		t.Fatal("redirect to forbidden origin succeeded")
	}

	if actual := policyEvaluate(t, ctx, p, fmt.Sprintf(`fetch(%q).then(()=>'allowed',()=>'blocked')`, allowed.URL+"/redirect")); actual != `"blocked"` {
		t.Fatalf("redirected fetch result = %s", actual)
	}

	// Chromium 155 may stall a redirected request requiring a CORS preflight
	// before reporting the destination through Fetch. Bound that path with an
	// abort signal; either browser rejection or abort must leave the server quiet.
	if actual := policyEvaluate(t, ctx, p, fmt.Sprintf(`fetch(%q,{headers:{'X-Probe':'yes'},signal:AbortSignal.timeout(250)}).then(()=>'allowed',()=> 'blocked-or-aborted')`, allowed.URL+"/redirect")); actual != `"blocked-or-aborted"` {
		t.Fatalf("redirect/preflight result = %s", actual)
	}

	popupPaused := make(chan struct{}, 1)
	driver.policy.interceptor.AddListener(func(_ context.Context, request *fetch.RequestPausedReply) bool {
		if request.Request.URL == denied.URL+"/popup" {
			select {
			case popupPaused <- struct{}{}:
			default:
			}
		}

		return true
	})
	policyEvaluate(t, ctx, p, fmt.Sprintf(`window.open(%q); true`, denied.URL+"/popup"))
	select {
	case <-popupPaused:
	case <-ctx.Done():
		t.Fatal("popup bypassed browser interception")
	}

	time.Sleep(100 * time.Millisecond)
	if count := forbidden.Load(); count != 0 {
		t.Fatalf("forbidden server received %d requests", count)
	}
}

func TestBrowserAllowedOriginsSchemesAndFrames(t *testing.T) {
	ctx, address, control := policyBrowser(t)
	var forbidden atomic.Int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forbidden.Add(1)
		_, _ = w.Write([]byte("unexpected"))
	})
	httpOnly := httptest.NewServer(handler)
	defer httpOnly.Close()
	tlsOnly := httptest.NewTLSServer(handler)
	defer tlsOnly.Close()

	// Listing the same host/port with the opposite scheme must block before
	// an HTTP request reaches the fixture handler.
	for _, test := range []struct{ allow, request string }{
		{strings.Replace(httpOnly.URL, "http:", "https:", 1), httpOnly.URL},
		{strings.Replace(tlsOnly.URL, "https:", "http:", 1), tlsOnly.URL},
		{strings.Replace(httpOnly.URL, "127.0.0.1", "localhost", 1), httpOnly.URL},
		{strings.Replace(httpOnly.URL, "127.0.0.1", "localhost", 1), strings.Replace(httpOnly.URL, "127.0.0.1", "child.localhost", 1)},
	} {
		driver := New(WithAddress(address), WithAllowedOrigins(test.allow))
		if _, err := driver.Open(ctx, drivers.Params{URL: test.request}); err == nil {
			t.Fatalf("unexpected navigation success for %s", test.request)
		}

		_ = driver.Close()
	}

	if forbidden.Load() != 0 {
		t.Fatal("scheme or hostname mismatch reached a forbidden fixture")
	}

	var bad atomic.Int32
	denied := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bad.Add(1)
	}))
	defer denied.Close()
	frame := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<script>fetch(%q).then(()=>parent.postMessage('frame-allowed','*'),()=>parent.postMessage('frame-blocked','*'))</script>`, denied.URL+"/from-frame")
	}))
	defer frame.Close()
	main := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body></body></html>"))
	}))
	defer main.Close()
	frameOrigin := strings.Replace(frame.URL, "127.0.0.1", "localhost", 1)
	driver := New(WithAddress(address), WithAllowedOrigins(main.URL, frameOrigin))
	defer driver.Close()
	p := policyPage(t, ctx, driver, main.URL)
	actual := policyEvaluate(t, ctx, p, fmt.Sprintf(`new Promise(resolve=>{window.addEventListener('message',e=>resolve(e.data),{once:true});const frame=document.createElement('iframe');frame.src=%q;document.body.appendChild(frame);})`, frameOrigin))
	if actual != `"frame-blocked"` || bad.Load() != 0 {
		t.Fatalf("cross-site frame result=%s forbidden requests=%d", actual, bad.Load())
	}

	targets, err := control.Target.GetTargets(ctx, target.NewGetTargetsArgs())
	if err != nil {
		t.Fatal(err)
	}

	for _, info := range targets.TargetInfos {
		if info.Type == "iframe" && strings.HasPrefix(info.URL, frameOrigin) {
			return
		}
	}

	t.Fatal("browser did not create the out-of-process iframe required by this test")
}

func TestBrowserAllowedOriginsWorkers(t *testing.T) {
	ctx, address, _ := policyBrowser(t)
	var forbidden atomic.Int32
	denied := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forbidden.Add(1)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		_, _ = w.Write([]byte("forbidden"))
	}))
	defer denied.Close()
	allowed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/worker.js" {
			w.Header().Set("Content-Type", "application/javascript")
			fmt.Fprintf(w, `fetch(%q).then(()=>postMessage('allowed'),()=>postMessage('blocked'));`, denied.URL+"/worker")

			return
		}

		if r.URL.Path == "/sw.js" {
			w.Header().Set("Content-Type", "application/javascript")
			fmt.Fprintf(w, `self.addEventListener('install',()=>self.skipWaiting()); self.addEventListener('activate',e=>e.waitUntil(self.clients.claim()));self.addEventListener('message',e=>e.waitUntil(fetch(%q).then(()=>e.source.postMessage('allowed'),()=>e.source.postMessage('blocked'))));`, denied.URL+"/service-worker")

			return
		}

		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>worker fixture</body></html>"))
	}))
	defer allowed.Close()
	driver := New(WithAddress(address), WithAllowedOrigins(allowed.URL))
	defer driver.Close()
	p := policyPage(t, ctx, driver, allowed.URL)
	if actual := policyEvaluate(t, ctx, p, `new Promise(resolve=>{window.worker=new Worker('/worker.js');worker.onmessage=e=>resolve(e.data);worker.onerror=()=>resolve('worker-error');})`); actual != `"blocked"` {
		t.Fatalf("worker result = %s", actual)
	}

	if actual := policyEvaluate(t, ctx, p, `(async()=>{await navigator.serviceWorker.register('/sw.js');const registration=await navigator.serviceWorker.ready;return await new Promise(resolve=>{navigator.serviceWorker.onmessage=e=>resolve(e.data);registration.active.postMessage('fetch');});})()`); actual != `"blocked"` {
		t.Fatalf("service worker result = %s", actual)
	}

	if forbidden.Load() != 0 {
		t.Fatal("worker reached forbidden server")
	}
}

func TestBrowserAllowedOriginsLifetime(t *testing.T) {
	ctx, address, control := policyBrowser(t)
	allowed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>fixture</body></html>"))
	}))
	defer allowed.Close()
	var forbidden atomic.Int32
	denied := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { forbidden.Add(1) }))
	defer denied.Close()
	driver := New(WithAddress(address), WithAllowedOrigins(allowed.URL))
	defer driver.Close()
	p := policyPage(t, ctx, driver, allowed.URL)
	policyConnection := driver.policy.conn
	_ = p.Close()
	paused := make(chan struct{}, 1)
	driver.policy.interceptor.AddListener(func(_ context.Context, request *fetch.RequestPausedReply) bool {
		if request.Request.URL == denied.URL+"/unrelated" {
			select {
			case paused <- struct{}{}:
			default:
			}
		}

		return true
	})

	otherContext, err := control.Target.CreateBrowserContext(ctx, target.NewCreateBrowserContextArgs().SetDisposeOnDetach(true))
	if err != nil {
		t.Fatal(err)
	}

	unrelated, err := control.Target.CreateTarget(ctx, target.NewCreateTargetArgs(denied.URL+"/unrelated").SetBrowserContextID(otherContext.BrowserContextID))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), policyCleanupTimeout)
		defer cancel()

		_, _ = control.Target.CloseTarget(cleanupCtx, target.NewCloseTargetArgs(unrelated.TargetID))
	})
	select {
	case <-paused:
	case <-ctx.Done():
		t.Fatal("unrelated context request was not intercepted")
	}

	if forbidden.Load() != 0 {
		t.Fatal("last page closure removed the policy from unrelated contexts")
	}

	p = policyPage(t, ctx, driver, allowed.URL)
	if driver.policy.conn != policyConnection {
		t.Fatal("driver replaced its persistent policy connection")
	}

	// A broader second policy must neither replace nor broaden the first.
	second := New(WithAddress(address), WithAllowedOrigins(allowed.URL, denied.URL))
	defer second.Close()
	secondPage := policyPage(t, ctx, second, allowed.URL)
	if _, err := second.Open(ctx, drivers.Params{URL: denied.URL}); err == nil || forbidden.Load() != 0 {
		t.Fatal("a second driver's policy broadened the existing restriction")
	}

	_ = secondPage.Close()
	_ = second.Close()
	if _, err := driver.Open(ctx, drivers.Params{URL: denied.URL}); err == nil || forbidden.Load() != 0 {
		t.Fatal("closing another driver disabled the surviving policy")
	}

	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			opened, openErr := driver.Open(ctx, drivers.Params{URL: allowed.URL})
			if openErr != nil {
				t.Errorf("concurrent Open: %v", openErr)

				return
			}

			_ = opened.Close()
		})
	}

	wg.Wait()
	if driver.policy.conn != policyConnection {
		t.Fatal("concurrent opens replaced the policy")
	}

	_ = policyConnection.Close()
	policyWait(t, func() bool {
		driver.policy.mu.Lock()
		defer driver.policy.mu.Unlock()

		return driver.policy.err != nil && len(driver.policy.sessions) == 0
	})
	if _, err := driver.Open(ctx, drivers.Params{URL: allowed.URL}); err == nil || !strings.Contains(err.Error(), "policy lost") {
		t.Fatalf("Open after policy loss = %v", err)
	}

	if _, err := driver.Parse(ctx, drivers.ParseParams{}); err == nil {
		t.Fatal("Parse accepted work after policy loss")
	}

	if err := p.Navigate(ctx, runtime.NewString(allowed.URL)); err == nil {
		t.Fatal("active page remained usable after policy loss")
	}
}

func TestBrowserAllowedOriginsOmittedAndIgnore(t *testing.T) {
	ctx, address, _ := policyBrowser(t)
	var images atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/image.png" {
			images.Add(1)

			return
		}

		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>fixture</body></html>"))
	}))
	defer server.Close()
	for _, restricted := range []bool{false, true} {
		options := []Option{WithAddress(address)}
		if restricted {
			options = append(options, WithAllowedOrigins(server.URL))
		}

		driver := New(options...)
		result, err := driver.Open(ctx, drivers.Params{
			URL:    server.URL,
			Ignore: &drivers.Ignore{Resources: []drivers.ResourceFilter{{Type: "Image", URL: "*/image.png"}}},
		})
		if err != nil {
			t.Fatal(err)
		}

		p := result.(*HTMLPage)
		if actual := policyEvaluate(t, ctx, p, `new Promise(resolve=>{const i=new Image();i.onerror=()=>resolve('blocked');i.onload=()=>resolve('allowed');i.src='/image.png';})`); actual != `"blocked"` || images.Load() != 0 {
			t.Fatalf("Ignore regression: result=%s requests=%d", actual, images.Load())
		}

		_ = p.Close()
		_ = driver.Close()
	}

	driver := New(WithAddress(address))
	p := policyPage(t, ctx, driver, server.URL)
	policyEvaluate(t, ctx, p, `fetch('/image.png').then(()=>true)`)
	if images.Load() != 1 {
		t.Fatal("omitted policy changed unrestricted behavior")
	}
}

func TestBrowserAllowedOriginsWebSocketGap(t *testing.T) {
	ctx, address, _ := policyBrowser(t)
	var handshakes atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
			handshakes.Add(1)
			http.Error(w, "handshake observed", http.StatusForbidden)

			return
		}

		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>fixture</body></html>"))
	}))
	defer server.Close()
	driver := New(WithAddress(address), WithAllowedOrigins(server.URL))
	defer driver.Close()
	p := policyPage(t, ctx, driver, server.URL)
	websocketURL := strings.Replace(server.URL, "http:", "ws:", 1) + "/socket"
	policyEvaluate(t, ctx, p, fmt.Sprintf(`new Promise(resolve=>{const ws=new WebSocket(%q);ws.onerror=()=>resolve('settled');ws.onopen=()=>{ws.close();resolve('open')};})`, websocketURL))
	if handshakes.Load() != 1 {
		t.Fatalf("WebSocket coverage changed; review documented limitation: handshakes=%d", handshakes.Load())
	}
}

func TestBrowserAllowedOriginsCloseCancelsRequests(t *testing.T) {
	ctx, address, _ := policyBrowser(t)
	started := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/hang" {
			started <- struct{}{}
			<-r.Context().Done()

			return
		}

		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>fixture</body></html>"))
	}))
	defer server.Close()
	driver := New(WithAddress(address), WithAllowedOrigins(server.URL))
	p := policyPage(t, ctx, driver, server.URL)
	finished := make(chan error, 1)
	go func() { finished <- p.Navigate(ctx, runtime.NewString(server.URL+"/hang")) }()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("navigation did not start")
	}

	closeStarted := time.Now()
	if err := driver.Close(); err != nil {
		t.Fatal(err)
	}

	if elapsed := time.Since(closeStarted); elapsed > policyCleanupTimeout+time.Second {
		t.Fatalf("driver shutdown exceeded its cleanup bound: %s", elapsed)
	}

	select {
	case err := <-finished:
		if err == nil {
			t.Fatal("pending navigation survived driver closure")
		}
	case <-ctx.Done():
		t.Fatal("driver closure did not unblock navigation")
	}

	if err := driver.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := driver.Open(ctx, drivers.Params{}); err == nil {
		t.Fatal("closed driver accepted work")
	}
}

func TestBrowserAllowedOriginsInvalidatesPendingOpen(t *testing.T) {
	for _, failure := range []string{"connection", "interception", "close"} {
		t.Run(failure, func(t *testing.T) {
			ctx, address, _ := policyBrowser(t)
			started := make(chan struct{}, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/pending" {
					started <- struct{}{}
					<-r.Context().Done()

					return
				}

				w.Header().Set("Content-Type", "text/html")
				_, _ = w.Write([]byte("<html><body>fixture</body></html>"))
			}))
			defer server.Close()
			driver := New(WithAddress(address), WithAllowedOrigins(server.URL))
			defer driver.Close()
			p := policyPage(t, ctx, driver, server.URL)
			finished := make(chan error, 1)
			go func() {
				_, err := driver.Open(ctx, drivers.Params{URL: server.URL + "/pending"})
				finished <- err
			}()

			select {
			case <-started:
			case <-ctx.Done():
				t.Fatal("pending Open did not start")
			}

			switch failure {
			case "connection":
				_ = driver.policy.conn.Close()
			case "interception":
				_ = driver.policy.interceptor.Close()
			case "close":
				_ = driver.Close()
			}

			select {
			case err := <-finished:
				if err == nil {
					t.Fatal("pending Open succeeded after policy loss")
				}
			case <-time.After(policyCleanupTimeout + time.Second):
				t.Fatal("pending Open was not canceled by policy loss")
			}

			policyWait(t, func() bool {
				p.mu.Lock()
				defer p.mu.Unlock()

				return bool(p.closed)
			})
			if _, err := driver.Open(ctx, drivers.Params{}); err == nil {
				t.Fatal("driver accepted work after losing its policy")
			}
		})
	}
}

func TestBrowserAllowedOriginsEmptyParse(t *testing.T) {
	ctx, address, _ := policyBrowser(t)
	var received atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { received.Add(1) }))
	defer server.Close()
	driver := New(WithAddress(address), WithAllowedOrigins())
	defer driver.Close()
	result, err := driver.Parse(ctx, drivers.ParseParams{Content: []byte("<html><body>parsed</body></html>")})
	if err != nil {
		t.Fatal(err)
	}

	p := result.(*HTMLPage)
	defer p.Close()
	if actual := policyEvaluate(t, ctx, p, fmt.Sprintf(`fetch(%q).then(()=>'allowed',()=> 'blocked')`, server.URL)); actual != `"blocked"` || received.Load() != 0 {
		t.Fatalf("empty policy result=%s requests=%d", actual, received.Load())
	}
}
