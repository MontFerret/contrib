package drivers

import (
	"context"

	"github.com/goccy/go-json"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// HTTPRequest HTTP request object.
type (
	HTTPRequest struct {
		URL     string
		Method  string
		Headers *HTTPHeaders
		Body    []byte
	}

	// requestMarshal is a structure that repeats HTTPRequest. It allows
	// easily Marshal the HTTPRequest object.
	requestMarshal struct {
		URL     string       `json:"url"`
		Method  string       `json:"method"`
		Headers *HTTPHeaders `json:"headers"`
		Body    []byte       `json:"body"`
	}
)

func (req *HTTPRequest) Type() runtime.Type {
	return HTTPRequestType
}

func (req *HTTPRequest) String() string {
	return req.URL
}

func (req *HTTPRequest) Compare(other runtime.Value) int {
	otherReq, ok := other.(*HTTPRequest)

	if !ok {
		return CompareTo(HTTPResponseType, other)
	}

	comp := req.Headers.CompareTo(otherReq.Headers)

	if comp != 0 {
		return comp
	}

	comp = runtime.NewString(req.Method).Compare(runtime.NewString(otherReq.Method))

	if comp != 0 {
		return comp
	}

	return runtime.NewString(req.URL).
		Compare(runtime.NewString(otherReq.URL))
}

func (req *HTTPRequest) Unwrap() any {
	return req
}

func (req *HTTPRequest) Hash() uint64 {
	return runtime.Parse(req).Hash()
}

func (req *HTTPRequest) Copy() runtime.Value {
	cop := *req
	return &cop
}

func (req *HTTPRequest) Get(_ context.Context, key runtime.Value) (runtime.Value, error) {
	if key == runtime.None || key == runtime.EmptyString {
		return req, nil
	}

	field := key.String()

	switch field {
	case "url", "URL":
		return runtime.NewString(req.URL), nil
	case "method":
		return runtime.NewString(req.Method), nil
	case "headers":
		return NewHTTPHeadersProxy(req.Headers), nil
	case "body":
		return runtime.NewBinary(req.Body), nil
	}

	return runtime.None, nil
}

func (req *HTTPRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(requestMarshal{
		URL:     req.URL,
		Method:  req.Method,
		Headers: req.Headers,
		Body:    req.Body,
	})
}
