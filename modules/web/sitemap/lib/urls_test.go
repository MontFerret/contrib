package lib

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestURLs(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/index.xml":
			_, _ = fmt.Fprintf(w, `<sitemapindex><sitemap><loc>%s/posts.xml</loc></sitemap></sitemapindex>`, server.URL)
		case "/posts.xml":
			_, _ = fmt.Fprintf(w, `<urlset><url><loc>%s/post</loc></url></urlset>`, server.URL)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	empty, err := URLs(t.Context(),
		runtime.NewString(server.URL+"/index.xml"),
		runtime.NewObjectWith(map[string]runtime.Value{
			"recursive": runtime.False,
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	emptyArr := mustRuntimeArray(t, empty)
	if length, err := emptyArr.Length(t.Context()); err != nil || length != 0 {
		t.Fatalf("expected empty array, got %v %v", length, err)
	}

	result, err := URLs(t.Context(), runtime.NewString(server.URL+"/index.xml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := mustRuntimeArray(t, result)
	if length, err := arr.Length(t.Context()); err != nil || length != 1 {
		t.Fatalf("expected one URL, got %v %v", length, err)
	}
}
