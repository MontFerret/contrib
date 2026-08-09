package core

import (
	"context"
	"reflect"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestCompileQueryTemplateWithoutBindings(t *testing.T) {
	t.Parallel()

	command, args, err := compileQueryTemplate(context.Background(), "  PING\t", nil)
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}
	if command != "PING" || len(args) != 0 {
		t.Fatalf("unexpected command: %q %#v", command, args)
	}
}

func TestCompileQueryTemplatePreservesStandaloneTypes(t *testing.T) {
	t.Parallel()

	command, args, err := compileQueryTemplate(
		context.Background(),
		"COMMAND $string $integer $float $boolean $binary",
		runtime.NewObjectWith(map[string]runtime.Value{
			"string":  runtime.NewString("Tim Voronov"),
			"integer": runtime.NewInt64(42),
			"float":   runtime.NewFloat(1.5),
			"boolean": runtime.True,
			"binary":  runtime.NewBinary([]byte{0, 1, 2}),
		}),
	)
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}

	want := []any{"Tim Voronov", int64(42), 1.5, true, []byte{0, 1, 2}}
	if command != "COMMAND" || !reflect.DeepEqual(args, want) {
		t.Fatalf("unexpected command: %q %#v, want %#v", command, args, want)
	}
}

func TestCompileQueryTemplateComposesEmbeddedBindings(t *testing.T) {
	t.Parallel()

	command, args, err := compileQueryTemplate(
		context.Background(),
		"GET tenant:$tenant:user:$id:$score:$active:$binary",
		runtime.NewObjectWith(map[string]runtime.Value{
			"tenant": runtime.NewString("acme"),
			"id":     runtime.NewInt(42),
			"score":  runtime.NewFloat(1.5),
			"active": runtime.True,
			"binary": runtime.NewBinary([]byte{0, 'x'}),
		}),
	)
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}

	want := []any{"tenant:acme:user:42:1.5:1:\x00x"}
	if command != "GET" || !reflect.DeepEqual(args, want) {
		t.Fatalf("unexpected command: %q %#v, want %#v", command, args, want)
	}
}

func TestCompileQueryTemplateQuotesAndEscapes(t *testing.T) {
	t.Parallel()

	command, args, err := compileQueryTemplate(
		context.Background(),
		`SET greeting "hello world" 'single value' "" "line\n\x41" "\$name" '\$other'`,
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}

	want := []any{"greeting", "hello world", "single value", "", "line\nA", "$name", "$other"}
	if command != "SET" || !reflect.DeepEqual(args, want) {
		t.Fatalf("unexpected command: %q %#v, want %#v", command, args, want)
	}
}

func TestCompileQueryTemplateResolvesQuotedBindingsAndSpread(t *testing.T) {
	t.Parallel()

	command, args, err := compileQueryTemplate(
		context.Background(),
		`COMMAND "$integer" '$value' "$items..."`,
		runtime.NewObjectWith(map[string]runtime.Value{
			"integer": runtime.NewInt(42),
			"value":   runtime.NewString("hello world"),
			"items": runtime.NewArrayWith(
				runtime.NewString("first"),
				runtime.True,
			),
		}),
	)
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}

	want := []any{int64(42), "hello world", "first", true}
	if command != "COMMAND" || !reflect.DeepEqual(args, want) {
		t.Fatalf("unexpected command: %q %#v, want %#v", command, args, want)
	}
}

func TestCompileQueryTemplateExpandsMultipleSpreadsInOrder(t *testing.T) {
	t.Parallel()

	command, args, err := compileQueryTemplate(
		context.Background(),
		"COMMAND $first... middle $empty... $second... $first...",
		runtime.NewObjectWith(map[string]runtime.Value{
			"first": runtime.NewArrayWith(
				runtime.NewString("a"),
				runtime.NewInt(2),
			),
			"empty": runtime.NewArray(0),
			"second": runtime.NewArrayWith(
				runtime.False,
				runtime.NewBinary([]byte("bytes")),
			),
		}),
	)
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}

	want := []any{"a", int64(2), "middle", false, []byte("bytes"), "a", int64(2)}
	if command != "COMMAND" || !reflect.DeepEqual(args, want) {
		t.Fatalf("unexpected command: %q %#v, want %#v", command, args, want)
	}
}

