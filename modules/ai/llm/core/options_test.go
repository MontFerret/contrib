package core

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecodeModelOptionsRequiresExplicitCredentialsAndPreservesModel(t *testing.T) {
	ctx := context.Background()
	options, err := DecodeModelOptions(ctx, object(map[string]runtime.Value{
		"model":   runtime.NewString("  opaque/model  "),
		"apiKey":  runtime.NewString("explicit"),
		"session": runtime.True,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if options.Model != "  opaque/model  " || !options.Session {
		t.Fatalf("unexpected model options: %#v", options)
	}

	_, err = DecodeModelOptions(ctx, object(map[string]runtime.Value{"model": runtime.NewString("m")}))
	requireCode(t, err, ErrInvalidOptions)
	_, err = DecodeModelOptions(ctx, object(map[string]runtime.Value{
		"model": runtime.NewString("m"), "apiKey": runtime.NewString("k"), "baseURL": runtime.NewString("x"),
	}))
	requireCode(t, err, ErrInvalidOptions)
}

func TestDecodeExecutionOptionsKeepsExplicitZeroAndChecksBounds(t *testing.T) {
	ctx := context.Background()
	options, err := DecodeExecutionOptions(ctx, object(map[string]runtime.Value{
		"temperature":     runtime.ZeroFloat,
		"maxOutputTokens": runtime.NewInt(20),
		"timeout":         runtime.NewInt(15),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if options.Temperature == nil || *options.Temperature != 0 || options.MaxOutputTokens != 20 || options.Timeout != 15*time.Millisecond {
		t.Fatalf("unexpected execution options: %#v", options)
	}

	invalid := []runtime.Value{
		object(map[string]runtime.Value{"temperature": runtime.NewFloat(2.1)}),
		object(map[string]runtime.Value{"maxOutputTokens": runtime.ZeroInt}),
		object(map[string]runtime.Value{"timeout": runtime.NewInt(-1)}),
		object(map[string]runtime.Value{"timeout": runtime.NewInt64(math.MaxInt64/int64(time.Millisecond) + 1)}),
		object(map[string]runtime.Value{"unknown": runtime.True}),
	}
	for _, value := range invalid {
		_, err := DecodeExecutionOptions(ctx, value)
		requireCode(t, err, ErrInvalidOptions)
	}
}

func TestDecodeSessionOptionsValidatesV1Policy(t *testing.T) {
	ctx := context.Background()
	options, err := DecodeSessionOptions(ctx, object(map[string]runtime.Value{
		"instructions": runtime.NewString("persistent"),
		"context": object(map[string]runtime.Value{
			"mode": runtime.NewString("local"), "overflow": runtime.NewString("error"),
			"maxTokens": runtime.NewInt(100), "reserveOutputTokens": runtime.NewInt(10),
		}),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if options.Context.MaxTokens != 100 || options.Context.ReserveOutputTokens != 10 {
		t.Fatalf("unexpected context options: %#v", options.Context)
	}

	invalidContexts := []*runtime.Object{
		object(map[string]runtime.Value{"mode": runtime.NewString("remote")}),
		object(map[string]runtime.Value{"overflow": runtime.NewString("truncate")}),
		object(map[string]runtime.Value{"maxTokens": runtime.ZeroInt}),
		object(map[string]runtime.Value{"reserveOutputTokens": runtime.NewInt(-1)}),
		object(map[string]runtime.Value{"maxTokens": runtime.NewInt(10), "reserveOutputTokens": runtime.NewInt(10)}),
		object(map[string]runtime.Value{"unknown": runtime.True}),
	}
	for _, value := range invalidContexts {
		_, err := DecodeSessionOptions(ctx, object(map[string]runtime.Value{"context": value}))
		requireCode(t, err, ErrInvalidOptions)
	}
	_, err = DecodeSessionOptions(ctx, object(map[string]runtime.Value{"unknown": runtime.True}))
	requireCode(t, err, ErrInvalidOptions)
}

func TestDecodeMessagesAndLabelsAreStrict(t *testing.T) {
	ctx := context.Background()
	messages, err := DecodeMessages(ctx, runtime.NewArrayWith(object(map[string]runtime.Value{
		"role": runtime.NewString("developer"), "content": runtime.NewString("text"),
	})))
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 || messages[0].Role != RoleDeveloper || messages[0].Content.Text != "text" {
		t.Fatalf("unexpected messages: %#v", messages)
	}

	_, err = DecodeMessages(ctx, runtime.NewArrayWith(object(map[string]runtime.Value{
		"role": runtime.NewString("tool"), "content": runtime.NewString("x"),
	})))
	requireCode(t, err, ErrInvalidOptions)
	_, err = DecodeMessages(ctx, runtime.NewArrayWith(object(map[string]runtime.Value{
		"role": runtime.NewString("user"), "content": runtime.NewString("x"), "name": runtime.NewString("x"),
	})))
	requireCode(t, err, ErrInvalidOptions)

	labels, err := DecodeLabels(ctx, runtime.NewArrayWith(runtime.NewString("a"), runtime.NewString("b")))
	if err != nil || len(labels) != 2 {
		t.Fatalf("unexpected labels: %#v, %v", labels, err)
	}
	_, err = DecodeLabels(ctx, runtime.NewArrayWith(runtime.NewString("a"), runtime.NewString("a")))
	requireCode(t, err, ErrInvalidOptions)
}

func TestFunctionOptionsRejectPositionalSchemaAndLabels(t *testing.T) {
	ctx := context.Background()
	_, _, err := DecodeOperationOptions(ctx, ModeExtract, object(map[string]runtime.Value{"schema": object(map[string]runtime.Value{})}))
	requireCode(t, err, ErrInvalidOptions)
	_, _, err = DecodeOperationOptions(ctx, ModeClassify, object(map[string]runtime.Value{
		"labels": runtime.NewArrayWith(runtime.NewString("a")),
	}))
	requireCode(t, err, ErrInvalidOptions)
}
