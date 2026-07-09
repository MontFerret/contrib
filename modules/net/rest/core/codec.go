package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func encodeRequestBody(ctx context.Context, value runtime.Value, encoding Encoding) ([]byte, string, error) {
	if runtime.TypeNone.Is(value) {
		return nil, "", nil
	}

	switch encoding {
	case EncodingJSON:
		converted, err := runtimeToNative(ctx, value)
		if err != nil {
			return nil, "", err
		}

		data, err := json.Marshal(converted)
		if err != nil {
			return nil, "", err
		}

		return data, "application/json", nil
	case EncodingText:
		return []byte(value.String()), "text/plain; charset=utf-8", nil
	case EncodingBytes:
		if binary, ok := value.(runtime.Binary); ok {
			return []byte(binary), "application/octet-stream", nil
		}

		return []byte(value.String()), "application/octet-stream", nil
	case EncodingForm:
		values := make(url.Values)
		if err := appendURLValues(ctx, values, "HTTP request body", value); err != nil {
			return nil, "", err
		}

		return []byte(values.Encode()), "application/x-www-form-urlencoded", nil
	default:
		return nil, "", fmt.Errorf("unsupported request encoding %q", encoding)
	}
}

func decodeResponseBody(body []byte, encoding Encoding) (runtime.Value, error) {
	switch encoding {
	case EncodingJSON:
		return decodeJSONBody(body)
	case EncodingText:
		return runtime.NewString(string(body)), nil
	case EncodingBytes:
		out := make([]byte, len(body))
		copy(out, body)

		return runtime.NewBinary(out), nil
	case EncodingForm:
		return decodeFormBody(body)
	default:
		return runtime.None, fmt.Errorf("unsupported response encoding %q", encoding)
	}
}

func decodeJSONBody(body []byte) (runtime.Value, error) {
	if len(bytes.TrimSpace(body)) == 0 {
		return runtime.None, nil
	}

	var decoded any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&decoded); err != nil {
		return runtime.None, err
	}

	normalized, err := normalizeJSONValue(decoded)
	if err != nil {
		return runtime.None, err
	}

	return runtime.ValueOf(normalized)
}

func decodeFormBody(body []byte) (runtime.Value, error) {
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return runtime.None, err
	}

	out := make(map[string]any, len(values))
	for key, items := range values {
		if len(items) == 1 {
			out[key] = items[0]
		} else {
			entries := make([]any, 0, len(items))
			for _, item := range items {
				entries = append(entries, item)
			}

			out[key] = entries
		}
	}

	return runtime.ValueOf(out)
}

func normalizeJSONValue(input any) (any, error) {
	switch value := input.(type) {
	case json.Number:
		if integer, err := value.Int64(); err == nil {
			return integer, nil
		}

		number, err := value.Float64()
		if err != nil {
			return nil, err
		}

		return number, nil
	case []any:
		out := make([]any, 0, len(value))
		for idx, item := range value {
			normalized, err := normalizeJSONValue(item)
			if err != nil {
				return nil, fmt.Errorf("at index %d: %w", idx, err)
			}

			out = append(out, normalized)
		}

		return out, nil
	case map[string]any:
		out := make(map[string]any, len(value))
		for key, item := range value {
			normalized, err := normalizeJSONValue(item)
			if err != nil {
				return nil, fmt.Errorf("at key %q: %w", key, err)
			}

			out[key] = normalized
		}

		return out, nil
	default:
		return input, nil
	}
}
