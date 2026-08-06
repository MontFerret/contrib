package drivers

import (
	"context"
	"strconv"

	"github.com/goccy/go-json"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
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

func (resp *HTTPResponse) Equal(ctx context.Context, other runtime.Value) (bool, error) {
	if _, err := checkComparisonContext(ctx); err != nil {
		return false, err
	}

	otherResp, ok := other.(*HTTPResponse)
	if !ok {
		return false, nil
	}

	comparison, err := resp.compare(ctx, otherResp)

	return comparison == runtime.Equal, err
}

func (resp *HTTPResponse) Compare(ctx context.Context, other runtime.Value) (runtime.Ordering, error) {
	if comparison, err := checkComparisonContext(ctx); err != nil {
		return comparison, err
	}

	otherResp, ok := other.(*HTTPResponse)
	if !ok {
		return invalidComparison(resp, other)
	}

	return resp.compare(ctx, otherResp)
}

func (resp *HTTPResponse) compare(ctx context.Context, other *HTTPResponse) (runtime.Ordering, error) {
	if comparison, err := checkComparisonContext(ctx); err != nil {
		return comparison, err
	}

	comparison, err := resp.Headers.Compare(ctx, other.Headers)
	if err != nil {
		return runtime.Equal, err
	}

	if comparison != runtime.Equal {
		return comparison, nil
	}

	// it makes no sense to compare Status strings
	// because they are always equal if StatusCode's are equal
	if resp.StatusCode < other.StatusCode {
		return runtime.Less, nil
	}

	if resp.StatusCode > other.StatusCode {
		return runtime.Greater, nil
	}

	return runtime.Equal, nil
}

func (resp *HTTPResponse) Unwrap() any {
	return resp
}

func (resp *HTTPResponse) Copy() runtime.Value {
	cop := *resp
	return &cop
}

func (resp *HTTPResponse) Hash() uint64 {
	content := strconv.FormatUint(resp.Headers.Hash(), 10) + "\x00" + strconv.Itoa(resp.StatusCode)

	return runtime.Hash(runtime.TypeName(resp.Type()), []byte(content))
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
		return resp.Headers, nil
	case "body":
		return runtime.NewBinary(resp.Body), nil
	case "responseTime":
		return runtime.NewFloat(resp.ResponseTime), nil
	}

	return runtime.None, nil
}
