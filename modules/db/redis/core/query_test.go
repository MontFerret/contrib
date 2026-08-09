package core

import (
	"context"
	"errors"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestQueryPreservesMixedArgumentOrder(t *testing.T) {
	t.Parallel()

	var got []any
	client := &fakeClient{
		commandFn: func(ctx context.Context) *goredis.CommandsInfoCmd {
			return commandInfoResult(ctx, map[string]*goredis.CommandInfo{
				"module.read": {ReadOnly: true},
			}, nil)
		},
		doFn: func(ctx context.Context, args ...any) *goredis.Cmd {
			got = append([]any(nil), args...)

			return commandResult(ctx, "OK", nil, args...)
		},
	}
	connection := newConnection(client)
	t.Cleanup(func() { _ = connection.Close() })

	out, err := connection.QueryOne(context.Background(), query(
		"redis",
		"MODULE.READ",
		runtime.NewString("key"),
		runtime.NewInt64(42),
		runtime.NewFloat(1.5),
		runtime.True,
		runtime.NewBinary([]byte{0, 1, 2}),
		runtime.NewString("NX"),
	))
	if err != nil {
		t.Fatalf("unexpected query error: %v", err)
	}
	if out != runtime.NewString("OK") {
		t.Fatalf("unexpected result: %v", out)
	}

	want := []any{"MODULE.READ", "key", int64(42), 1.5, true, []byte{0, 1, 2}, "NX"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected command arguments: got %#v, want %#v", got, want)
	}
}

func TestReadDialectRejectsMutationBeforeExecution(t *testing.T) {
	t.Parallel()

	var doCalls atomic.Int32
	client := &fakeClient{
		commandFn: func(ctx context.Context) *goredis.CommandsInfoCmd {
			return commandInfoResult(ctx, map[string]*goredis.CommandInfo{
				"set": {ReadOnly: false},
			}, nil)
		},
		doFn: func(ctx context.Context, args ...any) *goredis.Cmd {
			doCalls.Add(1)

			return commandResult(ctx, "OK", nil, args...)
		},
	}
	connection := newConnection(client)
	t.Cleanup(func() { _ = connection.Close() })

	_, err := connection.Query(context.Background(), query(
		"redis",
		"SET",
		runtime.NewString("key"),
		runtime.NewString("value"),
	))
	assertErrorContains(t, err, `command "SET" is not marked read-only; use redis_exec`)
	if doCalls.Load() != 0 {
		t.Fatalf("expected rejected mutation not to execute, got %d calls", doCalls.Load())
	}
}

func TestExecDialectDoesNotRequireMetadata(t *testing.T) {
	t.Parallel()

	var commandCalls atomic.Int32
	client := &fakeClient{
		commandFn: func(ctx context.Context) *goredis.CommandsInfoCmd {
			commandCalls.Add(1)

			return commandInfoResult(ctx, nil, errors.New("NOPERM command disabled"))
		},
		doFn: func(ctx context.Context, args ...any) *goredis.Cmd {
			return commandResult(ctx, "OK", nil, args...)
		},
	}
	connection := newConnection(client)
	t.Cleanup(func() { _ = connection.Close() })

	out, err := connection.QueryOne(context.Background(), query(
		"redis_exec",
		"SET",
		runtime.NewString("key"),
		runtime.NewString("value"),
	))
	if err != nil || out != runtime.NewString("OK") {
		t.Fatalf("unexpected exec result: %v, %v", out, err)
	}
	if commandCalls.Load() != 0 {
		t.Fatalf("expected redis_exec to bypass metadata, got %d calls", commandCalls.Load())
	}

	_, err = connection.Query(context.Background(), query("redis", "GET", runtime.NewString("key")))
	assertErrorContains(t, err, "cannot validate command")
}

func TestReadMetadataCachesAndRefreshesMissingCommand(t *testing.T) {
	t.Parallel()

	var commandCalls atomic.Int32
	client := &fakeClient{
		commandFn: func(ctx context.Context) *goredis.CommandsInfoCmd {
			call := commandCalls.Add(1)
			catalog := map[string]*goredis.CommandInfo{
				"get": {ReadOnly: true},
			}
			if call > 1 {
				catalog["module.read"] = &goredis.CommandInfo{ReadOnly: true}
			}

			return commandInfoResult(ctx, catalog, nil)
		},
		doFn: func(ctx context.Context, args ...any) *goredis.Cmd {
			return commandResult(ctx, "value", nil, args...)
		},
	}
	connection := newConnection(client)
	t.Cleanup(func() { _ = connection.Close() })

	for range 2 {
		if _, err := connection.QueryOne(context.Background(), query("redis", "GET", runtime.NewString("key"))); err != nil {
			t.Fatalf("unexpected cached query error: %v", err)
		}
	}
	if commandCalls.Load() != 1 {
		t.Fatalf("expected one catalog load, got %d", commandCalls.Load())
	}

	if _, err := connection.QueryOne(context.Background(), query("redis", "MODULE.READ")); err != nil {
		t.Fatalf("unexpected refreshed query error: %v", err)
	}
	if commandCalls.Load() != 2 {
		t.Fatalf("expected one refresh, got %d catalog calls", commandCalls.Load())
	}
}

func TestReadMetadataFailsClosedForUnknownCommand(t *testing.T) {
	t.Parallel()

	var commandCalls atomic.Int32
	var doCalls atomic.Int32
	client := &fakeClient{
		commandFn: func(ctx context.Context) *goredis.CommandsInfoCmd {
			commandCalls.Add(1)

			return commandInfoResult(ctx, map[string]*goredis.CommandInfo{
				"get": {ReadOnly: true},
			}, nil)
		},
		doFn: func(ctx context.Context, args ...any) *goredis.Cmd {
			doCalls.Add(1)

			return commandResult(ctx, "unexpected", nil, args...)
		},
	}
	connection := newConnection(client)
	t.Cleanup(func() { _ = connection.Close() })

	_, err := connection.Query(context.Background(), query("redis", "MODULE.UNKNOWN"))
	assertErrorContains(t, err, "absent from Redis command metadata; use redis_exec")
	if commandCalls.Load() != 2 {
		t.Fatalf("expected initial metadata load and one refresh, got %d calls", commandCalls.Load())
	}
	if doCalls.Load() != 0 {
		t.Fatalf("expected unknown read command not to execute, got %d calls", doCalls.Load())
	}
}

func TestRedisNilBecomesNone(t *testing.T) {
	t.Parallel()

	client := &fakeClient{
		doFn: func(ctx context.Context, args ...any) *goredis.Cmd {
			return commandResult(ctx, nil, goredis.Nil, args...)
		},
	}
	connection := newConnection(client)
	t.Cleanup(func() { _ = connection.Close() })

	value, err := connection.QueryOne(context.Background(), query("redis_exec", "GET", runtime.NewString("missing")))
	if err != nil {
		t.Fatalf("unexpected Redis nil error: %v", err)
	}
	if value != runtime.None {
		t.Fatalf("expected Redis nil to become NONE, got %v", value)
	}
}

func TestDecodeExecutionTimeoutForms(t *testing.T) {
	t.Parallel()

	cases := []struct {
		value runtime.Value
		name  string
		want  time.Duration
	}{
		{name: "duration", value: runtime.NewDuration(2 * time.Second), want: 2 * time.Second},
		{name: "string", value: runtime.NewString("250ms"), want: 250 * time.Millisecond},
		{name: "integer milliseconds", value: runtime.NewInt(25), want: 25 * time.Millisecond},
		{name: "fractional milliseconds", value: runtime.NewFloat(1.5), want: 1500 * time.Microsecond},
		{name: "zero", value: runtime.ZeroInt, want: 0},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			options, err := decodeExecutionOptions(context.Background(), runtime.NewObjectWith(map[string]runtime.Value{
				"timeout": tt.value,
			}))
			if err != nil {
				t.Fatalf("unexpected timeout decode error: %v", err)
			}
			if options.Timeout != tt.want {
				t.Fatalf("unexpected timeout: got %v, want %v", options.Timeout, tt.want)
			}
		})
	}
}

