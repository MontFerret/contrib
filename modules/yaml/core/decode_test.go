package core

import (
	"context"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecodeCore(t *testing.T) {
	ctx := context.Background()

	t.Run("decodes nested object with common scalar types", func(t *testing.T) {
		input := `
name: Alice
age: 30
height: 1.68
active: true
nickname: null
tags:
  - admin
  - ops
profile:
  city: Boston
`

		result, err := Decode(ctx, runtime.NewString(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		obj := mustRuntimeObject(t, result)
		assertObjectFieldString(t, ctx, obj, "name", "Alice")
		assertObjectFieldInt(t, ctx, obj, "age", 30)
		assertObjectFieldFloat(t, ctx, obj, "height", 1.68)
		assertObjectFieldBool(t, ctx, obj, "active", true)
		assertObjectFieldNone(t, ctx, obj, "nickname")

		tags := mustObjectFieldArray(t, ctx, obj, "tags")
		assertArrayLen(t, ctx, tags, 2)
		assertRuntimeStringValue(t, mustArrayAtIndex(t, ctx, tags, 0), "admin")
		assertRuntimeStringValue(t, mustArrayAtIndex(t, ctx, tags, 1), "ops")

		profile := mustObjectFieldObject(t, ctx, obj, "profile")
		assertObjectFieldString(t, ctx, profile, "city", "Boston")
	})

	t.Run("decodes scalar document", func(t *testing.T) {
		result, err := Decode(ctx, runtime.NewString("hello"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assertRuntimeStringValue(t, result, "hello")
	})

	t.Run("decodes sequence document", func(t *testing.T) {
		result, err := Decode(ctx, runtime.NewString("- one\n- 2\n- false\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustRuntimeArray(t, result)
		assertArrayLen(t, ctx, arr, 3)
		assertRuntimeStringValue(t, mustArrayAtIndex(t, ctx, arr, 0), "one")
		assertRuntimeIntValue(t, mustArrayAtIndex(t, ctx, arr, 1), 2)
		assertRuntimeBoolValue(t, mustArrayAtIndex(t, ctx, arr, 2), false)
	})

	t.Run("resolves anchors aliases and merge keys into plain values", func(t *testing.T) {
		input := `
defaults: &defaults
  role: admin
  active: true
shared: &shared
  nested:
    enabled: true
user:
  <<: [*defaults, *shared]
  name: Alice
`

		result, err := Decode(ctx, runtime.NewString(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		root := mustRuntimeObject(t, result)
		user := mustObjectFieldObject(t, ctx, root, "user")
		assertObjectFieldString(t, ctx, user, "role", "admin")
		assertObjectFieldBool(t, ctx, user, "active", true)
		assertObjectFieldString(t, ctx, user, "name", "Alice")

		nested := mustObjectFieldObject(t, ctx, user, "nested")
		assertObjectFieldBool(t, ctx, nested, "enabled", true)
	})

	t.Run("rejects multiple documents", func(t *testing.T) {
		_, err := Decode(ctx, runtime.NewString("---\na: 1\n---\nb: 2\n"))
		if err == nil {
			t.Fatal("expected multiple document error")
		}

		if !strings.Contains(err.Error(), "multiple YAML documents provided to YAML::DECODE") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("rejects empty input", func(t *testing.T) {
		_, err := Decode(ctx, runtime.NewString(" \n"))
		if err == nil {
			t.Fatal("expected empty input error")
		}

		if !strings.Contains(err.Error(), "empty YAML input") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("rejects malformed yaml", func(t *testing.T) {
		_, err := Decode(ctx, runtime.NewString("name: [unterminated"))
		if err == nil {
			t.Fatal("expected malformed YAML error")
		}

		if _, ok := err.(*YAMLError); !ok {
			t.Fatalf("expected *YAMLError, got %T", err)
		}

		if !strings.Contains(err.Error(), "invalid YAML document") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestDecodeAllCore(t *testing.T) {
	ctx := context.Background()

	t.Run("decodes all documents in order", func(t *testing.T) {
		input := "---\nname: Alice\n---\n- 1\n- 2\n---\ntrue\n"

		result, err := DecodeAll(ctx, runtime.NewString(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustRuntimeArray(t, result)
		assertArrayLen(t, ctx, arr, 3)

		first := mustRuntimeObject(t, mustArrayAtIndex(t, ctx, arr, 0))
		assertObjectFieldString(t, ctx, first, "name", "Alice")

		second := mustRuntimeArray(t, mustArrayAtIndex(t, ctx, arr, 1))
		assertArrayLen(t, ctx, second, 2)
		assertRuntimeIntValue(t, mustArrayAtIndex(t, ctx, second, 0), 1)
		assertRuntimeIntValue(t, mustArrayAtIndex(t, ctx, second, 1), 2)

		assertRuntimeBoolValue(t, mustArrayAtIndex(t, ctx, arr, 2), true)
	})

	t.Run("rejects empty input", func(t *testing.T) {
		_, err := DecodeAll(ctx, runtime.NewString(""))
		if err == nil {
			t.Fatal("expected empty input error")
		}
	})
}
