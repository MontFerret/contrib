package core

import (
	"context"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

var (
	dblQuotedReplacer = strings.NewReplacer(
		"\"", "\\\"",
		"\\", "\\\\",
		"\x00", `\u0000`,
		"\x01", `\u0001`,
		"\x02", `\u0002`,
		"\x03", `\u0003`,
		"\x04", `\u0004`,
		"\x05", `\u0005`,
		"\x06", `\u0006`,
		"\x07", `\u0007`,
		"\b", `\b`,
		"\t", `\t`,
		"\n", `\n`,
		"\x0b", `\u000b`,
		"\f", `\f`,
		"\r", `\r`,
		"\x0e", `\u000e`,
		"\x0f", `\u000f`,
		"\x10", `\u0010`,
		"\x11", `\u0011`,
		"\x12", `\u0012`,
		"\x13", `\u0013`,
		"\x14", `\u0014`,
		"\x15", `\u0015`,
		"\x16", `\u0016`,
		"\x17", `\u0017`,
		"\x18", `\u0018`,
		"\x19", `\u0019`,
		"\x1a", `\u001a`,
		"\x1b", `\u001b`,
		"\x1c", `\u001c`,
		"\x1d", `\u001d`,
		"\x1e", `\u001e`,
		"\x1f", `\u001f`,
		"\x7f", `\u007f`,
	)
)

type (
	fieldKind int

	mapEntry struct {
		value runtime.Value
		key   string
	}

	encoder struct {
		opts    EncodeOptions
		builder strings.Builder
	}
)

const (
	fieldInline fieldKind = iota
	fieldTable
	fieldArrayTables
)

// Encode serializes a Ferret runtime value into TOML text.
func Encode(ctx context.Context, value runtime.Value, opts EncodeOptions) (string, error) {
	root, err := runtime.CastMap(value)
	if err != nil {
		return "", wrapError(err, "top-level TOML value must be an object")
	}

	enc := &encoder{opts: opts}

	if err := enc.writeTableBody(ctx, nil, root); err != nil {
		return "", err
	}

	return strings.TrimSuffix(enc.builder.String(), "\n"), nil
}

func (e *encoder) writeTableBody(ctx context.Context, path []string, table runtime.Map) error {
	entries, err := collectMapEntries(ctx, table, e.opts.SortKeys)
	if err != nil {
		return err
	}

	inlineEntries := make([]mapEntry, 0, len(entries))
	tableEntries := make([]mapEntry, 0, len(entries))
	arrayTableEntries := make([]mapEntry, 0, len(entries))

	for _, entry := range entries {
		kind, err := e.classifyFieldValue(ctx, entry.value)
		if err != nil {
			return runtime.Errorf(err, "at key %q", entry.key)
		}

		switch kind {
		case fieldInline:
			inlineEntries = append(inlineEntries, entry)
		case fieldTable:
			tableEntries = append(tableEntries, entry)
		case fieldArrayTables:
			arrayTableEntries = append(arrayTableEntries, entry)
		default:
			return newErrorf("unsupported TOML field classification for key %q", entry.key)
		}
	}

	for _, entry := range inlineEntries {
		value, err := e.renderInlineValue(ctx, entry.value)
		if err != nil {
			return runtime.Errorf(err, "at key %q", entry.key)
		}

		e.builder.WriteString(formatKeyPart(entry.key))
		e.builder.WriteString(" = ")
		e.builder.WriteString(value)
		e.builder.WriteByte('\n')
	}

	for _, entry := range tableEntries {
		child, err := runtime.CastMap(entry.value)
		if err != nil {
			return runtime.Errorf(err, "at key %q", entry.key)
		}

		childPath := appendPath(path, entry.key)
		e.startTable(childPath)

		if err := e.writeTableBody(ctx, childPath, child); err != nil {
			return runtime.Errorf(err, "at key %q", entry.key)
		}
	}

	for _, entry := range arrayTableEntries {
		list, ok := entry.value.(runtime.List)
		if !ok {
			return runtime.Errorf(newError("arrays of tables require array values"), "at key %q", entry.key)
		}

		childPath := appendPath(path, entry.key)
		if err := e.writeArrayTables(ctx, childPath, list); err != nil {
			return runtime.Errorf(err, "at key %q", entry.key)
		}
	}

	return nil
}

func (e *encoder) classifyFieldValue(ctx context.Context, value runtime.Value) (fieldKind, error) {
	if isNoneValue(value) {
		return fieldInline, newError("TOML does not support null values")
	}

	switch current := value.(type) {
	case runtime.String, runtime.Boolean, runtime.Int, runtime.Float, runtime.DateTime:
		return fieldInline, nil
	case runtime.Map:
		return fieldTable, nil
	case runtime.List:
		return e.classifyListField(ctx, current)
	case runtime.Binary:
		return fieldInline, newError("unsupported value type for TOML encoding: Binary")
	case runtime.Iterator:
		return fieldInline, newError("unsupported value type for TOML encoding: Iterator")
	case runtime.Observable:
		return fieldInline, newError("unsupported value type for TOML encoding: Observable")
	case runtime.Queryable:
		return fieldInline, newError("unsupported value type for TOML encoding: Queryable")
	default:
		return fieldInline, newErrorf("unsupported value type for TOML encoding: %T", value)
	}
}

func (e *encoder) classifyListField(ctx context.Context, list runtime.List) (fieldKind, error) {
	items, err := collectListItems(ctx, list)
	if err != nil {
		return fieldInline, err
	}

	if len(items) == 0 {
		return fieldInline, nil
	}

	hasObjects := false
	hasNonObjects := false

	for idx, item := range items {
		if isNoneValue(item) {
			return fieldInline, runtime.Errorf(newError("TOML arrays cannot contain null values"), "at index %d", idx)
		}

		if _, ok := item.(runtime.Map); ok {
			hasObjects = true
			continue
		}

		hasNonObjects = true
	}

	if hasObjects && hasNonObjects {
		return fieldInline, newError("mixed arrays containing objects and non-objects are not representable in TOML")
	}

	if hasObjects {
		return fieldArrayTables, nil
	}

	if _, err := e.renderInlineArray(ctx, list); err != nil {
		return fieldInline, err
	}

	return fieldInline, nil
}

func (e *encoder) renderInlineValue(ctx context.Context, value runtime.Value) (string, error) {
	if isNoneValue(value) {
		return "", newError("TOML does not support null values")
	}

	switch current := value.(type) {
	case runtime.String:
		return quoteString(current.String()), nil
	case runtime.Boolean:
		return strconv.FormatBool(bool(current)), nil
	case runtime.Int:
		return strconv.FormatInt(int64(current), 10), nil
	case runtime.Float:
		return formatFloat(float64(current)), nil
	case runtime.DateTime:
		return encodeTemporalValue(current, e.opts)
	case runtime.List:
		return e.renderInlineArray(ctx, current)
	case runtime.Map:
		return "", newError("objects (inline tables) inside TOML arrays are not supported by this encoder; use arrays-of-tables instead")
	case runtime.Binary:
		return "", newError("unsupported value type for TOML encoding: Binary")
	case runtime.Iterator:
		return "", newError("unsupported value type for TOML encoding: Iterator")
	case runtime.Observable:
		return "", newError("unsupported value type for TOML encoding: Observable")
	case runtime.Queryable:
		return "", newError("unsupported value type for TOML encoding: Queryable")
	default:
		return "", newErrorf("unsupported value type for TOML encoding: %T", value)
	}
}

func (e *encoder) renderInlineArray(ctx context.Context, list runtime.List) (string, error) {
	items, err := collectListItems(ctx, list)
	if err != nil {
		return "", err
	}

	if len(items) == 0 {
		return "[]", nil
	}

	parts := make([]string, 0, len(items))

	for idx, item := range items {
		if _, ok := item.(runtime.Map); ok {
			return "", runtime.Errorf(newError("objects are not representable inside TOML arrays in v1"), "at index %d", idx)
		}

		encoded, err := e.renderInlineValue(ctx, item)
		if err != nil {
			return "", runtime.Errorf(err, "at index %d", idx)
		}

		parts = append(parts, encoded)
	}

	return "[" + strings.Join(parts, ", ") + "]", nil
}

func (e *encoder) writeArrayTables(ctx context.Context, path []string, list runtime.List) error {
	items, err := collectListItems(ctx, list)
	if err != nil {
		return err
	}

	for idx, item := range items {
		table, err := runtime.CastMap(item)
		if err != nil {
			return runtime.Errorf(newError("arrays of tables require object elements"), "at index %d", idx)
		}

		e.startArrayTable(path)

		if err := e.writeTableBody(ctx, path, table); err != nil {
			return runtime.Errorf(err, "at index %d", idx)
		}
	}

	return nil
}

func (e *encoder) startTable(path []string) {
	if e.builder.Len() > 0 {
		e.builder.WriteByte('\n')
	}

	e.builder.WriteByte('[')
	e.builder.WriteString(formatKeyPath(path))
	e.builder.WriteString("]\n")
}

func (e *encoder) startArrayTable(path []string) {
	if e.builder.Len() > 0 {
		e.builder.WriteByte('\n')
	}

	e.builder.WriteString("[[")
	e.builder.WriteString(formatKeyPath(path))
	e.builder.WriteString("]]\n")
}

func collectMapEntries(ctx context.Context, m runtime.Map, sortKeys bool) ([]mapEntry, error) {
	entries := make([]mapEntry, 0)

	if err := m.ForEach(ctx, func(_ context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		name, ok := key.(runtime.String)
		if !ok {
			return false, newError("TOML object keys must be strings")
		}

		entries = append(entries, mapEntry{
			key:   name.String(),
			value: value,
		})

		return true, nil
	}); err != nil {
		return nil, err
	}

	if sortKeys {
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].key < entries[j].key
		})
	}

	return entries, nil
}

