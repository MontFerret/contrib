package drivers

import (
	"context"
	"net/textproto"
	"strings"

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

func NewHTTPHeadersProxy(header *HTTPHeaders) runtime.Value {
	return sdk.NewProxyWithType(HTTPHeadersType, header)
}

func (h *HTTPHeaders) Type() runtime.Type {
	return HTTPHeadersType
}

func (h *HTTPHeaders) Compare(other runtime.Value) int {
	otherHeaders, ok := other.(*HTTPHeaders)

	if !ok {
		return CompareTo(HTTPHeadersType, other)
	}

	return h.CompareTo(otherHeaders)
}

func (h *HTTPHeaders) CompareTo(other *HTTPHeaders) int {
	if len(h.Data) > len(other.Data) {
		return 1
	} else if len(h.Data) < len(other.Data) {
		return -1
	}

	for k := range h.Data {
		c := strings.Compare(h.Data.Get(k), other.Data.Get(k))

		if c != 0 {
			return c
		}
	}

	return 0
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
	var hash uint64

	for k, v := range h.Data {
		hash += runtime.String(k).Hash()

		for _, vv := range v {
			hash += runtime.String(vv).Hash()
		}
	}

	return hash
}

func (h *HTTPHeaders) Copy() runtime.Value {
	return &HTTPHeaders{h.Data}
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
