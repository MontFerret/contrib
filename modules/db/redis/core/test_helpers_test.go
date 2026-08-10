package core

import (
	"context"
	"strings"
	"testing"

	goredis "github.com/redis/go-redis/v9"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func commandResult(ctx context.Context, value any, err error, args ...any) *goredis.Cmd {
	cmd := goredis.NewCmd(ctx, args...)
	cmd.SetVal(value)
	cmd.SetErr(err)

	return cmd
}

func commandInfoResult(
	ctx context.Context,
	value map[string]*goredis.CommandInfo,
	err error,
) *goredis.CommandsInfoCmd {
	cmd := goredis.NewCommandsInfoCmd(ctx, "command")
	cmd.SetVal(value)
	cmd.SetErr(err)

	return cmd
}

func query(kind, template string, bindings ...map[string]runtime.Value) runtime.Query {
	q := runtime.Query{
		Kind:       runtime.NewString(kind),
		Expression: runtime.NewString(template),
	}
	if len(bindings) != 0 {
		q.Params = runtime.NewObjectWith(bindings[0])
	}

	return q
}

func assertErrorContains(t *testing.T, err error, expected string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error containing %q", expected)
	}
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error containing %q, got %q", expected, err.Error())
	}
}

func objectValue(t *testing.T, ctx context.Context, object runtime.Map, key string) runtime.Value {
	t.Helper()

	value, found, err := object.Lookup(ctx, runtime.NewString(key))
	if err != nil {
		t.Fatalf("lookup %q: %v", key, err)
	}
	if !found {
		t.Fatalf("missing object field %q", key)
	}

	return value
}
