package core

import (
	"context"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecodeCore(t *testing.T) {
	ctx := context.Background()

	t.Run("decodes nested tables arrays and inline tables", func(t *testing.T) {
		input := `
title = "Ferret"
tags = ["config", "toml"]
meta = { enabled = true }

[server]
host = "localhost"
port = 8080

[[plugins]]
name = "html"

[[plugins]]
name = "json"
`

		result, err := Decode(ctx, runtime.NewString(input), DefaultDecodeOptions())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		root := mustRuntimeObject(t, result)
		assertRuntimeStringValue(t, mustObjectField(t, ctx, root, "title"), "Ferret")

		tags := mustObjectFieldArray(t, ctx, root, "tags")
		assertArrayLen(t, ctx, tags, 2)
		assertRuntimeStringValue(t, mustArrayAtIndex(t, ctx, tags, 0), "config")
		assertRuntimeStringValue(t, mustArrayAtIndex(t, ctx, tags, 1), "toml")

		meta := mustObjectFieldObject(t, ctx, root, "meta")
		assertRuntimeBoolValue(t, mustObjectField(t, ctx, meta, "enabled"), true)

		server := mustObjectFieldObject(t, ctx, root, "server")
		assertRuntimeStringValue(t, mustObjectField(t, ctx, server, "host"), "localhost")
		assertRuntimeIntValue(t, mustObjectField(t, ctx, server, "port"), 8080)

		plugins := mustObjectFieldArray(t, ctx, root, "plugins")
		assertArrayLen(t, ctx, plugins, 2)
		assertRuntimeStringValue(t, mustObjectField(t, ctx, mustRuntimeObject(t, mustArrayAtIndex(t, ctx, plugins, 0)), "name"), "html")
		assertRuntimeStringValue(t, mustObjectField(t, ctx, mustRuntimeObject(t, mustArrayAtIndex(t, ctx, plugins, 1)), "name"), "json")
	})

	t.Run("decodes datetime values as strings by default", func(t *testing.T) {
		input := `
offset = 1979-05-27T07:32:00Z
local_dt = 1979-05-27T07:32:00
local_date = 1979-05-27
local_time = 07:32:00
`

		result, err := Decode(ctx, runtime.NewString(input), DefaultDecodeOptions())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		root := mustRuntimeObject(t, result)
		assertRuntimeStringValue(t, mustObjectField(t, ctx, root, "offset"), "1979-05-27T07:32:00Z")
		assertRuntimeStringValue(t, mustObjectField(t, ctx, root, "local_dt"), "1979-05-27T07:32:00")
		assertRuntimeStringValue(t, mustObjectField(t, ctx, root, "local_date"), "1979-05-27")
		assertRuntimeStringValue(t, mustObjectField(t, ctx, root, "local_time"), "07:32:00")
	})

	t.Run("decodes datetime values as native when requested", func(t *testing.T) {
		input := `
offset = 1979-05-27T07:32:00Z
local_dt = 1979-05-27T07:32:00
local_date = 1979-05-27
local_time = 07:32:00
`

		opts := DefaultDecodeOptions()
		opts.DateTime = DecodeDateTimeNative

		result, err := Decode(ctx, runtime.NewString(input), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		root := mustRuntimeObject(t, result)
		assertRuntimeDateTimeLocation(t, mustObjectField(t, ctx, root, "offset"), "UTC")
		assertRuntimeDateTimeLocation(t, mustObjectField(t, ctx, root, "local_dt"), localDateTimeLocation)
		assertRuntimeDateTimeLocation(t, mustObjectField(t, ctx, root, "local_date"), localDateLocation)
		assertRuntimeDateTimeLocation(t, mustObjectField(t, ctx, root, "local_time"), localTimeLocation)
	})

	t.Run("rejects malformed toml", func(t *testing.T) {
		_, err := Decode(ctx, runtime.NewString("title = [unterminated"), DefaultDecodeOptions())
		if err == nil {
			t.Fatal("expected malformed TOML error")
		}

		if _, ok := err.(*TOMLError); !ok {
			t.Fatalf("expected *TOMLError, got %T", err)
		}

		if !strings.Contains(err.Error(), "invalid TOML document") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
