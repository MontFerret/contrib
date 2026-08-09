package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/MontFerret/ferret/v2"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk/sdktest"
)

func TestNewSmoke(t *testing.T) {
	module := New()
	if module == nil {
		t.Fatal("expected module to be non-nil")
	}
	if module.Name() != "db/redis" {
		t.Fatalf("expected module name %q, got %q", "db/redis", module.Name())
	}
}

func TestOpenValidationThroughFerret(t *testing.T) {
	t.Parallel()

	_, err := runFQL(t, `RETURN DB::REDIS::OPEN({})`)
	assertErrorContains(t, err, "url is required")

	_, err = runFQL(t, `RETURN DB::REDIS::OPEN({ url: "unix:///tmp/redis.sock" })`)
	assertErrorContains(t, err, "unsupported Redis URL scheme")
}

func TestCloseRejectsWrongHandleThroughFerret(t *testing.T) {
	t.Parallel()

	_, err := runFQL(t, `RETURN DB::REDIS::CLOSE("invalid")`)
	assertErrorContains(t, err, "expected Redis connection handle")
}

func TestIntegrationRedisThroughFerret(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL is not set")
	}

	prefix := fmt.Sprintf("ferret:redis:%d", time.Now().UnixNano())
	stringKey := prefix + ":string"
	hashKey := prefix + ":hash"
	listKey := prefix + ":list"
	counterKey := prefix + ":counter"
	missingKey := prefix + ":missing"

	t.Cleanup(func() {
		_, _ = runFQL(t, `
			LET redis = DB::REDIS::OPEN({ url: @redisURL })
			LET deleted = QUERY ONE "DEL" IN redis USING redis_exec WITH {
				params: [@stringKey, @hashKey, @listKey, @counterKey, @missingKey]
			}
			RETURN DB::REDIS::CLOSE(redis)
		`,
			ferret.WithRuntimeParam("redisURL", runtime.NewString(redisURL)),
			ferret.WithRuntimeParam("stringKey", runtime.NewString(stringKey)),
			ferret.WithRuntimeParam("hashKey", runtime.NewString(hashKey)),
			ferret.WithRuntimeParam("listKey", runtime.NewString(listKey)),
			ferret.WithRuntimeParam("counterKey", runtime.NewString(counterKey)),
			ferret.WithRuntimeParam("missingKey", runtime.NewString(missingKey)),
		)
	})

	output, err := runFQL(t, `
		LET redis = DB::REDIS::OPEN({ url: @redisURL })
		LET set = QUERY ONE "SET" IN redis USING redis_exec WITH {
			params: [@stringKey, "Tim", "EX", 60, "NX"]
		}
		LET blockedSet = QUERY ONE "SET" IN redis USING redis_exec WITH {
			params: [@stringKey, "Other", "EX", 60, "NX"]
		}
		LET value = QUERY ONE "GET" IN redis USING redis WITH {
			params: [@stringKey]
		}
		LET hset = QUERY ONE "HSET" IN redis USING redis_exec WITH {
			params: [@hashKey, "name", "Tim", "age", 42, "score", 1.5, "active", true]
		}
		LET profile = QUERY ONE "HGETALL" IN redis USING redis WITH {
			params: [@hashKey]
		}
		LET pushed = QUERY ONE "RPUSH" IN redis USING redis_exec WITH {
			params: [@listKey, "first", "second", "third"]
		}
		LET items = QUERY "LRANGE" IN redis USING redis WITH {
			params: [@listKey, 0, -1]
		}
		LET firstIncrement = QUERY ONE "INCR" IN redis USING redis_exec WITH {
			params: [@counterKey]
		}
		LET secondIncrement = QUERY ONE "INCR" IN redis USING redis_exec WITH {
			params: [@counterKey]
		}
		LET missing = QUERY ONE "GET" IN redis USING redis WITH {
			params: [@missingKey]
		}
		LET deleted = QUERY ONE "DEL" IN redis USING redis_exec WITH {
			params: [@counterKey]
		}
		LET closed = DB::REDIS::CLOSE(redis)
		RETURN {
			set,
			blockedSet,
			value,
			hset,
			profile,
			pushed,
			items,
			firstIncrement,
			secondIncrement,
			missing,
			deleted,
			closed
		}
	`,
		ferret.WithRuntimeParam("redisURL", runtime.NewString(redisURL)),
		ferret.WithRuntimeParam("stringKey", runtime.NewString(stringKey)),
		ferret.WithRuntimeParam("hashKey", runtime.NewString(hashKey)),
		ferret.WithRuntimeParam("listKey", runtime.NewString(listKey)),
		ferret.WithRuntimeParam("counterKey", runtime.NewString(counterKey)),
		ferret.WithRuntimeParam("missingKey", runtime.NewString(missingKey)),
	)
	if err != nil {
		t.Fatalf("unexpected Redis integration error: %v", err)
	}
	assertIntegrationOutput(t, output)

	_, err = runFQL(t, `
		LET redis = DB::REDIS::OPEN({ url: @redisURL })
		RETURN QUERY ONE "SET" IN redis USING redis WITH {
			params: [@stringKey, "blocked"]
		}
	`,
		ferret.WithRuntimeParam("redisURL", runtime.NewString(redisURL)),
		ferret.WithRuntimeParam("stringKey", runtime.NewString(stringKey)),
	)
	assertErrorContains(t, err, "is not marked read-only; use redis_exec")

	_, err = runFQL(t, `
		LET redis = DB::REDIS::OPEN({ url: @redisURL })
		RETURN QUERY ONE "GET" IN redis USING redis
	`, ferret.WithRuntimeParam("redisURL", runtime.NewString(redisURL)))
	assertErrorContains(t, err, "wrong number of arguments")

	_, err = runFQL(t, `
		LET redis = DB::REDIS::OPEN({ url: @redisURL })
		RETURN QUERY ONE "FERRET_UNKNOWN_COMMAND" IN redis USING redis_exec
	`, ferret.WithRuntimeParam("redisURL", runtime.NewString(redisURL)))
	assertErrorContains(t, err, "unknown command")

	_, err = runFQL(t, `
		LET redis = DB::REDIS::OPEN({ url: @redisURL })
		RETURN QUERY ONE "BLPOP" IN redis USING redis_exec WITH {
			params: [@missingKey, 1]
		} OPTIONS {
			timeout: "10ms"
		}
	`,
		ferret.WithRuntimeParam("redisURL", runtime.NewString(redisURL)),
		ferret.WithRuntimeParam("missingKey", runtime.NewString(missingKey)),
	)
	if err == nil {
		t.Fatal("expected blocking command to observe query timeout")
	}
}

