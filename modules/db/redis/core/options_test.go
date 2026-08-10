package core

import (
	"context"
	"errors"
	"testing"

	goredis "github.com/redis/go-redis/v9"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestParseClientOptions(t *testing.T) {
	t.Parallel()

	opts, err := parseClientOptions(" redis://ferret:secret@localhost:6380/3?protocol=2 ")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if opts.Addr != "localhost:6380" {
		t.Fatalf("unexpected address: %q", opts.Addr)
	}
	if opts.Username != "ferret" || opts.Password != "secret" {
		t.Fatalf("unexpected credentials: %q %q", opts.Username, opts.Password)
	}
	if opts.DB != 3 || opts.Protocol != 2 {
		t.Fatalf("unexpected database/protocol: %d/%d", opts.DB, opts.Protocol)
	}
	if !opts.ContextTimeoutEnabled {
		t.Fatal("expected context deadlines to be enabled")
	}

	tlsOpts, err := parseClientOptions("rediss://localhost:6380/0")
	if err != nil {
		t.Fatalf("unexpected TLS parse error: %v", err)
	}
	if tlsOpts.TLSConfig == nil {
		t.Fatal("expected rediss URL to enable TLS")
	}
}

func TestParseClientOptionsRejectsUnsupportedSources(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"", "unix:///tmp/redis.sock", "http://localhost:6379"} {
		_, err := parseClientOptions(input)
		if err == nil {
			t.Fatalf("expected %q to fail", input)
		}
	}
}

func TestDecodeOpenOptions(t *testing.T) {
	t.Parallel()

	opts, err := DecodeOpenOptions(t.Context(), runtime.NewObjectWith(map[string]runtime.Value{
		"url": runtime.NewString("redis://localhost:6379/0"),
	}))
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if opts.URL != "redis://localhost:6379/0" {
		t.Fatalf("unexpected URL: %q", opts.URL)
	}

	_, err = DecodeOpenOptions(t.Context(), runtime.NewObjectWith(map[string]runtime.Value{
		"uri": runtime.NewString("redis://localhost:6379"),
	}))
	assertErrorContains(t, err, "unknown field")
}

func TestOpenClosesClientWhenPingFails(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("ping failed")
	client := &fakeClient{
		pingFn: func(ctx context.Context) *goredis.StatusCmd {
			cmd := goredis.NewStatusCmd(ctx, "ping")
			cmd.SetErr(wantErr)

			return cmd
		},
	}

	_, err := openWithFactory(
		context.Background(),
		OpenOptions{URL: "redis://localhost:6379"},
		func(*goredis.Options) redisClient { return client },
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected ping error, got %v", err)
	}
	if client.closeCalls() != 1 {
		t.Fatalf("expected failed client to close once, got %d", client.closeCalls())
	}
}
