package core

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

func cloneTokenParameters(parameters Parameters) Parameters {
	if parameters == nil {
		return make(Parameters)
	}

	cloned := make(Parameters, len(parameters))

	for key, values := range parameters {
		cloned[key] = append([]string(nil), values...)
	}

	return cloned
}

func validateTokenParameters(parameters Parameters, reserved map[string]struct{}) error {
	for key := range parameters {
		if key == "" {
			return fmt.Errorf("%w: parameter name is required", ErrInvalidGrant)
		}

		for _, char := range key {
			if unicode.IsControl(char) {
				return fmt.Errorf("%w: parameter name contains control characters", ErrInvalidGrant)
			}
		}

		if _, found := reserved[key]; found || tokenIsClientAuthParameter(key) {
			return fmt.Errorf("%w: parameter %q is reserved", ErrInvalidGrant, key)
		}
	}

	return nil
}

func tokenIsClientAuthParameter(name string) bool {
	switch name {
	case "client_id", "client_secret":
		return true
	default:
		return strings.HasPrefix(name, "client_assertion")
	}
}

func encodeTokenScope(scopes []string) (string, error) {
	if len(scopes) == 0 {
		return "", nil
	}

	for _, scope := range scopes {
		if err := validateTokenScopePart(scope); err != nil {
			return "", err
		}
	}

	return strings.Join(scopes, " "), nil
}

func validateTokenScopeString(scope string) error {
	if scope == "" {
		return fmt.Errorf("%w: scope must not be empty", ErrInvalidGrant)
	}

	for _, part := range strings.Split(scope, " ") {
		if err := validateTokenScopePart(part); err != nil {
			return err
		}
	}

	return nil
}

func validateTokenScopePart(scope string) error {
	if scope == "" {
		return fmt.Errorf("%w: scope values must not be empty", ErrInvalidGrant)
	}

	for index := 0; index < len(scope); index++ {
		char := scope[index]
		if char != 0x21 && (char < 0x23 || char > 0x5b) && (char < 0x5d || char > 0x7e) {
			return fmt.Errorf("%w: scope contains an invalid character", ErrInvalidGrant)
		}
	}

	return nil
}

func tokenScopeFallback(parameters Parameters) (string, error) {
	values := parameters["scope"]
	if len(values) == 0 {
		return "", nil
	}

	parts := make([]string, 0, len(values))

	for _, value := range values {
		if err := validateTokenScopeString(value); err != nil {
			return "", err
		}

		parts = append(parts, value)
	}

	return strings.Join(parts, " "), nil
}

func tokenScopesAreSubset(requested []string, existing string) bool {
	if len(requested) == 0 || existing == "" {
		return true
	}

	allowed := make(map[string]struct{})
	for _, scope := range strings.Split(existing, " ") {
		if scope != "" {
			allowed[scope] = struct{}{}
		}
	}

	for _, scope := range requested {
		if _, found := allowed[scope]; !found {
			return false
		}
	}

	return true
}

func validateExtensionGrantType(grantType string) error {
	if grantType == "" {
		return fmt.Errorf("%w: grant_type is required", ErrInvalidGrant)
	}
	if strings.TrimSpace(grantType) != grantType || strings.ContainsAny(grantType, "\r\n\t") {
		return fmt.Errorf("%w: grant_type must be an absolute URI", ErrInvalidGrant)
	}

	parsed, err := url.Parse(grantType)
	if err != nil || !parsed.IsAbs() || parsed.Scheme == "" || parsed.Fragment != "" {
		return fmt.Errorf("%w: grant_type must be an absolute URI", ErrInvalidGrant)
	}

	return nil
}
