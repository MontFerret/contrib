package drivers

import (
	"context"
	"net/textproto"
	"sort"
	"strings"

	"github.com/goccy/go-json"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// HTTPHeaders HTTP header object
type HTTPHeaders struct {
	Data textproto.MIMEHeader
}

func NewHTTPHeaders() *HTTPHeaders {
	return &HTTPHeaders{textproto.MIMEHeader{}}
}

func NewHTTPHeadersWith(values map[string][]string) *HTTPHeaders {
	return &HTTPHeaders{textproto.MIMEHeader(values)}
}

func (h *HTTPHeaders) Type() runtime.Type {
	return HTTPHeadersType
}

func (h *HTTPHeaders) Equal(ctx context.Context, other runtime.Value) (bool, error) {
	if _, err := checkComparisonContext(ctx); err != nil {
		return false, err
	}

	otherHeaders, ok := other.(*HTTPHeaders)
	if !ok {
		return false, nil
	}

	comparison, err := h.compare(ctx, otherHeaders)

	return comparison == runtime.Equal, err
}

func (h *HTTPHeaders) Compare(ctx context.Context, other runtime.Value) (runtime.Ordering, error) {
	if comparison, err := checkComparisonContext(ctx); err != nil {
		return comparison, err
	}

	otherHeaders, ok := other.(*HTTPHeaders)
	if !ok {
		return invalidComparison(h, other)
	}

	return h.compare(ctx, otherHeaders)
}

func (h *HTTPHeaders) compare(ctx context.Context, other *HTTPHeaders) (runtime.Ordering, error) {
	if comparison, err := checkComparisonContext(ctx); err != nil {
		return comparison, err
	}

	if len(h.Data) < len(other.Data) {
		return runtime.Less, nil
	}
	if len(h.Data) > len(other.Data) {
		return runtime.Greater, nil
	}

	keys := sortedHeaderKeys(h.Data)
	otherKeys := sortedHeaderKeys(other.Data)
	for idx, key := range keys {
		if comparison := compareStrings(key, otherKeys[idx]); comparison != runtime.Equal {
			return comparison, nil
		}

		values := h.Data[key]
		otherValues := other.Data[key]
		if len(values) < len(otherValues) {
			return runtime.Less, nil
		}
		if len(values) > len(otherValues) {
			return runtime.Greater, nil
		}

		for valueIdx, value := range values {
			if comparison := compareStrings(value, otherValues[valueIdx]); comparison != runtime.Equal {
				return comparison, nil
			}
		}
	}

	return runtime.Equal, nil
}

func (h *HTTPHeaders) String() string {
	var sb strings.Builder

	sb.WriteString("HTTP Headers:\n")

	for k, v := range h.Data {
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(strings.Join(v, ", "))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (h *HTTPHeaders) Hash() uint64 {
	var builder strings.Builder

	for _, key := range sortedHeaderKeys(h.Data) {
		builder.WriteString(key)
		builder.WriteByte(0)
		for _, value := range h.Data[key] {
			builder.WriteString(value)
			builder.WriteByte(0)
		}
	}

	return runtime.Hash(runtime.TypeName(h.Type()), []byte(builder.String()))
}

func (h *HTTPHeaders) Copy() runtime.Value {
	return &HTTPHeaders{h.Data}
}

func (h *HTTPHeaders) MarshalJSON() ([]byte, error) {
	data := make(map[string]string, len(h.Data))

	for key, values := range h.Data {
		data[key] = strings.Join(values, ", ")
	}

	return json.Marshal(data)
}

func (h *HTTPHeaders) Get(_ context.Context, key runtime.Value) (runtime.Value, error) {
	return runtime.String(h.Data.Get(key.String())), nil
}

func (h *HTTPHeaders) Set(_ context.Context, key, value runtime.Value) error {
	h.Data.Set(key.String(), value.String())

	return nil
}

func (h *HTTPHeaders) Iterate(_ context.Context) (runtime.Iterator, error) {
	return sdk.NewMapIterator(h.Data), nil
}

func (h *HTTPHeaders) Clone() *HTTPHeaders {
	clone := make(textproto.MIMEHeader)

	for k, v := range h.Data {
		clone[k] = v
	}

	return &HTTPHeaders{Data: clone}
}

func sortedHeaderKeys(headers textproto.MIMEHeader) []string {
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}
