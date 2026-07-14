package openai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
)

func TestExecutorMapsTextRequestAndResponse(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "environment-secret")
	t.Setenv("OPENAI_ADMIN_KEY", "environment-admin-secret")
	t.Setenv("OPENAI_BASE_URL", "https://attacker.invalid/v1")
	t.Setenv("OPENAI_CUSTOM_HEADERS", "Authorization: Bearer environment-header-secret")

	fake := newFakeHTTPClient(func(context.Context, *ferrethttp.Request) (*ferrethttp.Response, error) {
		return jsonResponse(http.StatusOK, successResponseBody("hello")), nil
	})
	model := newTestModel(t, "  opaque/model:v1  ")
	temperature := 0.0

	response, err := model.Generate(networkContext(fake), core.Request{
		Messages: []core.Message{
			{Role: core.RoleSystem, Content: core.Content{Type: core.ContentText, Text: "system message"}},
			{Role: core.RoleDeveloper, Content: core.Content{Type: core.ContentText, Text: "developer message"}},
			{Role: core.RoleUser, Content: core.Content{Type: core.ContentText, Text: "user message"}},
			{Role: core.RoleAssistant, Content: core.Content{Type: core.ContentText, Text: "assistant message"}},
		},
		Instructions: "top-level instructions",
		Options: core.ExecutionOptions{
			Temperature:     &temperature,
			MaxOutputTokens: 321,
		},
	})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	requests := fake.Requests()
	if len(requests) != 1 {
		t.Fatalf("expected one provider request, got %d", len(requests))
	}
	request := requests[0]
	if request.Method != http.MethodPost {
		t.Fatalf("expected POST, got %s", request.Method)
	}
	if request.URL != "https://api.openai.com/v1/responses" {
		t.Fatalf("expected fixed Responses endpoint, got %q", request.URL)
	}
	if got := http.Header(request.Headers).Get("Authorization"); got != "Bearer explicit-secret" {
		t.Fatalf("expected explicit API key, got %q", got)
	}

	payload := decodeRequestBody(t, request)
	if payload["model"] != "  opaque/model:v1  " {
		t.Fatalf("expected opaque model pass-through, got %#v", payload["model"])
	}
	if payload["store"] != false {
		t.Fatalf("expected store=false, got %#v", payload["store"])
	}
	if payload["truncation"] != "disabled" {
		t.Fatalf("expected disabled truncation, got %#v", payload["truncation"])
	}
	if payload["instructions"] != "top-level instructions" {
		t.Fatalf("unexpected instructions: %#v", payload["instructions"])
	}
	if payload["temperature"] != float64(0) {
		t.Fatalf("expected explicit zero temperature, got %#v", payload["temperature"])
	}
	if payload["max_output_tokens"] != float64(321) {
		t.Fatalf("unexpected max output tokens: %#v", payload["max_output_tokens"])
	}
	if _, found := payload["previous_response_id"]; found {
		t.Fatal("previous_response_id must not be sent")
	}
	if _, found := payload["conversation"]; found {
		t.Fatal("conversation must not be sent")
	}

	input, ok := payload["input"].([]any)
	if !ok || len(input) != 4 {
		t.Fatalf("expected four input messages, got %#v", payload["input"])
	}
	expectedRoles := []string{"system", "developer", "user", "assistant"}
	for idx, expected := range expectedRoles {
		message, ok := input[idx].(map[string]any)
		if !ok {
			t.Fatalf("message %d has unexpected shape: %#v", idx, input[idx])
		}
		if message["role"] != expected {
			t.Fatalf("message %d: expected role %q, got %#v", idx, expected, message["role"])
		}
		if _, ok := message["content"].(string); !ok {
			t.Fatalf("message %d: expected string content, got %#v", idx, message["content"])
		}
	}

	if response.ID != "resp_123" || response.Model != "resolved-model" || response.Text != "hello" {
		t.Fatalf("unexpected normalized response: %#v", response)
	}
	if response.Usage != (core.Usage{InputTokens: 11, OutputTokens: 7, TotalTokens: 18}) {
		t.Fatalf("unexpected usage: %#v", response.Usage)
	}
	if !json.Valid(response.RawJSON) || !strings.Contains(string(response.RawJSON), `"resp_123"`) {
		t.Fatalf("expected copied raw provider JSON, got %q", response.RawJSON)
	}
}