func collectListItems(ctx context.Context, list runtime.List) ([]runtime.Value, error) {
	items := make([]runtime.Value, 0)

	if err := runtime.ForEach(ctx, list, func(_ context.Context, value, _ runtime.Value) (runtime.Boolean, error) {
		items = append(items, value)
		return true, nil
	}); err != nil {
		if err == io.EOF {
			return items, nil
		}

		return nil, err
	}

	return items, nil
}

func appendPath(path []string, key string) []string {
	out := make([]string, len(path)+1)
	copy(out, path)
	out[len(path)] = key

	return out
}

func formatKeyPath(path []string) string {
	parts := make([]string, len(path))

	for idx, part := range path {
		parts[idx] = formatKeyPart(part)
	}

	return strings.Join(parts, ".")
}

func formatKeyPart(key string) string {
	if key == "" {
		return `""`
	}

	for _, r := range key {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			continue
		}

		return quoteString(key)
	}

	return key
}

func quoteString(input string) string {
	return `"` + dblQuotedReplacer.Replace(input) + `"`
}

func formatFloat(value float64) string {
	switch {
	case math.IsNaN(value):
		if math.Signbit(value) {
			return "-nan"
		}

		return "nan"
	case math.IsInf(value, 1):
		return "inf"
	case math.IsInf(value, -1):
		return "-inf"
	default:
		return floatAddDecimal(strconv.FormatFloat(value, 'g', -1, 64))
	}
}

func floatAddDecimal(raw string) string {
	for _, ch := range raw {
		if ch == 'e' || ch == 'E' {
			return raw
		}

		if ch == '.' {
			return raw
		}
	}

	return raw + ".0"
}

func isNoneValue(value runtime.Value) bool {
	return value == nil || value == runtime.None
}
