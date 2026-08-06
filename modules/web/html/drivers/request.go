package drivers

import (
	"context"
	"strconv"

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

func (req *HTTPRequest) Equal(ctx context.Context, other runtime.Value) (bool, error) {
	if _, err := checkComparisonContext(ctx); err != nil {
		return false, err
	}

	otherReq, ok := other.(*HTTPRequest)
	if !ok {
		return false, nil
	}

	comparison, err := req.compare(ctx, otherReq)

	return comparison == runtime.Equal, err
}

func (req *HTTPRequest) Compare(ctx context.Context, other runtime.Value) (runtime.Ordering, error) {
	if comparison, err := checkComparisonContext(ctx); err != nil {
		return comparison, err
	}

	otherReq, ok := other.(*HTTPRequest)
	if !ok {
		return invalidComparison(req, other)
	}

	return req.compare(ctx, otherReq)
}

func (req *HTTPRequest) compare(ctx context.Context, other *HTTPRequest) (runtime.Ordering, error) {
	if comparison, err := checkComparisonContext(ctx); err != nil {
		return comparison, err
	}

	comparison, err := req.Headers.Compare(ctx, other.Headers)
	if err != nil {
		return runtime.Equal, err
	}

	if comparison != runtime.Equal {
		return comparison, nil
	}

	if comparison = compareStrings(req.Method, other.Method); comparison != runtime.Equal {
		return comparison, nil
	}

	return compareStrings(req.URL, other.URL), nil
}

func (req *HTTPRequest) Unwrap() any {
	return req
}

func (req *HTTPRequest) Hash() uint64 {
	content := strconv.FormatUint(req.Headers.Hash(), 10) + "\x00" + req.Method + "\x00" + req.URL

	return runtime.Hash(runtime.TypeName(req.Type()), []byte(content))
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
		return req.Headers, nil
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
