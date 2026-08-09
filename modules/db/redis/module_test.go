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

	id := fmt.Sprintf("ferret:redis:%d", time.Now().UnixNano())
	sessionKey := "session:" + id
	profileKey := "profile:" + id
	rolesKey := "roles:" + id
	listKey := "list:" + id
	counterKey := "counter:" + id
	missingKey := "missing:" + id
	mgetFirstKey := "mget:first:" + id
	mgetSecondKey := "mget:second:" + id
	quotedKey := "quoted:" + id

	t.Cleanup(func() {
		_, _ = runFQL(t, `
			LET redis = DB::REDIS::OPEN({ url: @redisURL })
			LET deleted = QUERY ONE "DEL $keys..." IN redis USING redis_exec WITH {
				keys: [
					@sessionKey,
					@profileKey,
					@rolesKey,
					@listKey,
					@counterKey,
					@missingKey,
					@mgetFirstKey,
					@mgetSecondKey,
					@quotedKey
				]
			}
			RETURN DB::REDIS::CLOSE(redis)
		`,
			ferret.WithRuntimeParam("redisURL", runtime.NewString(redisURL)),
			ferret.WithRuntimeParam("sessionKey", runtime.NewString(sessionKey)),
			ferret.WithRuntimeParam("profileKey", runtime.NewString(profileKey)),
			ferret.WithRuntimeParam("rolesKey", runtime.NewString(rolesKey)),
			ferret.WithRuntimeParam("listKey", runtime.NewString(listKey)),
			ferret.WithRuntimeParam("counterKey", runtime.NewString(counterKey)),
			ferret.WithRuntimeParam("missingKey", runtime.NewString(missingKey)),
			ferret.WithRuntimeParam("mgetFirstKey", runtime.NewString(mgetFirstKey)),
			ferret.WithRuntimeParam("mgetSecondKey", runtime.NewString(mgetSecondKey)),
			ferret.WithRuntimeParam("quotedKey", runtime.NewString(quotedKey)),
		)
	})

	output, err := runFQL(t, `
		LET redis = DB::REDIS::OPEN({ url: @redisURL })
		LET set = QUERY ONE "SET session:$id $value EX $ttl NX" IN redis USING redis_exec WITH {
			id: @id,
			value: "hello world",
			ttl: 60
		}
		LET blockedSet = QUERY ONE "SET session:$id $value EX $ttl NX" IN redis USING redis_exec WITH {
			id: @id,
			value: "Other",
			ttl: 60
		}
		LET value = QUERY ONE "GET session:$id" IN redis USING redis WITH {
			id: @id
		}
		LET hset = QUERY ONE "HSET profile:$id name $name age $age score $score active $active" IN redis USING redis_exec WITH {
			id: @id,
			name: "Tim",
			age: 42,
			score: 1.5,
			active: true
		}
		LET profile = QUERY ONE "HGETALL profile:$id" IN redis USING redis WITH {
			id: @id
		}
		LET mset = QUERY ONE "MSET $pairs..." IN redis USING redis_exec WITH {
			pairs: [@mgetFirstKey, "first", @mgetSecondKey, "second"]
		}
		LET mget = QUERY "MGET $keys..." IN redis USING redis WITH {
			keys: [@mgetFirstKey, @mgetSecondKey]
		}
		LET quotedSet = QUERY ONE 'SET $key "hello quoted world"' IN redis USING redis_exec WITH {
			key: @quotedKey
		}
		LET quotedValue = QUERY ONE "GET $key" IN redis USING redis WITH {
			key: @quotedKey
		}
		LET sadd = QUERY ONE "SADD roles:$id $roles..." IN redis USING redis_exec WITH {
			id: @id,
			roles: ["admin", "editor"]
		}
		LET roleCount = QUERY ONE "SCARD roles:$id" IN redis USING redis WITH {
			id: @id
		}
		LET pushed = QUERY ONE "RPUSH $key $items..." IN redis USING redis_exec WITH {
			key: @listKey,
			items: ["first", "second", "third"]
		}
		LET items = QUERY "LRANGE $key $start $stop" IN redis USING redis WITH {
			key: @listKey,
			start: 0,
			stop: -1
		}
		LET firstIncrement = QUERY ONE "INCR $key" IN redis USING redis_exec WITH {
			key: @counterKey
		}
		LET secondIncrement = QUERY ONE "INCR $key" IN redis USING redis_exec WITH {
			key: @counterKey
		}
		LET missing = QUERY ONE "GET missing:$id" IN redis USING redis WITH {
			id: @id
		}
		LET deleted = QUERY ONE "DEL $key" IN redis USING redis_exec WITH {
			key: @counterKey
		}
		LET closed = DB::REDIS::CLOSE(redis)
		RETURN {
			set,
			blockedSet,
			value,
			hset,
			profile,
			mset,
			mget,
			quotedSet,
			quotedValue,
			sadd,
			roleCount,
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
		ferret.WithRuntimeParam("id", runtime.NewString(id)),
		ferret.WithRuntimeParam("listKey", runtime.NewString(listKey)),
		ferret.WithRuntimeParam("counterKey", runtime.NewString(counterKey)),
		ferret.WithRuntimeParam("mgetFirstKey", runtime.NewString(mgetFirstKey)),
		ferret.WithRuntimeParam("mgetSecondKey", runtime.NewString(mgetSecondKey)),
		ferret.WithRuntimeParam("quotedKey", runtime.NewString(quotedKey)),
	)
	if err != nil {
		t.Fatalf("unexpected Redis integration error: %v", err)
	}
	assertIntegrationOutput(t, output)

	_, err = runFQL(t, `
		LET redis = DB::REDIS::OPEN({ url: @redisURL })
		RETURN QUERY ONE "SET session:$id blocked" IN redis USING redis WITH {
			id: @id
		}
	`,
		ferret.WithRuntimeParam("redisURL", runtime.NewString(redisURL)),
		ferret.WithRuntimeParam("id", runtime.NewString(id)),
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
		RETURN QUERY ONE "BLPOP $key $seconds" IN redis USING redis_exec WITH {
			key: @missingKey,
			seconds: 1
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
		MSet            string            `json:"mset"`
		MGet            []string          `json:"mget"`
		QuotedSet       string            `json:"quotedSet"`
		QuotedValue     string            `json:"quotedValue"`
		Items           []string          `json:"items"`
		HSet            int               `json:"hset"`
		SAdd            int               `json:"sadd"`
		RoleCount       int               `json:"roleCount"`
		Pushed          int               `json:"pushed"`
		FirstIncrement  int               `json:"firstIncrement"`
		SecondIncrement int               `json:"secondIncrement"`
		Deleted         int               `json:"deleted"`
		Closed          bool              `json:"closed"`
	}
	if err := json.Unmarshal(output.Content, &actual); err != nil {
		t.Fatalf("failed to decode Redis integration output: %v", err)
	}

	if actual.Set != "OK" || actual.BlockedSet != nil || actual.Value != "hello world" || actual.HSet != 4 ||
		actual.Profile["name"] != "Tim" || actual.Profile["age"] != "42" ||
		actual.Profile["score"] != "1.5" || actual.Profile["active"] != "1" ||
		actual.MSet != "OK" || !equalStrings(actual.MGet, []string{"first", "second"}) ||
		actual.QuotedSet != "OK" || actual.QuotedValue != "hello quoted world" ||
		actual.SAdd != 2 || actual.RoleCount != 2 ||
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

func assertErrorContains(t *testing.T, err error, expected string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error containing %q", expected)
	}
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error containing %q, got %q", expected, err.Error())
	}
}
