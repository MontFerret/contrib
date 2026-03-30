package drivers

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/goccy/go-json"
)

// HTTPResponse HTTP response object.
type (
	HTTPResponse struct {
		Headers      *HTTPHeaders
		URL          string
		Status       string
		Body         []byte
		StatusCode   int
		ResponseTime float64
	}

	// responseMarshal is a structure that repeats HTTPResponse. It allows
	// easily Marshal the HTTPResponse object.
	responseMarshal struct {
		Headers      *HTTPHeaders `json:"headers"`
		URL          string       `json:"url"`
		Status       string       `json:"status"`
		Body         []byte       `json:"body"`
		StatusCode   int          `json:"status_code"`
		ResponseTime float64      `json:"response_time"`
	}
)

func (resp *HTTPResponse) Type() runtime.Type {
	return HTTPResponseType
}

func (resp *HTTPResponse) String() string {
	return resp.Status
}

func (resp *HTTPResponse) Compare(other runtime.Value) int {
	otherResp, ok := other.(*HTTPResponse)

	if !ok {
		return CompareTo(HTTPResponseType, other)
	}

	comp := resp.Headers.CompareTo(otherResp.Headers)
	if comp != 0 {
		return comp
	}

	// it makes no sense to compare Status strings
	// because they are always equal if StatusCode's are equal
	return runtime.NewInt(resp.StatusCode).
		Compare(runtime.NewInt(resp.StatusCode))
}

func (resp *HTTPResponse) Unwrap() any {
	return resp
}

func (resp *HTTPResponse) Copy() runtime.Value {
	cop := *resp
	return &cop
}

func (resp *HTTPResponse) Hash() uint64 {
	return runtime.Parse(resp).Hash()
}

func (resp *HTTPResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(responseMarshal{
		URL:          resp.URL,
		StatusCode:   resp.StatusCode,
		Status:       resp.Status,
		Headers:      resp.Headers,
		Body:         resp.Body,
		ResponseTime: resp.ResponseTime,
	})
}

func (resp *HTTPResponse) Get(_ context.Context, key runtime.Value) (runtime.Value, error) {
	if key == runtime.None || key == runtime.EmptyString {
		return resp, nil
	}

	field := key.String()

	switch field {
	case "url", "URL":
		return runtime.NewString(resp.URL), nil
	case "status":
		return runtime.NewString(resp.Status), nil
	case "statusCode":
		return runtime.NewInt(resp.StatusCode), nil
	case "headers":
		return NewHTTPHeadersProxy(resp.Headers), nil
	case "body":
		return runtime.NewBinary(resp.Body), nil
	case "responseTime":
		return runtime.NewFloat(resp.ResponseTime), nil
	}

	return runtime.None, nil
}
