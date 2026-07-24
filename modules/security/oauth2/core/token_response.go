package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"mime"
	"net/url"
	"strconv"
	"strings"
	"time"

	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
)

type parsedTokenResponse struct {
	token               *TokenSet
	refreshTokenPresent bool
	scopePresent        bool
}

var (
	tokenResponseFields = map[string]struct{}{
		"access_token":  {},
		"token_type":    {},
		"refresh_token": {},
		"scope":         {},
		"expires_in":    {},
		"id_token":      {},
	}

	tokenErrorFields = map[string]struct{}{
		"error":             {},
		"error_description": {},
		"error_uri":         {},
	}
)

func parseTokenEndpointResponse(
	response *ferrethttp.Response,
	now time.Time,
	scopeFallback string,
	refreshTokenFallback string,
	secrets []string,
) (*TokenSet, error) {
	if response == nil {
		return nil, fmt.Errorf("%w: token endpoint returned no response", ErrInvalidTokenResponse)
	}

	if response.StatusCode != 200 {
		if !isTokenJSONContentType(tokenResponseHeader(response.Headers, "Content-Type")) {
			return nil, fmt.Errorf(
				"%w: token endpoint returned HTTP status %d with a non-JSON OAuth error response",
				ErrInvalidTokenResponse,
				response.StatusCode,
			)
		}

		oauthError, err := parseTokenOAuthError(response.Body, response.StatusCode, secrets)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: token endpoint returned HTTP status %d with a malformed OAuth error response",
				ErrInvalidTokenResponse,
				response.StatusCode,
			)
		}

		return nil, oauthError
	}

	if !isTokenJSONContentType(tokenResponseHeader(response.Headers, "Content-Type")) {
		return nil, fmt.Errorf("%w: token endpoint response must use application/json", ErrInvalidTokenResponse)
	}

	parsed, err := parseTokenSuccess(response.Body, now)
	if err != nil {
		return nil, err
	}
	if !parsed.scopePresent {
		parsed.token.Scope = scopeFallback
	}
	if !parsed.refreshTokenPresent || parsed.token.RefreshToken == "" {
		parsed.token.RefreshToken = refreshTokenFallback
	}

	return parsed.token, nil
}

func parseTokenSuccess(body []byte, now time.Time) (*parsedTokenResponse, error) {
	fields, err := decodeTokenJSONObject(body, tokenResponseFields)
	if err != nil {
		return nil, fmt.Errorf("%w: malformed JSON", ErrInvalidTokenResponse)
	}

	accessToken, found, err := decodeTokenString(fields, "access_token")
	if err != nil || !found || accessToken == "" {
		return nil, fmt.Errorf("%w: access_token must be a non-empty string", ErrInvalidTokenResponse)
	}

	tokenType, found, err := decodeTokenString(fields, "token_type")
	if err != nil || !found || tokenType == "" {
		return nil, fmt.Errorf("%w: token_type must be a non-empty string", ErrInvalidTokenResponse)
	}

	result := &parsedTokenResponse{
		token: &TokenSet{
			AccessToken: accessToken,
			TokenType:   tokenType,
			Extra:       make(map[string]any),
		},
	}

	if value, present, decodeErr := decodeTokenString(fields, "refresh_token"); decodeErr != nil {
		return nil, fmt.Errorf("%w: refresh_token must be a string", ErrInvalidTokenResponse)
	} else if present {
		result.refreshTokenPresent = true
		result.token.RefreshToken = value
	}

	if value, present, decodeErr := decodeTokenString(fields, "id_token"); decodeErr != nil {
		return nil, fmt.Errorf("%w: id_token must be a string", ErrInvalidTokenResponse)
	} else if present {
		result.token.IDToken = value
	}

	if value, present, decodeErr := decodeTokenString(fields, "scope"); decodeErr != nil {
		return nil, fmt.Errorf("%w: scope must be a string", ErrInvalidTokenResponse)
	} else if present {
		if err := validateTokenScopeString(value); err != nil {
			return nil, fmt.Errorf("%w: scope is malformed", ErrInvalidTokenResponse)
		}

		result.scopePresent = true
		result.token.Scope = value
	}

	if raw, present := fields["expires_in"]; present {
		expiresIn, decodeErr := decodeTokenExpiresIn(raw)
		if decodeErr != nil {
			return nil, decodeErr
		}

		result.token.ExpiresIn = expiresIn
		result.token.ExpiresAt = now.Add(expiresIn)
	}

	for name, raw := range fields {
		if _, known := tokenResponseFields[name]; known {
			continue
		}

		value, decodeErr := decodeTokenUnknown(raw)
		if decodeErr != nil {
			return nil, fmt.Errorf("%w: malformed provider field", ErrInvalidTokenResponse)
		}

		result.token.Extra[name] = value
	}

	if len(result.token.Extra) == 0 {
		result.token.Extra = nil
	}

	return result, nil
}

