package core

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecodeClientConfigString(t *testing.T) {
	t.Parallel()

	cfg, err := DecodeClientConfig(context.Background(), runtime.NewString("https://api.example.test"))
	if err != nil {
		t.Fatalf("unexpected config error: %v", err)
	}

	if cfg.BaseURL != "https://api.example.test" {
		t.Fatalf("expected base URL, got %q", cfg.BaseURL)
	}
	if cfg.RequestEncoding != EncodingJSON || cfg.ResponseEncoding != EncodingJSON {
		t.Fatalf("expected JSON defaults, got %q/%q", cfg.RequestEncoding, cfg.ResponseEncoding)
	}
	if cfg.ResponseMode != ResponseModeBody {
		t.Fatalf("expected body response mode, got %q", cfg.ResponseMode)
	}
	if cfg.ErrorMode != ErrorModeRaise {
		t.Fatalf("expected raise error mode, got %q", cfg.ErrorMode)
	}
}

func TestDecodeClientConfigObject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg, err := DecodeClientConfig(ctx, object(t, map[string]runtime.Value{
		"baseUrl": runtime.NewString("https://api.example.test"),
		"headers": object(t, map[string]runtime.Value{
			"Authorization": runtime.NewString("Bearer token"),
			"X-Multi":       runtime.NewArrayWith(runtime.NewString("one"), runtime.NewString("two")),
			"X-None":        runtime.None,
		}),
		"encoding":         runtime.NewString("text"),
		"requestEncoding":  runtime.NewString("form"),
		"responseEncoding": runtime.NewString("bytes"),
		"timeout":          runtime.NewInt(2500),
		"response":         runtime.NewString("full"),
		"errorMode":        runtime.NewString("response"),
	}))
	if err != nil {
		t.Fatalf("unexpected config error: %v", err)
	}

	if cfg.Headers.Get("Authorization") != "Bearer token" {
		t.Fatalf("expected authorization header, got %q", cfg.Headers.Get("Authorization"))
	}
	if got := cfg.Headers.Values("X-Multi"); len(got) != 2 || got[0] != "one" || got[1] != "two" {
		t.Fatalf("unexpected repeated header values: %v", got)
	}
	if cfg.Headers.Get("X-None") != "" {
		t.Fatalf("expected none header to be ignored, got %q", cfg.Headers.Get("X-None"))
	}
	if cfg.RequestEncoding != EncodingForm {
		t.Fatalf("expected form request encoding, got %q", cfg.RequestEncoding)
	}
	if cfg.ResponseEncoding != EncodingBytes {
		t.Fatalf("expected bytes response encoding, got %q", cfg.ResponseEncoding)
	}
	if cfg.Timeout != int64(2500*time.Millisecond) {
		t.Fatalf("expected 2500ms timeout, got %d", cfg.Timeout)
	}
	if cfg.ResponseMode != ResponseModeFull {
		t.Fatalf("expected full response mode, got %q", cfg.ResponseMode)
	}
	if cfg.ErrorMode != ErrorModeResponse {
		t.Fatalf("expected response error mode, got %q", cfg.ErrorMode)
	}
}

func TestDecodeClientConfigErrors(t *testing.T) {
	t.Parallel()

	_, err := DecodeClientConfig(context.Background(), object(t, map[string]runtime.Value{}))
	if err == nil {
		t.Fatal("expected missing base URL error")
	}
	if !strings.Contains(err.Error(), "NET::REST::CLIENT config.baseUrl is required") {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = DecodeClientConfig(context.Background(), object(t, map[string]runtime.Value{
		"baseUrl":  runtime.NewString("https://api.example.test"),
		"encoding": runtime.NewString("xml"),
	}))
	if err == nil {
		t.Fatal("expected unsupported encoding error")
	}
	if !strings.Contains(err.Error(), `NET::REST::CLIENT config.encoding: unsupported encoding "xml"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
