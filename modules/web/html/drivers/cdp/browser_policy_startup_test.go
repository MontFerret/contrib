package cdp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

func TestBrowserPolicyStartupFailsClosed(t *testing.T) {
	for _, cancellation := range []bool{false, true} {
		t.Run(map[bool]string{false: "unsupported Fetch", true: "canceled startup"}[cancellation], func(t *testing.T) {
			var targets atomic.Int32
			var connections atomic.Int32
			var endpoint string
			upgrader := websocket.Upgrader{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/json/version" {
					_ = json.NewEncoder(w).Encode(map[string]string{"webSocketDebuggerUrl": strings.Replace(endpoint, "http:", "ws:", 1) + "/browser"})

					return
				}

				conn, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					return
				}

				connections.Add(1)
				defer conn.Close()

				for {
					var request struct {
						ID     uint64 `json:"id"`
						Method string `json:"method"`
					}

					if err := conn.ReadJSON(&request); err != nil {
						return
					}

					if strings.HasPrefix(request.Method, "Target.") {
						targets.Add(1)
					}

					if request.Method == "Fetch.enable" && !cancellation {
						_ = conn.WriteJSON(map[string]any{"id": request.ID, "error": map[string]any{"code": -32601, "message": "Fetch unavailable"}})
					}
				}
			}))
			defer server.Close()
			endpoint = server.URL
			driver := New(WithAddress(endpoint), WithAllowedOrigins("https://example.com"))
			defer driver.Close()
			ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
			defer cancel()

			_, err := driver.Open(ctx, drivers.Params{URL: "https://example.com"})
			if err == nil || (cancellation && !errors.Is(err, context.DeadlineExceeded)) {
				t.Fatalf("startup failure = %v", err)
			}

			if targets.Load() != 0 || connections.Load() != 1 {
				t.Fatalf("failed policy created work: target commands=%d connections=%d", targets.Load(), connections.Load())
			}

			if driver.policy.conn != nil {
				t.Fatal("failed startup retained its connection")
			}
		})
	}
}
