package core

import (
	"fmt"
	"strings"
)

type Encoding string

const (
	EncodingJSON  Encoding = "json"
	EncodingText  Encoding = "text"
	EncodingBytes Encoding = "bytes"
	EncodingForm  Encoding = "form"
)

type ResponseMode string

const (
	ResponseModeBody ResponseMode = "body"
	ResponseModeFull ResponseMode = "full"
)

type ErrorMode string

const (
	ErrorModeRaise    ErrorMode = "raise"
	ErrorModeResponse ErrorMode = "response"
)

func parseEncoding(input string) (Encoding, error) {
	switch enc := Encoding(strings.ToLower(strings.TrimSpace(input))); enc {
	case EncodingJSON, EncodingText, EncodingBytes, EncodingForm:
		return enc, nil
	default:
		return "", fmt.Errorf("unsupported encoding %q", input)
	}
}

func parseResponseMode(input string) (ResponseMode, error) {
	switch mode := ResponseMode(strings.ToLower(strings.TrimSpace(input))); mode {
	case ResponseModeBody, ResponseModeFull:
		return mode, nil
	default:
		return "", fmt.Errorf("unsupported response mode %q", input)
	}
}

func parseErrorMode(input string) (ErrorMode, error) {
	switch mode := ErrorMode(strings.ToLower(strings.TrimSpace(input))); mode {
	case ErrorModeRaise, ErrorModeResponse:
		return mode, nil
	default:
		return "", fmt.Errorf("unsupported error mode %q", input)
	}
}
