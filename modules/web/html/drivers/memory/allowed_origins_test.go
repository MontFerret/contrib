package memory

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

func TestAllowedOriginRequests(t *testing.T) {
	var forbiddenHits, redirectHits, retryHits atomic.Int32
	forbidden := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		forbiddenHits.Add(1)
	}))
	defer forbidden.Close()

	listed := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "<title>listed destination</title>")
	}))
	defer listed.Close()

	start := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/retry":
			if retryHits.Add(1) == 1 {
				w.WriteHeader(http.StatusServiceUnavailable)

				return
			}

			http.Redirect(w, r, "/allowed", http.StatusTemporaryRedirect)
		case "/allowed":
			http.Redirect(w, r, listed.URL+"/document", http.StatusFound)
		case "/blocked":
			redirectHits.Add(1)
			http.Redirect(w, r, forbidden.URL+"/private?token=secret", http.StatusFound)
		default:
			http.NotFound(w, r)
		}
	}))
	defer start.Close()

	drv := New(
		WithAllowedOrigins(start.URL, listed.URL),
		WithCustomTransport(listed.Client().Transport.(*http.Transport)),
		WithMaxRetries(3),
		func(opts *Options) { opts.Backoff = func(int) time.Duration { return 0 } },
	)
	defer drv.Close()

	page, err := drv.Open(t.Context(), drivers.Params{URL: start.URL + "/retry"})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "listed destination", page.GetMainFrame().GetTitle().String())
	assert.NoError(t, page.Close())
	assert.EqualValues(t, 2, retryHits.Load())

	for _, url := range []string{forbidden.URL, start.URL + "/blocked", "about:blank"} {
		_, err := drv.Open(t.Context(), drivers.Params{URL: url})
		assert.ErrorIs(t, err, errOriginDenied)
		if url == start.URL+"/blocked" && err != nil {
			assert.ErrorContains(t, err, forbidden.URL)
			assert.NotContains(t, err.Error(), "private")
			assert.NotContains(t, err.Error(), "secret")
		}
	}

	assert.EqualValues(t, 1, redirectHits.Load())
	assert.Zero(t, forbiddenHits.Load())
}

func TestAllowedOriginConfiguration(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		hits.Add(1)
	}))
	defer server.Close()

	input := []string{server.URL}
	option := WithAllowedOrigins(input...)
	input[0] = "https://other.example"
	for _, test := range []struct {
		name    string
		options []Option
		denied  bool
	}{
		{name: "omitted"},
		{name: "empty", options: []Option{WithAllowedOrigins()}, denied: true},
		{name: "snapshot replaces invalid", options: []Option{WithAllowedOrigins("bad"), option}},
		{name: "last replaces snapshot", options: []Option{option, WithAllowedOrigins()}, denied: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			driver := New(test.options...)
			defer driver.Close()

			before := hits.Load()
			page, err := driver.Open(t.Context(), drivers.Params{URL: server.URL})
			if test.denied {
				assert.ErrorIs(t, err, errOriginDenied)
				assert.Equal(t, before, hits.Load())
			} else {
				if err != nil {
					t.Fatal(err)
				}

				assert.NoError(t, page.Close())
				assert.Equal(t, before+1, hits.Load())
			}

			before = hits.Load()
			page, err = driver.Parse(t.Context(), drivers.ParseParams{Content: []byte("<img src='" + server.URL + "'>")})
			if err != nil {
				t.Fatal(err)
			}

			assert.NoError(t, page.Close())
			assert.Equal(t, before, hits.Load())
		})
	}

	before := hits.Load()
	driver := New(option, WithAllowedOrigins("https://user:secret@example.com"))
	defer driver.Close()

	_, err := driver.Open(t.Context(), drivers.Params{URL: server.URL})
	if !assert.ErrorContains(t, err, "allowed origin 1") {
		t.FailNow()
	}

	assert.NotContains(t, err.Error(), "secret")

	_, err = driver.Parse(t.Context(), drivers.ParseParams{})
	assert.ErrorContains(t, err, "allowed origin 1")
	assert.Equal(t, before, hits.Load())
}

func TestAllowedOriginProxy(t *testing.T) {
	var hits atomic.Int32
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		hits.Add(1)
		fmt.Fprint(w, "<title>"+req.URL.Host+"</title>")
	}))
	defer proxy.Close()

	driver := New(WithProxy(proxy.URL), WithAllowedOrigins("http://listed.test", "http://İ.test"))
	defer driver.Close()

	page, err := driver.Open(t.Context(), drivers.Params{URL: "http://LISTED.test:80/doc"})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "LISTED.test:80", page.GetMainFrame().GetTitle().String())
	assert.NoError(t, page.Close())

	page, err = driver.Open(t.Context(), drivers.Params{URL: "http://İ.test/doc"})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "xn--i-9bb.test", page.GetMainFrame().GetTitle().String())
	assert.NoError(t, page.Close())

	for _, url := range []string{"https://listed.test", "http://listed.test:81", "http://sub.listed.test", "http://i.test"} {
		_, err := driver.Open(t.Context(), drivers.Params{URL: url})
		assert.ErrorIs(t, err, errOriginDenied)
	}

	assert.EqualValues(t, 2, hits.Load())
}

func TestAllowedOriginCancellationIsPerOpen(t *testing.T) {
	started := make(chan string, 2)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		started <- req.URL.Path

		select {
		case <-release:
			fmt.Fprint(w, "<title>ready</title>")
		case <-req.Context().Done():
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	driver := New(WithAllowedOrigins(server.URL))
	defer driver.Close()

	open := func(ctx context.Context, path string) <-chan error {
		done := make(chan error, 1)
		go func() {
			page, err := driver.Open(ctx, drivers.Params{URL: server.URL + path})
			if err == nil {
				err = page.Close()
			}

			done <- err
		}()

		select {
		case actual := <-started:
			assert.Equal(t, path, actual)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}

		return done
	}

	waiting := open(ctx, "/waiting")
	caller, cancelCaller := context.WithCancel(ctx)
	defer cancelCaller()

	canceled := open(caller, "/cancel")
	cancelCaller()
	assert.ErrorIs(t, <-canceled, context.Canceled)

	_, err := driver.Open(ctx, drivers.Params{URL: "http://denied.example"})
	assert.ErrorIs(t, err, errOriginDenied)

	other := New(WithAllowedOrigins())
	defer other.Close()

	_, err = other.Open(ctx, drivers.Params{URL: server.URL})
	assert.ErrorIs(t, err, errOriginDenied)

	close(release)
	assert.NoError(t, <-waiting)
}
