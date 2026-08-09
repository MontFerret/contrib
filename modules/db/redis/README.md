# DB::REDIS

The Redis module provides reusable Redis connection handles under the `DB::REDIS` namespace. It exposes Redis commands through Ferret's generic `QUERY` abstraction instead of maintaining one Ferret function per command.

The command name is the query expression and `WITH.params` is the ordered Redis argument list. This keeps Redis command grammar—including positional keys, values, subcommands, and modifiers such as `EX`, `NX`, and `XX`—as the source of truth.

## Opening and Closing Connections

Open a Redis connection with a standard `redis://` or TLS-enabled `rediss://` URL:

```fql
LET redis = DB::REDIS::OPEN({
    url: "redis://localhost:6379/0"
})

RETURN DB::REDIS::CLOSE(redis)
```

The URL can contain Redis username/password authentication and a database number:

```text
redis://user:password@localhost:6379/3
```

go-redis uses RESP3 by default. Select a protocol explicitly with `?protocol=2` or `?protocol=3` when a server or response contract requires it.

`unix://` URLs are intentionally unsupported because runtime access to user-controlled filesystem paths must go through Ferret's filesystem policy.

Connections are opaque Ferret resources. Explicit `CLOSE` is idempotent, and Ferret also closes live resources when their execution ownership ends.

## Reading Data

Use the `redis` dialect for commands that the connected server marks as read-only:

```fql
LET value = QUERY ONE "GET" IN redis USING redis WITH {
    params: ["user:42"]
}
```

The module loads the server's `COMMAND` metadata on the first `redis` query and caches it. A command is executed through this dialect only when Redis marks it `readonly`. Unknown commands, commands without that flag, and metadata lookup failures are rejected before command execution.

The Redis account therefore needs permission to execute `COMMAND` for the read dialect. `redis_exec` remains usable when a Redis-compatible server does not expose command metadata.

Hash reads use the same generic form:

```fql
LET profile = QUERY ONE "HGETALL" IN redis USING redis WITH {
    params: ["user:42"]
}
```

With RESP3, `HGETALL` is naturally decoded as a Ferret object. RESP2 returns the server's flat array representation instead.

Sorted-set reads remain positional:

```fql
LET leaders = QUERY "ZRANGE" IN redis USING redis WITH {
    params: ["leaderboard", 0, 9, "WITHSCORES"]
}
```

## Mutating Data

Use `redis_exec` for commands that mutate Redis. It permits arbitrary one-shot Redis commands, including read commands, and does not require command metadata.

Set a value with Redis modifiers:

```fql
LET result = QUERY ONE "SET" IN redis USING redis_exec WITH {
    params: [
        "session:123",
        value,
        "EX", 300,
        "NX"
    ]
}
```

Write mixed scalar values to a hash:

```fql
LET added = QUERY ONE "HSET" IN redis USING redis_exec WITH {
    params: [
        "user:42",
        "name", "Tim",
        "age", 42,
        "score", 1.5,
        "active", true
    ]
}
```

Increment and delete keys without command-specific wrappers:

```fql
LET count = QUERY ONE "INCR" IN redis USING redis_exec WITH {
    params: ["visits:user:42"]
}

LET deleted = QUERY ONE "DEL" IN redis USING redis_exec WITH {
    params: ["session:123", "visits:user:42"]
}
```

List writes follow the same contract:

```fql
LET length = QUERY ONE "RPUSH" IN redis USING redis_exec WITH {
    params: ["jobs", "first", "second", "third"]
}
```

## Query Contract

The v1 query shape is deliberately small:

```fql
QUERY "<command>" IN redis USING redis WITH {
    params: [<ordered command arguments>]
} OPTIONS {
    timeout: "2s"
}
```

- The command must be one non-empty token, such as `GET`, `HGETALL`, or `JSON.GET`.
- Put Redis subcommands in `params`; for example, use command `CLIENT` with `params: ["LIST"]`, not command `"CLIENT LIST"`.
- `WITH` may be omitted for commands with no arguments. When present, it must be an object containing only `params`, and `params` must be an array.
- Argument order is preserved exactly. Command modifiers are normal `params` elements and are never promoted to named fields.
- String, integer, floating-point, Boolean, and Binary Ferret values are supported as arguments.
- `NONE`, arrays, objects, and other complex Ferret values are rejected rather than serialized implicitly.
- `OPTIONS.timeout` accepts a duration string or Ferret duration, or a numeric millisecond value. Zero adds no query deadline; negative values and unknown options are rejected.

Client-level connection, socket, pool, retry, database, and protocol options supported by go-redis can be supplied through the Redis URL. Per-query timeout remains in Ferret `OPTIONS`, not `WITH`.

## Result Values

Redis responses are converted centrally:

| Redis response | Ferret value |
| --- | --- |
| nil | `NONE` |
| bulk/status string | String |
| integer | Int |
| double | Float |
| Boolean | Boolean |
| array or set | Array |
| RESP3 map with string keys | Object |

Nested arrays and maps are converted recursively. Redis errors—including errors nested in aggregate RESP values—surface as Ferret errors.

Ferret `QUERY` must return a result list. A top-level Redis array becomes that list directly; a scalar, object, or `NONE` response becomes a one-item list. Use `QUERY ONE` for scalar or object command responses such as `GET`, `SET`, `HGETALL` under RESP3, `INCR`, and `DEL`.

## Stateful Redis Features

The v1 module sends one command at a time through a pooled go-redis client. It does not provide pipeline, transaction, Pub/Sub, or long-lived stream-consumer abstractions, and connection-scoped command sequences have no stable same-connection guarantee.

Ordinary Stream and blocking commands that work as a single request/response operation can use the generic executor. Give blocking operations an explicit `OPTIONS.timeout`.

## Testing

Unit tests do not require Redis. Live integration tests run only when `REDIS_URL` is set:

```sh
REDIS_URL='redis://localhost:6379/0?protocol=3' \
GOWORK=off go test ./...
```

GitHub Actions runs the same live test path against a Redis service container.