func assertIntegrationOutput(t *testing.T, output *ferret.Output) {
	t.Helper()

	var actual struct {
		BlockedSet      any               `json:"blockedSet"`
		Missing         any               `json:"missing"`
		Profile         map[string]string `json:"profile"`
		Set             string            `json:"set"`
		Value           string            `json:"value"`
		Items           []string          `json:"items"`
		HSet            int               `json:"hset"`
		Pushed          int               `json:"pushed"`
		FirstIncrement  int               `json:"firstIncrement"`
		SecondIncrement int               `json:"secondIncrement"`
		Deleted         int               `json:"deleted"`
		Closed          bool              `json:"closed"`
	}
	if err := json.Unmarshal(output.Content, &actual); err != nil {
		t.Fatalf("failed to decode Redis integration output: %v", err)
	}

	if actual.Set != "OK" || actual.BlockedSet != nil || actual.Value != "Tim" || actual.HSet != 4 ||
		actual.Profile["name"] != "Tim" || actual.Profile["age"] != "42" ||
		actual.Profile["score"] != "1.5" || actual.Profile["active"] != "1" ||
		actual.Pushed != 3 || !equalStrings(actual.Items, []string{"first", "second", "third"}) ||
		actual.FirstIncrement != 1 || actual.SecondIncrement != 2 || actual.Missing != nil ||
		actual.Deleted != 1 || !actual.Closed {
		t.Fatalf("unexpected Redis integration output: %#v", actual)
	}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}

	return true
}

func runFQL(t *testing.T, query string, opts ...ferret.Option) (*ferret.Output, error) {
	t.Helper()

	engineOptions := append([]ferret.Option{ferret.WithModules(New())}, opts...)
	harness := sdktest.New(t, engineOptions...)

	return harness.Run(context.Background(), query)
}

func assertOutputBool(t *testing.T, output *ferret.Output, expected bool) {
	t.Helper()

	var actual bool
	if err := json.Unmarshal(output.Content, &actual); err != nil {
		t.Fatalf("failed to decode output bool: %v", err)
	}
	if actual != expected {
		t.Fatalf("expected output %v, got %v", expected, actual)
	}
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