func TestExecutorMapsStructuredTextFormat(t *testing.T) {
	t.Parallel()

	schema, err := core.NewSchema(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
		"required":             []any{"name"},
		"additionalProperties": false,
	})
	if err != nil {
		t.Fatalf("compile schema: %v", err)
	}
	fake := newFakeHTTPClient(func(context.Context, *ferrethttp.Request) (*ferrethttp.Response, error) {
		return jsonResponse(http.StatusOK, successResponseBody(`{"name":"Ada"}`)), nil
	})
	model := newTestModel(t, "gpt-test")

	response, err := model.GenerateStructured(networkContext(fake), core.StructuredRequest{
		Request: core.Request{
			Messages: []core.Message{
				{Role: core.RoleUser, Content: core.Content{Type: core.ContentText, Text: "extract"}},
			},
		},
		Name:        "extract_result",
		Description: "extracted data",
		Schema:      schema,
	})
	if err != nil {
		t.Fatalf("generate structured: %v", err)
	}
	if response.Text != `{"name":"Ada"}` {
		t.Fatalf("unexpected structured text: %q", response.Text)
	}

	payload := decodeRequestBody(t, fake.Requests()[0])
	textConfig, ok := payload["text"].(map[string]any)
	if !ok {
		t.Fatalf("expected text configuration, got %#v", payload["text"])
	}
	format, ok := textConfig["format"].(map[string]any)
	if !ok {
		t.Fatalf("expected structured format, got %#v", textConfig["format"])
	}
	if format["type"] != "json_schema" || format["name"] != "extract_result" || format["strict"] != true {
		t.Fatalf("unexpected structured format: %#v", format)
	}
	if format["description"] != "extracted data" {
		t.Fatalf("unexpected schema description: %#v", format["description"])
	}
	payloadSchema, ok := format["schema"].(map[string]any)
	if !ok || payloadSchema["additionalProperties"] != false {
		t.Fatalf("unexpected JSON Schema: %#v", format["schema"])
	}
}

func TestExecutorStructuredOutputValidationUsesCoreDispatcher(t *testing.T) {
	t.Parallel()

	schema, err := core.NewSchema(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
		"required":             []string{"name"},
		"additionalProperties": false,
	})
	if err != nil {
		t.Fatalf("compile schema: %v", err)
	}

	tests := []struct {
		name     string
		output   string
		expected core.ErrorCode
	}{
		{name: "malformed JSON", output: `TOP_SECRET {]`, expected: core.ErrInvalidStructuredOutput},
		{name: "schema mismatch", output: `{"name":7,"leak":"TOP_SECRET"}`, expected: core.ErrSchemaValidation},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fake := newFakeHTTPClient(func(context.Context, *ferrethttp.Request) (*ferrethttp.Response, error) {
				return jsonResponse(http.StatusOK, successResponseBody(test.output)), nil
			})
			model := newTestModel(t, "gpt-test")

			_, err := core.Execute(networkContext(fake), model, core.OperationRequest{
				Mode:  core.ModeExtract,
				Input: "Ada",
				Semantic: core.SemanticOptions{
					Schema: schema,
				},
			})
			if code := errorCode(t, err); code != test.expected {
				t.Fatalf("expected %s, got %s", test.expected, code)
			}
			if strings.Contains(err.Error(), "TOP_SECRET") {
				t.Fatalf("error exposed structured provider output: %v", err)
			}
			if len(fake.Requests()) != 1 {
				t.Fatalf("expected one provider request, got %d", len(fake.Requests()))
			}
		})
	}
}