func TestCompileQueryTemplatePreservesLiteralDollars(t *testing.T) {
	t.Parallel()

	command, args, err := compileQueryTemplate(
		context.Background(),
		`JSON.GET key $ $.path $[0] $9 \$name`,
		runtime.NewObjectWith(map[string]runtime.Value{
			"unused": runtime.NewString("ignored"),
		}),
	)
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}

	want := []any{"key", "$", "$.path", "$[0]", "$9", "$name"}
	if command != "JSON.GET" || !reflect.DeepEqual(args, want) {
		t.Fatalf("unexpected command: %q %#v, want %#v", command, args, want)
	}
}

func TestCompileQueryTemplateAllowsUnusedBindingsAndDoesNotTreatParamsSpecially(t *testing.T) {
	t.Parallel()

	command, args, err := compileQueryTemplate(
		context.Background(),
		"PING",
		runtime.NewObjectWith(map[string]runtime.Value{
			"params": runtime.NewArrayWith(runtime.NewString("not transported")),
			"other":  runtime.NewObject(),
		}),
	)
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}
	if command != "PING" || len(args) != 0 {
		t.Fatalf("unexpected command: %q %#v", command, args)
	}
}

func TestCompileQueryTemplateRejectsInvalidBindings(t *testing.T) {
	t.Parallel()

	cases := []struct {
		bindings runtime.Value
		name     string
		template string
		want     string
	}{
		{
			name:     "missing",
			template: "GET $key",
			want:     `missing Redis query binding "key"`,
		},
		{
			name:     "case sensitive",
			template: "GET $key",
			bindings: runtime.NewObjectWith(map[string]runtime.Value{"KEY": runtime.NewString("value")}),
			want:     `missing Redis query binding "key"`,
		},
		{
			name:     "spread is not list",
			template: "MGET $keys...",
			bindings: runtime.NewObjectWith(map[string]runtime.Value{"keys": runtime.NewString("key")}),
			want:     "must be an array or list",
		},
		{
			name:     "unsupported spread item",
			template: "MGET $keys...",
			bindings: runtime.NewObjectWith(map[string]runtime.Value{
				"keys": runtime.NewArrayWith(runtime.NewString("key"), runtime.None),
			}),
			want: "spread binding $keys...[1]",
		},
		{
			name:     "array without spread",
			template: "SET key $value",
			bindings: runtime.NewObjectWith(map[string]runtime.Value{"value": runtime.NewArray(0)}),
			want:     "unsupported Redis argument type Array",
		},
		{
			name:     "none embedded",
			template: "GET key:$id",
			bindings: runtime.NewObjectWith(map[string]runtime.Value{"id": runtime.None}),
			want:     "NONE is not a supported Redis command argument",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := compileQueryTemplate(context.Background(), tt.template, tt.bindings)
			assertErrorContains(t, err, tt.want)
		})
	}
}

func TestCompileQueryTemplateRejectsMalformedTemplates(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		template string
		want     string
	}{
		{name: "empty", template: " \t\n", want: "must contain a command"},
		{name: "empty command", template: `"" key`, want: "non-empty command"},
		{name: "dynamic command", template: "$command key", want: "command must be static"},
		{name: "spread command", template: "$commands... key", want: "command must be static"},
		{name: "unterminated double quote", template: `SET key "value`, want: "unterminated quoted"},
		{name: "unterminated single quote", template: "SET key 'value", want: "unterminated quoted"},
		{name: "quote inside token", template: `SET pre"value"`, want: "must start with its quote"},
		{name: "content after quote", template: `SET "value"suffix`, want: "must be followed by whitespace"},
		{name: "unsupported double escape", template: `SET key "\q"`, want: "unsupported escape"},
		{name: "unsupported single escape", template: `SET key '\n'`, want: "unsupported escape"},
		{name: "short hex escape", template: `SET key "\x1"`, want: "requires two hexadecimal digits"},
		{name: "embedded spread prefix", template: "MGET prefix:$keys...", want: "must occupy the entire"},
		{name: "embedded spread suffix", template: "MGET $keys...suffix", want: "must occupy the entire"},
		{name: "spread with another binding", template: "MGET $keys...:$other", want: "must occupy the entire"},
		{name: "extra spread dot", template: "MGET $keys....", want: "must occupy the entire"},
	}

	bindings := runtime.NewObjectWith(map[string]runtime.Value{
		"command":  runtime.NewString("GET"),
		"commands": runtime.NewArrayWith(runtime.NewString("GET")),
		"keys":     runtime.NewArrayWith(runtime.NewString("key")),
		"other":    runtime.NewString("other"),
	})

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := compileQueryTemplate(context.Background(), tt.template, bindings)
			assertErrorContains(t, err, tt.want)
		})
	}
}
