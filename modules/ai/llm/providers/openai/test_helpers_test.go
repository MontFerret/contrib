package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
	ferretnet "github.com/MontFerret/ferret/v2/pkg/net"
	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
)

func networkContext(client ferrethttp.Client) context.Context {
	return ferretnet.WithNetwork(
		context.Background(),
		ferretnet.New(ferretnet.WithHTTPClient(client)),
	)
}

func newTestModel(t *testing.T, modelName string) core.Model {
	t.Helper()

	model, err := NewFactory().NewModel(context.Background(), core.ModelOptions{
		Model:  modelName,
		APIKey: "explicit-secret",
	})
	if err != nil {
		t.Fatalf("create model: %v", err)
	}

	return model
}

func jsonResponse(statusCode int, body []byte) *ferrethttp.Response {
	return &ferrethttp.Response{
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		StatusCode: statusCode,
		Headers: ferrethttp.Headers{
			"Content-Type": []string{"application/json"},
		},
		Body: body,
	}
}

func successResponseBody(text string) []byte {
	body, err := json.Marshal(map[string]any{
		"id":     "resp_123",
		"object": "response",
		"status": "completed",
		"model":  "resolved-model",
		"output": []any{
			map[string]any{
				"id":     "msg_123",
				"type":   "message",
				"role":   "assistant",
				"status": "completed",
				"content": []any{
					map[string]any{
						"type":        "output_text",
						"text":        text,
						"annotations": []any{},
					},
				},
			},
		},
		"usage": map[string]any{
			"input_tokens":  11,
			"output_tokens": 7,
			"total_tokens":  18,
			"input_tokens_details": map[string]any{
				"cached_tokens": 0,
			},
			"output_tokens_details": map[string]any{
				"reasoning_tokens": 0,
			},
		},
	})
	if err != nil {
		panic(err)
	}

	return body
}

func decodeRequestBody(t *testing.T, request *ferrethttp.Request) map[string]any {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal(request.Body, &payload); err != nil {
		t.Fatalf("decode request body: %v", err)
	}

	return payload
}

func errorCode(t *testing.T, err error) core.ErrorCode {
	t.Helper()

	code, ok := core.CodeOf(err)
	if !ok {
		t.Fatalf("expected typed AI::LLM error, got %T: %v", err, err)
	}

	return code
}