func TestExecutorMapsClassificationSchemaThroughCoreDispatcher(t *testing.T) {
	t.Parallel()

	fake := newFakeHTTPClient(func(context.Context, *ferrethttp.Request) (*ferrethttp.Response, error) {
		return jsonResponse(http.StatusOK, successResponseBody(`{"label":"positive"}`)), nil
	})
	model := newTestModel(t, "gpt-test")

	_, err := core.Execute(networkContext(fake), model, core.OperationRequest{
		Mode:  core.ModeClassify,
		Input: "This is excellent.",
		Semantic: core.SemanticOptions{
			Labels: []string{"positive", "negative"},
		},
	})
	if err != nil {
		t.Fatalf("classify: %v", err)
	}

	payload := decodeRequestBody(t, fake.Requests()[0])
	textConfig, ok := payload["text"].(map[string]any)
	if !ok {
		t.Fatalf("expected text configuration, got %#v", payload["text"])
	}
	format, ok := textConfig["format"].(map[string]any)
	if !ok {
		t.Fatalf("expected structured format, got %#v", textConfig["format"])
	}
	if format["name"] != "classification_result" || format["strict"] != true {
		t.Fatalf("unexpected classification format: %#v", format)
	}
	schema, ok := format["schema"].(map[string]any)
	if !ok || schema["additionalProperties"] != false {
		t.Fatalf("unexpected classification schema: %#v", format["schema"])
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected classification properties: %#v", schema["properties"])
	}
	label, ok := properties["label"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected label schema: %#v", properties["label"])
	}
	enum, ok := label["enum"].([]any)
	if !ok || len(enum) != 2 || enum[0] != "positive" || enum[1] != "negative" {
		t.Fatalf("unexpected classification enum: %#v", label["enum"])
	}
}

func TestExecutorNormalizesHTTPFailuresWithoutRetryOrDetails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		apiCode    string
		apiType    string
		expected   core.ErrorCode
		statusCode int
	}{
		{name: "unauthorized", statusCode: http.StatusUnauthorized, apiCode: "invalid_api_key", expected: core.ErrAuth},
		{name: "forbidden", statusCode: http.StatusForbidden, apiCode: "access_denied", expected: core.ErrAuth},
		{name: "rate limit", statusCode: http.StatusTooManyRequests, apiCode: "rate_limit_exceeded", expected: core.ErrRateLimit},
		{name: "request timeout", statusCode: http.StatusRequestTimeout, apiCode: "timeout", expected: core.ErrTimeout},
		{name: "context limit", statusCode: http.StatusBadRequest, apiCode: "context_length_exceeded", expected: core.ErrContextLimit},
		{name: "context limit type", statusCode: http.StatusBadRequest, apiType: "context_length_exceeded", expected: core.ErrContextLimit},
		{name: "provider", statusCode: http.StatusInternalServerError, apiCode: "server_error", expected: core.ErrProvider},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			apiType := test.apiType
			if apiType == "" {
				apiType = "provider_type"
			}
			body, err := json.Marshal(map[string]any{
				"error": map[string]any{
					"message": "sensitive-provider-detail",
					"type":    apiType,
					"param":   nil,
					"code":    test.apiCode,
				},
			})
			if err != nil {
				t.Fatalf("marshal error response: %v", err)
			}

			fake := newFakeHTTPClient(func(context.Context, *ferrethttp.Request) (*ferrethttp.Response, error) {
				return jsonResponse(test.statusCode, body), nil
			})
			model := newTestModel(t, "gpt-test")

			_, err = model.Generate(networkContext(fake), core.Request{
				Messages: []core.Message{
					{Role: core.RoleUser, Content: core.Content{Type: core.ContentText, Text: "hello"}},
				},
			})
			if code := errorCode(t, err); code != test.expected {
				t.Fatalf("expected %s, got %s", test.expected, code)
			}
			if strings.Contains(err.Error(), "sensitive-provider-detail") || strings.Contains(err.Error(), "explicit-secret") {
				t.Fatalf("error exposed provider or credential details: %v", err)
			}
			if len(fake.Requests()) != 1 {
				t.Fatalf("expected one request without retries, got %d", len(fake.Requests()))
			}
		})
	}
}

