package core

import (
	"context"
	"testing"

	goredis "github.com/redis/go-redis/v9"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestConnectionResourceAndClose(t *testing.T) {
	t.Parallel()

	client := &fakeClient{}
	connection := newConnection(client)

	if connection.String() != "<db.redis.connection>" {
		t.Fatalf("unexpected connection display: %s", connection.String())
	}
	if connection.Copy() != connection {
		t.Fatal("expected connection copy to preserve identity")
	}
	if connection.ResourceID() == 0 {
		t.Fatal("expected non-zero resource ID")
	}

	if err := connection.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
	if err := connection.Close(); err != nil {
		t.Fatalf("unexpected idempotent close error: %v", err)
	}
	if client.closeCalls() != 1 {
		t.Fatalf("expected client close once, got %d", client.closeCalls())
	}

	_, err := connection.Query(context.Background(), query("redis_exec", "PING"))
	assertErrorContains(t, err, "redis connection has been closed")
}

func TestConnectionQueryModifiersUseResultList(t *testing.T) {
	t.Parallel()

	client := &fakeClient{
		doFn: func(ctx context.Context, args ...any) *goredis.Cmd {
			return commandResult(ctx, []any{"first", "second"}, nil, args...)
		},
	}
	connection := newConnection(client)
	t.Cleanup(func() { _ = connection.Close() })

	q := query("redis_exec", "LRANGE")
	one, err := connection.QueryOne(context.Background(), q)
	if err != nil || one != runtime.NewString("first") {
		t.Fatalf("unexpected QUERY ONE result: %v, %v", one, err)
	}

	count, err := connection.QueryCount(context.Background(), q)
	if err != nil || count != 2 {
		t.Fatalf("unexpected QUERY COUNT result: %v, %v", count, err)
	}

	exists, err := connection.QueryExists(context.Background(), q)
	if err != nil || !exists {
		t.Fatalf("unexpected QUERY EXISTS result: %v, %v", exists, err)
	}
}