func TestQueryTimeoutCancelsMetadataAndExecution(t *testing.T) {
	t.Parallel()

	client := &fakeClient{
		doFn: func(ctx context.Context, args ...any) *goredis.Cmd {
			<-ctx.Done()

			return commandResult(ctx, nil, ctx.Err(), args...)
		},
	}
	connection := newConnection(client)
	t.Cleanup(func() { _ = connection.Close() })

	q := query("redis_exec", "BLPOP", runtime.NewString("key"), runtime.NewInt(1))
	q.Options = runtime.NewObjectWith(map[string]runtime.Value{
		"timeout": runtime.NewString("10ms"),
	})

	started := time.Now()
	_, err := connection.Query(context.Background(), q)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline error, got %v", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("query timeout took too long: %v", elapsed)
	}
}

func TestQueryValidation(t *testing.T) {
	t.Parallel()

	connection := newConnection(&fakeClient{})
	t.Cleanup(func() { _ = connection.Close() })

	cases := []struct {
		name string
		q    runtime.Query
		want string
	}{
		{
			name: "dialect",
			q:    query("sql", "GET"),
			want: `unsupported dialect "sql"`,
		},
		{
			name: "empty command",
			q:    query("redis", ""),
			want: "must not be empty",
		},
		{
			name: "compound command",
			q:    query("redis", "CLIENT LIST"),
			want: "must be a single token",
		},
		{
			name: "with not object",
			q: runtime.Query{
				Kind:       runtime.NewString("redis_exec"),
				Expression: runtime.NewString("PING"),
				Params:     runtime.NewString("bad"),
			},
			want: "WITH must be an object",
		},
		{
			name: "unknown with field",
			q: runtime.Query{
				Kind:       runtime.NewString("redis_exec"),
				Expression: runtime.NewString("GET"),
				Params: runtime.NewObjectWith(map[string]runtime.Value{
					"key": runtime.NewString("value"),
				}),
			},
			want: "unsupported field",
		},
		{
			name: "params not array",
			q: runtime.Query{
				Kind:       runtime.NewString("redis_exec"),
				Expression: runtime.NewString("GET"),
				Params: runtime.NewObjectWith(map[string]runtime.Value{
					"params": runtime.NewString("key"),
				}),
			},
			want: "WITH.params must be an array",
		},
		{
			name: "none argument",
			q:    query("redis_exec", "SET", runtime.None),
			want: "NONE is not a supported",
		},
		{
			name: "complex argument",
			q:    query("redis_exec", "SET", runtime.NewObject()),
			want: "unsupported Redis argument type Object",
		},
		{
			name: "options not object",
			q: runtime.Query{
				Kind:       runtime.NewString("redis_exec"),
				Expression: runtime.NewString("PING"),
				Options:    runtime.NewString("bad"),
			},
			want: "OPTIONS must be an object",
		},
		{
			name: "unknown option",
			q: runtime.Query{
				Kind:       runtime.NewString("redis_exec"),
				Expression: runtime.NewString("PING"),
				Options: runtime.NewObjectWith(map[string]runtime.Value{
					"retry": runtime.True,
				}),
			},
			want: "unsupported field",
		},
		{
			name: "negative timeout",
			q: runtime.Query{
				Kind:       runtime.NewString("redis_exec"),
				Expression: runtime.NewString("PING"),
				Options: runtime.NewObjectWith(map[string]runtime.Value{
					"timeout": runtime.NewString("-1ms"),
				}),
			},
			want: "greater than or equal to 0",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := connection.Query(context.Background(), tt.q)
			assertErrorContains(t, err, tt.want)
		})
	}
}

func TestRedisServerErrorsSurface(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("ERR wrong number of arguments")
	client := &fakeClient{
		doFn: func(ctx context.Context, args ...any) *goredis.Cmd {
			return commandResult(ctx, nil, wantErr, args...)
		},
	}
	connection := newConnection(client)
	t.Cleanup(func() { _ = connection.Close() })

	_, err := connection.Query(context.Background(), query("redis_exec", "GET"))
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected server error, got %v", err)
	}
}
