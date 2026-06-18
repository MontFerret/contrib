package core

import (
	"context"
	"net/http"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type RequestData struct {
	Query   runtime.Value
	Body    runtime.Value
	Headers http.Header
	Method  string
	HasBody bool
}

func DecodeRequestData(ctx context.Context, value runtime.Value) (RequestData, error) {
	data := RequestData{
		Method:  http.MethodGet,
		Query:   runtime.None,
		Headers: make(http.Header),
		Body:    runtime.None,
	}

	if runtime.TypeNone.Is(value) {
		return data, nil
	}

	obj, err := requireMap(ctx, value, "HTTP query WITH")
	if err != nil {
		return data, err
	}

	methodProvided := false
	if method, found, err := lookupString(ctx, obj, "method", "HTTP query WITH"); err != nil {
		return data, err
	} else if found {
		data.Method = strings.ToUpper(strings.TrimSpace(method))
		methodProvided = true
	}

	if query, found, err := lookupValue(ctx, obj, "query"); err != nil {
		return data, err
	} else if found {
		data.Query = query
	}

	if headers, found, err := lookupValue(ctx, obj, "headers"); err != nil {
		return data, err
	} else if found {
		data.Headers, err = decodeHeaders(ctx, headers, "HTTP query WITH.headers")
		if err != nil {
			return data, err
		}
	}

	if body, found, err := lookupValue(ctx, obj, "body"); err != nil {
		return data, err
	} else if found {
		data.Body = body
		data.HasBody = !runtime.TypeNone.Is(body)
	}

	if !methodProvided && data.HasBody {
		data.Method = http.MethodPost
	}

	if data.Method == "" {
		data.Method = http.MethodGet
	}

	return data, nil
}