func parseTokenOAuthError(body []byte, statusCode int, secrets []string) (*Error, error) {
	fields, err := decodeTokenJSONObject(body, tokenErrorFields)
	if err != nil {
		return nil, fmt.Errorf("%w: malformed OAuth error response", ErrInvalidTokenResponse)
	}

	code, found, err := decodeTokenString(fields, "error")
	if err != nil || !found || code == "" {
		return nil, fmt.Errorf("%w: OAuth error response requires error", ErrInvalidTokenResponse)
	}
	if !validOAuthErrorText(code) {
		return nil, fmt.Errorf("%w: error does not conform to OAuth syntax", ErrInvalidTokenResponse)
	}

	description, descriptionPresent, err := decodeTokenString(fields, "error_description")
	if err != nil {
		return nil, fmt.Errorf("%w: error_description must be a string", ErrInvalidTokenResponse)
	}
	if descriptionPresent && !validOAuthErrorText(description) {
		return nil, fmt.Errorf(
			"%w: error_description does not conform to OAuth syntax",
			ErrInvalidTokenResponse,
		)
	}

	errorURI, errorURIPresent, err := decodeTokenString(fields, "error_uri")
	if err != nil {
		return nil, fmt.Errorf("%w: error_uri must be a string", ErrInvalidTokenResponse)
	}
	if errorURIPresent && !validOAuthErrorURI(errorURI) {
		return nil, fmt.Errorf(
			"%w: error_uri does not conform to OAuth syntax",
			ErrInvalidTokenResponse,
		)
	}

	return &Error{
		Operation:   "token request",
		Code:        sanitizeTokenText(code, secrets),
		Description: sanitizeTokenText(description, secrets),
		URI:         sanitizeTokenText(errorURI, secrets),
		StatusCode:  statusCode,
	}, nil
}

func validOAuthErrorText(value string) bool {
	if value == "" {
		return false
	}

	for index := 0; index < len(value); index++ {
		char := value[index]
		if char < 0x20 || char > 0x7e || char == '"' || char == '\\' {
			return false
		}
	}

	return true
}

func validOAuthErrorURI(value string) bool {
	for index := 0; index < len(value); index++ {
		char := value[index]
		if char < 0x21 || char > 0x7e || char == '"' || char == '\\' {
			return false
		}
	}

	_, err := url.Parse(value)

	return err == nil
}

func decodeTokenJSONObject(body []byte, knownFields map[string]struct{}) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()

	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return nil, errors.New("expected JSON object")
	}

	fields := make(map[string]json.RawMessage)
	seenKnown := make(map[string]struct{})

	for decoder.More() {
		nameToken, tokenErr := decoder.Token()
		if tokenErr != nil {
			return nil, tokenErr
		}

		name, ok := nameToken.(string)
		if !ok {
			return nil, errors.New("expected object field name")
		}

		if _, known := knownFields[name]; known {
			if _, duplicate := seenKnown[name]; duplicate {
				return nil, fmt.Errorf("duplicate field %q", name)
			}

			seenKnown[name] = struct{}{}
		}

		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return nil, err
		}

		fields[name] = append(json.RawMessage(nil), raw...)
	}

	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') {
		return nil, errors.New("expected end of JSON object")
	}

	if token, err := decoder.Token(); !errors.Is(err, io.EOF) {
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("unexpected trailing token %v", token)
	}

	return fields, nil
}

func decodeTokenString(
	fields map[string]json.RawMessage,
	name string,
) (string, bool, error) {
	raw, found := fields[name]
	if !found {
		return "", false, nil
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", true, err
	}

	return value, true, nil
}

func decodeTokenExpiresIn(raw json.RawMessage) (time.Duration, error) {
	value := strings.TrimSpace(string(raw))
	if value == "" {
		return 0, fmt.Errorf("%w: expires_in must be a non-negative integer", ErrInvalidTokenResponse)
	}

	for index := 0; index < len(value); index++ {
		if value[index] < '0' || value[index] > '9' {
			return 0, fmt.Errorf("%w: expires_in must be a non-negative integer", ErrInvalidTokenResponse)
		}
	}

	seconds, err := strconv.ParseUint(value, 10, 64)
	if err != nil || seconds > uint64(math.MaxInt64/int64(time.Second)) {
		return 0, fmt.Errorf("%w: expires_in is out of range", ErrInvalidTokenResponse)
	}

	return time.Duration(seconds) * time.Second, nil
}

func decodeTokenUnknown(raw json.RawMessage) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()

	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("provider field contains trailing JSON")
	}

	return value, nil
}

func isTokenJSONContentType(value string) bool {
	mediaType, _, err := mime.ParseMediaType(value)

	return err == nil && strings.EqualFold(mediaType, "application/json")
}

func tokenResponseHeader(headers ferrethttp.Headers, name string) string {
	for key, values := range headers {
		if strings.EqualFold(key, name) && len(values) > 0 {
			return values[0]
		}
	}

	return ""
}
