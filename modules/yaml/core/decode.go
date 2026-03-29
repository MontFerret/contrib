package core

import (
	"context"
	"errors"
	"io"
	"math"
	"strconv"
	"strings"

	yaml "github.com/goccy/go-yaml"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Decode eagerly decodes a single YAML document into a Ferret runtime value.
func Decode(ctx context.Context, data runtime.String) (runtime.Value, error) {
	docs, err := decodeDocuments(data)
	if err != nil {
		return nil, err
	}

	if len(docs) > 1 {
		return nil, newYAMLError("multiple YAML documents provided to YAML::DECODE")
	}

	return normalizeValue(ctx, docs[0])
}

// DecodeAll eagerly decodes all YAML documents into an array of Ferret values.
func DecodeAll(ctx context.Context, data runtime.String) (runtime.Value, error) {
	docs, err := decodeDocuments(data)
	if err != nil {
		return nil, err
	}

	out := runtime.NewArray(len(docs))

	for idx, doc := range docs {
		value, err := normalizeValue(ctx, doc)
		if err != nil {
			return nil, runtime.Errorf(err, "at document %d", idx+1)
		}

		if err := out.Append(ctx, value); err != nil {
			return nil, err
		}
	}

	return out, nil
}

func decodeDocuments(data runtime.String) ([]any, error) {
	if strings.TrimSpace(data.String()) == "" {
		return nil, newYAMLError("empty YAML input")
	}

	decoder := yaml.NewDecoder(strings.NewReader(data.String()))
	docs := make([]any, 0, 1)

	for {
		var doc any

		if err := decoder.Decode(&doc); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, wrapYAMLError(err, "invalid YAML document")
		}

		docs = append(docs, doc)
	}

	if len(docs) == 0 {
		return nil, newYAMLError("empty YAML input")
	}

	return docs, nil
}

func normalizeValue(ctx context.Context, input any) (runtime.Value, error) {
	switch value := input.(type) {
	case nil:
		return runtime.None, nil
	case bool:
		return runtime.NewBoolean(value), nil
	case string:
		return runtime.NewString(value), nil
	case int:
		return runtime.NewInt(value), nil
	case int8:
		return runtime.NewInt(int(value)), nil
	case int16:
		return runtime.NewInt(int(value)), nil
	case int32:
		return runtime.NewInt(int(value)), nil
	case int64:
		return runtime.NewInt64(value), nil
	case uint:
		if uint64(value) > math.MaxInt64 {
			return nil, newYAMLErrorf("invalid YAML integer %d exceeds Ferret int range", value)
		}

		return runtime.NewInt64(int64(value)), nil
	case uint8:
		return runtime.NewInt(int(value)), nil
	case uint16:
		return runtime.NewInt(int(value)), nil
	case uint32:
		return runtime.NewInt64(int64(value)), nil
	case uint64:
		if value > math.MaxInt64 {
			return nil, newYAMLErrorf("invalid YAML integer %d exceeds Ferret int range", value)
		}

		return runtime.NewInt64(int64(value)), nil
	case float32:
		return runtime.NewFloat(float64(value)), nil
	case float64:
		return runtime.NewFloat(value), nil
	case []any:
		return normalizeArray(ctx, value)
	case map[string]any:
		return normalizeStringMap(ctx, value)
	case map[any]any:
		return normalizeInterfaceMap(ctx, value)
	default:
		return nil, newYAMLErrorf("invalid YAML document: unsupported value type %T", input)
	}
}

func normalizeArray(ctx context.Context, items []any) (runtime.Value, error) {
	out := runtime.NewArray(len(items))

	for idx, item := range items {
		value, err := normalizeValue(ctx, item)
		if err != nil {
			return nil, runtime.Errorf(err, "at index %d", idx)
		}

		if err := out.Append(ctx, value); err != nil {
			return nil, err
		}
	}

	return out, nil
}

func normalizeStringMap(ctx context.Context, input map[string]any) (runtime.Value, error) {
	raw := make(map[any]any, len(input))

	for key, value := range input {
		raw[key] = value
	}

	return normalizeInterfaceMap(ctx, raw)
}

func normalizeInterfaceMap(ctx context.Context, input map[any]any) (runtime.Value, error) {
	out := runtime.NewObject()
	explicit := make([]mapEntry, 0, len(input))

	for rawKey, rawValue := range input {
		key, err := normalizeMapKey(rawKey)
		if err != nil {
			return nil, err
		}

		if key == "<<" {
			if err := applyMerge(ctx, out, rawValue); err != nil {
				return nil, err
			}

			continue
		}

		explicit = append(explicit, mapEntry{
			key:   key,
			value: rawValue,
		})
	}

	for _, entry := range explicit {
		value, err := normalizeValue(ctx, entry.value)
		if err != nil {
			return nil, runtime.Errorf(err, "at key %q", entry.key)
		}

		if err := out.Set(ctx, runtime.NewString(entry.key), value); err != nil {
			return nil, err
		}
	}

	return out, nil
}

func applyMerge(ctx context.Context, out *runtime.Object, rawValue any) error {
	switch value := rawValue.(type) {
	case map[string]any:
		merged, err := normalizeStringMap(ctx, value)
		if err != nil {
			return err
		}

		return mergeObject(ctx, out, merged.(runtime.Map))
	case map[any]any:
		merged, err := normalizeInterfaceMap(ctx, value)
		if err != nil {
			return err
		}

		return mergeObject(ctx, out, merged.(runtime.Map))
	case []any:
		for idx, item := range value {
			if err := applyMerge(ctx, out, item); err != nil {
				return runtime.Errorf(err, "at merge index %d", idx)
			}
		}

		return nil
	default:
		return newYAMLError("invalid YAML document")
	}
}

func mergeObject(ctx context.Context, target *runtime.Object, source runtime.Map) error {
	return source.ForEach(ctx, func(ctx context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		if _, found, err := target.Lookup(ctx, key); err != nil {
			return false, err
		} else if found {
			return true, nil
		}

		if err := target.Set(ctx, key, value); err != nil {
			return false, err
		}

		return true, nil
	})
}

func normalizeMapKey(key any) (string, error) {
	switch value := key.(type) {
	case nil:
		return "null", nil
	case string:
		return value, nil
	case bool:
		return strconv.FormatBool(value), nil
	case int:
		return strconv.Itoa(value), nil
	case int8:
		return strconv.FormatInt(int64(value), 10), nil
	case int16:
		return strconv.FormatInt(int64(value), 10), nil
	case int32:
		return strconv.FormatInt(int64(value), 10), nil
	case int64:
		return strconv.FormatInt(value, 10), nil
	case uint:
		return strconv.FormatUint(uint64(value), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(value), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(value), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(value), 10), nil
	case uint64:
		return strconv.FormatUint(value, 10), nil
	case float32:
		return strconv.FormatFloat(float64(value), 'g', -1, 32), nil
	case float64:
		return strconv.FormatFloat(value, 'g', -1, 64), nil
	default:
		return "", newYAMLErrorf("invalid YAML document: unsupported mapping key type %T", key)
	}
}

type mapEntry struct {
	value any
	key   string
}