func TestExecutorNormalizesRefusalAndIncompleteResponses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		body     string
		expected core.ErrorCode
	}{
		{
			name:     "refusal",
			body:     `{"id":"resp","status":"completed","model":"gpt-test","output":[{"type":"message","content":[{"type":"refusal","refusal":"sensitive refusal text"}]}]}`,
			expected: core.ErrRefusal,
		},
		{
			name:     "content filter",
			body:     `{"id":"resp","status":"incomplete","model":"gpt-test","incomplete_details":{"reason":"content_filter"},"output":[]}`,
			expected: core.ErrRefusal,
		},
		{
			name:     "output limit",
			body:     `{"id":"resp","status":"incomplete","model":"gpt-test","incomplete_details":{"reason":"max_output_tokens"},"output":[]}`,
			expected: core.ErrProvider,
		},
		{
			name:     "failed context limit",
			body:     `{"id":"resp","status":"failed","model":"gpt-test","error":{"code":"context_length_exceeded","message":"sensitive"},"output":[]}`,
			expected: core.ErrContextLimit,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fake := newFakeHTTPClient(func(context.Context, *ferrethttp.Request) (*ferrethttp.Response, error) {
				return jsonResponse(http.StatusOK, []byte(test.body)), nil
			})
			model := newTestModel(t, "gpt-test")
			_, err := model.Generate(networkContext(fake), core.Request{})
			if code := errorCode(t, err); code != test.expected {
				t.Fatalf("expected %s, got %s", test.expected, code)
			}
			if strings.Contains(err.Error(), "sensitive") {
				t.Fatalf("error exposed response details: %v", err)
			}
		})
	}
}

func TestExecutorHonorsRequestTimeout(t *testing.T) {
	t.Parallel()

	fake := newFakeHTTPClient(func(ctx context.Context, _ *ferrethttp.Request) (*ferrethttp.Response, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	})
	model := newTestModel(t, "gpt-test")

	_, err := model.Generate(networkContext(fake), core.Request{
		Options: core.ExecutionOptions{Timeout: time.Millisecond},
	})
	if code := errorCode(t, err); code != core.ErrTimeout {
		t.Fatalf("expected timeout, got %s", code)
	}
}

func TestExecutorPreservesCallerCancellation(t *testing.T) {
	t.Parallel()

	fake := newFakeHTTPClient(func(ctx context.Context, _ *ferrethttp.Request) (*ferrethttp.Response, error) {
		<-ctx.Done()

		return nil, ctx.Err()
	})
	model := newTestModel(t, "gpt-test")
	ctx, cancel := context.WithCancel(networkContext(fake))
	cancel()

	_, err := model.Generate(ctx, core.Request{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestExecutorHonorsFerretNetworkPolicy(t *testing.T) {
	t.Parallel()

	client := ferrethttp.New(ferrethttp.WithBlockedHosts("api.openai.com"))
	model := newTestModel(t, "gpt-test")

	_, err := model.Generate(networkContext(client), core.Request{})
	if code := errorCode(t, err); code != core.ErrProvider {
		t.Fatalf("expected sanitized provider error, got %s", code)
	}
	if strings.Contains(err.Error(), "api.openai.com") {
		t.Fatalf("error exposed transport details: %v", err)
	}
}

func TestExecutorRejectsUnsupportedProviderMessageValues(t *testing.T) {
	t.Parallel()

	model := newTestModel(t, "gpt-test")

	_, err := model.Generate(context.Background(), core.Request{
		Messages: []core.Message{
			{Role: core.Role("tool"), Content: core.Content{Type: core.ContentText, Text: "hello"}},
		},
	})
	if code := errorCode(t, err); code != core.ErrInvalidOptions {
		t.Fatalf("expected invalid options, got %s", code)
	}

	_, err = model.Generate(context.Background(), core.Request{
		Messages: []core.Message{
			{Role: core.RoleUser, Content: core.Content{Type: core.ContentType("image"), Text: "hello"}},
		},
	})
	if code := errorCode(t, err); code != core.ErrInvalidOptions {
		t.Fatalf("expected invalid options, got %s", code)
	}
}

func TestNormalizeErrorTreatsWrappedDeadlineAsTimeout(t *testing.T) {
	t.Parallel()

	err := normalizeError(errors.Join(errors.New("transport wrapper"), context.DeadlineExceeded))
	if code := errorCode(t, err); code != core.ErrTimeout {
		t.Fatalf("expected timeout, got %s", code)
	}
}

func TestNormalizeErrorPreservesWrappedCancellation(t *testing.T) {
	t.Parallel()

	err := normalizeError(errors.Join(errors.New("transport wrapper"), context.Canceled))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
