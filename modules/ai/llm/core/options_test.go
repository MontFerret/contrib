package core

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
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

func TestDecodeOptionsUseSDKFieldMatching(t *testing.T) {
	ctx := context.Background()

	model, err := DecodeModelOptions(ctx, object(map[string]runtime.Value{
		"MODEL":   runtime.NewString("opaque/model"),
		"APIKEY":  runtime.NewString("explicit"),
		"SESSION": runtime.True,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if model.Model != "opaque/model" || model.APIKey != "explicit" || !model.Session {
		t.Fatalf("unexpected model options: %#v", model)
	}

	session, err := DecodeSessionOptions(ctx, object(map[string]runtime.Value{
		"INSTRUCTIONS": runtime.NewString("persistent"),
		"CONTEXT": object(map[string]runtime.Value{
			"MODE":                runtime.NewString("local"),
			"OVERFLOW":            runtime.NewString("error"),
			"MAXTOKENS":           runtime.NewInt(100),
			"RESERVEOUTPUTTOKENS": runtime.NewInt(10),
		}),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if session.Instructions != "persistent" ||
		session.Context.MaxTokens != 100 ||
		session.Context.ReserveOutputTokens != 10 {
		t.Fatalf("unexpected session options: %#v", session)
	}

	execution, err := DecodeExecutionOptions(ctx, object(map[string]runtime.Value{
		"TEMPERATURE":     runtime.ZeroFloat,
		"MAXOUTPUTTOKENS": runtime.NewInt(20),
		"TIMEOUT":         runtime.NewInt(15),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if execution.Temperature == nil ||
		*execution.Temperature != 0 ||
		execution.MaxOutputTokens != 20 ||
		execution.Timeout != 15*time.Millisecond {
		t.Fatalf("unexpected execution options: %#v", execution)
	}

	messages, err := DecodeMessages(ctx, runtime.NewArrayWith(object(map[string]runtime.Value{
		"ROLE":    runtime.NewString("user"),
		"CONTENT": runtime.NewString("hello"),
	})))
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 || messages[0].Role != RoleUser || messages[0].Content.Text != "hello" {
		t.Fatalf("unexpected messages: %#v", messages)
	}
}

func TestDecodeOptionsPreserveNoneSemantics(t *testing.T) {
	ctx := context.Background()

	invalid := []func() error{
		func() error {
			_, err := DecodeModelOptions(ctx, object(map[string]runtime.Value{
				"model": runtime.NewString("m"), "apiKey": runtime.NewString("k"), "session": runtime.None,
			}))
			return err
		},
		func() error {
			_, err := DecodeExecutionOptions(ctx, object(map[string]runtime.Value{"temperature": runtime.None}))
			return err
		},
		func() error {
			_, err := DecodeSessionOptions(ctx, object(map[string]runtime.Value{"instructions": runtime.None}))
			return err
		},
		func() error {
			_, err := DecodeSemanticOptions(ctx, ModeGenerate, object(map[string]runtime.Value{
				"instructions": runtime.None,
			}))
			return err
		},
		func() error {
			_, err := DecodeSemanticOptions(ctx, ModeClassify, object(map[string]runtime.Value{
				"labels": runtime.None,
			}))
			return err
		},
	}

	for _, run := range invalid {
		requireCode(t, run(), ErrInvalidOptions)
	}

	_, err := DecodeSemanticOptions(ctx, ModeExtract, object(map[string]runtime.Value{
		"schema": runtime.None,
	}))
	requireCode(t, err, ErrInvalidSchema)

	session, err := DecodeSessionOptions(ctx, object(map[string]runtime.Value{"context": runtime.None}))
	if err != nil {
		t.Fatal(err)
	}
	if session.Context.Mode != "local" || session.Context.Overflow != "error" {
		t.Fatalf("unexpected default session context: %#v", session.Context)
	}

	chat, err := DecodeSemanticOptions(ctx, ModeChat, object(map[string]runtime.Value{"messages": runtime.None}))
	if err != nil {
		t.Fatal(err)
	}
	if chat.Messages != nil {
		t.Fatalf("expected no chat messages, got %#v", chat.Messages)
	}
}

func TestSDKDecodeTypeErrorsUseStableModuleCode(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		run  func() error
		name string
	}{
		{
			name: "model",
			run: func() error {
				_, err := DecodeModelOptions(ctx, object(map[string]runtime.Value{
					"model": runtime.NewInt(1), "apiKey": runtime.NewString("k"),
				}))
				return err
			},
		},
		{
			name: "execution",
			run: func() error {
				_, err := DecodeExecutionOptions(ctx, object(map[string]runtime.Value{
					"temperature": runtime.NewString("cold"),
				}))
				return err
			},
		},
		{
			name: "session",
			run: func() error {
				_, err := DecodeSessionOptions(ctx, object(map[string]runtime.Value{
					"context": runtime.NewString("local"),
				}))
				return err
			},
		},
		{
			name: "messages",
			run: func() error {
				_, err := DecodeMessages(ctx, runtime.NewArrayWith(object(map[string]runtime.Value{
					"role": runtime.NewString("user"), "content": runtime.NewInt(1),
				})))
				return err
			},
		},
		{
			name: "labels",
			run: func() error {
				_, err := DecodeLabels(ctx, runtime.NewArrayWith(runtime.NewInt(1)))
				return err
			},
		},
		{
			name: "operation",
			run: func() error {
				_, _, err := DecodeOperationOptions(ctx, ModeSummarize, object(map[string]runtime.Value{
					"style": runtime.True,
				}))
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.run()
			requireCode(t, err, ErrInvalidOptions)
		})
	}
}

func TestSDKDecodeErrorsExposeOnlySafeConversionDetails(t *testing.T) {
	_, err := DecodeExecutionOptions(context.Background(), object(map[string]runtime.Value{
		"temperature": runtime.NewString("sensitive-option-value"),
	}))
	requireCode(t, err, ErrInvalidOptions)

	if !strings.Contains(err.Error(), "temperature") {
		t.Fatalf("expected field path in conversion error, got %v", err)
	}
	if strings.Contains(err.Error(), "sensitive-option-value") {
		t.Fatalf("conversion error exposed the option value: %v", err)
	}
}

func TestSDKDecodeErrorsHideUnsafeSourceDetails(t *testing.T) {
	const privateDetail = "private iterator detail"

	iterator := sdk.NewSliceIteratorWithEncoding(
		[]int{1},
		sdk.NewCodec[int](func(context.Context, int) (runtime.Value, error) {
			return runtime.None, errors.New(privateDetail)
		}, nil),
	)
	source := sdk.NewIteratorValue(iterator)

	_, err := decodeValue[[]string](
		context.Background(),
		source,
		"labels",
		runtime.TypeIterable,
	)
	requireCode(t, err, ErrInvalidOptions)

	if strings.Contains(err.Error(), privateDetail) {
		t.Fatalf("source error exposed private detail: %v", err)
	}
	if !strings.Contains(err.Error(), "labels are invalid") {
		t.Fatalf("expected sanitized options error, got %v", err)
	}
}

func TestDecodeOptionsPreserveCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := DecodeModelOptions(ctx, object(map[string]runtime.Value{
		"model":  runtime.NewString("m"),
		"apiKey": runtime.NewString("k"),
	}))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}

	_, err = DecodeMessages(ctx, runtime.NewArrayWith(object(map[string]runtime.Value{
		"role": runtime.NewString("user"), "content": runtime.NewString("text"),
	})))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
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

func TestModeSpecificOptionsRemainStrictWithSDKDecoding(t *testing.T) {
	ctx := context.Background()

	_, err := DecodeSemanticOptions(ctx, ModeGenerate, object(map[string]runtime.Value{
		"STYLE": runtime.NewString("concise"),
	}))
	requireCode(t, err, ErrInvalidOptions)

	_, _, err = DecodeOperationOptions(ctx, ModeExtract, object(map[string]runtime.Value{
		"LABELS": runtime.NewArrayWith(runtime.NewString("one")),
	}))
	requireCode(t, err, ErrInvalidOptions)

	semantic, execution, err := DecodeOperationOptions(ctx, ModeSummarize, object(map[string]runtime.Value{
		"STYLE":           runtime.NewString("concise"),
		"MAXWORDS":        runtime.NewInt(25),
		"INSTRUCTIONS":    runtime.NewString("Prefer active voice."),
		"TEMPERATURE":     runtime.ZeroFloat,
		"MAXOUTPUTTOKENS": runtime.NewInt(80),
		"TIMEOUT":         runtime.ZeroInt,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if semantic.Style != "concise" ||
		semantic.MaxWords != 25 ||
		semantic.Instructions != "Prefer active voice." {
		t.Fatalf("unexpected semantic options: %#v", semantic)
	}
	if execution.Temperature == nil ||
		*execution.Temperature != 0 ||
		execution.MaxOutputTokens != 80 ||
		execution.Timeout != 0 {
		t.Fatalf("unexpected execution options: %#v", execution)
	}
}

func TestUnsupportedModePreservesOptionValidationPrecedence(t *testing.T) {
	unsupported := Mode("unsupported")

	_, err := DecodeSemanticOptions(context.Background(), unsupported, runtime.NewString("not options"))
	requireCode(t, err, ErrInvalidOptions)

	_, _, err = DecodeOperationOptions(context.Background(), unsupported, runtime.NewString("not options"))
	requireCode(t, err, ErrInvalidOptions)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = DecodeSemanticOptions(ctx, unsupported, runtime.NewObject())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}

	_, _, err = DecodeOperationOptions(ctx, unsupported, runtime.NewObject())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}

	_, err = DecodeSemanticOptions(context.Background(), unsupported, runtime.NewObject())
	requireCode(t, err, ErrUnsupportedOperation)

	_, _, err = DecodeOperationOptions(context.Background(), unsupported, runtime.NewObject())
	requireCode(t, err, ErrUnsupportedOperation)
}
