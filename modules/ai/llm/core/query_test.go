package core

import (
	"context"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestExecuteQuerySupportsEveryUsingMode(t *testing.T) {
	executor := &fakeExecutor{
		generateFn: func(_ context.Context, request Request) (Response, error) {
			return Response{Text: string(request.Messages[len(request.Messages)-1].Content.Text)}, nil
		},
		generateStructFn: func(_ context.Context, request StructuredRequest) (Response, error) {
			if request.Name == "classification_result" {
				return Response{Text: "{\"label\":\"yes\"}"}, nil
			}

			return Response{Text: "{\"value\":1}"}, nil
		},
	}
	model := testModel(executor)
	ctx := context.Background()

	for _, kind := range []string{"", "generate", "chat", "summarize"} {
		value, err := model.QueryOne(ctx, runtime.Query{
			Kind:       runtime.NewString(kind),
			Expression: runtime.NewString("payload"),
		})
		if err != nil {
			t.Fatalf("USING %q: %v", kind, err)
		}
		if value.String() != "payload" {
			t.Fatalf("USING %q returned %v", kind, value)
		}
	}

	schema := object(map[string]runtime.Value{
		"type": runtime.NewString("object"),
		"properties": object(map[string]runtime.Value{
			"value": object(map[string]runtime.Value{"type": runtime.NewString("integer")}),
		}),
		"required": runtime.NewArrayWith(runtime.NewString("value")),
	})
	value, err := model.QueryOne(ctx, runtime.Query{
		Kind:       runtime.NewString("extract"),
		Expression: runtime.NewString("payload"),
		Params:     object(map[string]runtime.Value{"schema": schema}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if objectValue(t, value, "value").String() != "1" {
		t.Fatalf("unexpected extraction: %v", value)
	}

	value, err = model.QueryOne(ctx, runtime.Query{
		Kind:       runtime.NewString("classify"),
		Expression: runtime.NewString("payload"),
		Params: object(map[string]runtime.Value{
			"labels": runtime.NewArrayWith(runtime.NewString("yes"), runtime.NewString("no")),
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if objectValue(t, value, "label").String() != "yes" {
		t.Fatalf("unexpected classification: %v", value)
	}
}

func TestExecuteQueryRejectsUnsupportedModeAndUnknownKeys(t *testing.T) {
	model := testModel(&fakeExecutor{})
	ctx := context.Background()

	_, err := model.QueryOne(ctx, runtime.Query{Kind: runtime.NewString("embed")})
	requireCode(t, err, ErrUnsupportedOperation)
	_, err = model.QueryOne(ctx, runtime.Query{
		Kind:   runtime.NewString("generate"),
		Params: object(map[string]runtime.Value{"unknown": runtime.True}),
	})
	requireCode(t, err, ErrInvalidOptions)
	_, err = model.QueryOne(ctx, runtime.Query{
		Kind:    runtime.NewString("generate"),
		Options: object(map[string]runtime.Value{"unknown": runtime.True}),
	})
	requireCode(t, err, ErrInvalidOptions)
	_, err = model.QueryOne(ctx, runtime.Query{Kind: runtime.NewString("extract")})
	requireCode(t, err, ErrInvalidSchema)
	_, err = model.QueryOne(ctx, runtime.Query{Kind: runtime.NewString("classify")})
	requireCode(t, err, ErrInvalidOptions)
}

func TestExecuteQuerySeparatesWithFromOptions(t *testing.T) {
	executor := &fakeExecutor{}
	model := testModel(executor)
	ctx := context.Background()

	_, err := model.QueryOne(ctx, runtime.Query{
		Kind:       runtime.NewString("summarize"),
		Expression: runtime.NewString("payload"),
		Params: object(map[string]runtime.Value{
			"style":        runtime.NewString("concise"),
			"instructions": runtime.NewString("Prefer active voice."),
		}),
		Options: object(map[string]runtime.Value{
			"temperature":     runtime.ZeroFloat,
			"maxOutputTokens": runtime.NewInt(80),
			"timeout":         runtime.ZeroInt,
		}),
	})
	if err != nil {
		t.Fatal(err)
	}

	requests := executor.Requests()
	if len(requests) != 1 {
		t.Fatalf("expected one provider request, got %d", len(requests))
	}
	request := requests[0]
	if request.Options.Temperature == nil || *request.Options.Temperature != 0 || request.Options.MaxOutputTokens != 80 {
		t.Fatalf("query OPTIONS were not preserved: %#v", request.Options)
	}
	if !strings.Contains(request.Instructions, "concise") || !strings.Contains(request.Instructions, "Prefer active voice.") {
		t.Fatalf("query WITH data was not preserved: %q", request.Instructions)
	}

	_, err = model.QueryOne(ctx, runtime.Query{
		Kind:       runtime.NewString("generate"),
		Expression: runtime.NewString("payload"),
		Params:     object(map[string]runtime.Value{"temperature": runtime.ZeroFloat}),
	})
	requireCode(t, err, ErrInvalidOptions)

	_, err = model.QueryOne(ctx, runtime.Query{
		Kind:       runtime.NewString("summarize"),
		Expression: runtime.NewString("payload"),
		Options:    object(map[string]runtime.Value{"style": runtime.NewString("concise")}),
	})
	requireCode(t, err, ErrInvalidOptions)
}
