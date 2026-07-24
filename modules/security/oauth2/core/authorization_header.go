package core

import (
	"fmt"
	"strings"
)

// AuthorizationHeader returns a canonical Bearer authorization header.
func AuthorizationHeader(token *TokenSet) (map[string]string, error) {
	if token == nil {
		return nil, fmt.Errorf("%w: token is required", ErrInvalidGrant)
	}
	if !strings.EqualFold(token.TokenType, "Bearer") {
		return nil, fmt.Errorf("%w: token type must be Bearer", ErrInvalidGrant)
	}
	if token.AccessToken == "" || !validBearerToken(token.AccessToken) {
		return nil, fmt.Errorf("%w: access token is not a valid Bearer token", ErrInvalidGrant)
	}

	return map[string]string{
		"Authorization": "Bearer " + token.AccessToken,
	}, nil
}

func validBearerToken(token string) bool {
	padding := false

	for index := 0; index < len(token); index++ {
		char := token[index]
		if char == '=' {
			padding = true

			continue
		}
		if padding {
			return false
		}
		if (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '.' || char == '_' || char == '~' ||
			char == '+' || char == '/' {
			continue
		}

		return false
	}

	return token != ""
}
