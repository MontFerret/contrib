package core

import (
	"net/url"
	"sort"
	"strings"
	"unicode"

	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
)

func tokenSubmittedSecrets(client *Client, form Parameters, headers ferrethttp.Headers) []string {
	secrets := make([]string, 0, 8)
	if client != nil && client.ClientSecret != "" {
		secrets = append(secrets, client.ClientSecret)
	}

	for key, values := range form {
		if tokenSensitiveParameter(key) {
			secrets = append(secrets, values...)
		}
	}

	for key, values := range headers {
		if strings.EqualFold(key, "Authorization") {
			for _, value := range values {
				secrets = append(secrets, value)

				if separator := strings.IndexByte(value, ' '); separator >= 0 {
					secrets = append(secrets, value[separator+1:])
				}
			}
		}
	}

	return secrets
}

func tokenSensitiveParameter(name string) bool {
	name = strings.ToLower(name)

	switch name {
	case "access_token", "client_secret", "refresh_token", "id_token",
		"assertion", "code", "code_verifier", "authorization":
		return true
	default:
		return strings.HasPrefix(name, "client_assertion")
	}
}

func sanitizeTokenText(input string, secrets []string) string {
	var builder strings.Builder
	builder.Grow(len(input))

	for _, char := range input {
		if unicode.IsControl(char) {
			builder.WriteByte(' ')

			continue
		}

		builder.WriteRune(char)
	}

	result := builder.String()
	variants := make([]string, 0, len(secrets)*2)

	for _, secret := range secrets {
		if secret == "" {
			continue
		}

		variants = append(variants, secret)
		if encoded := url.QueryEscape(secret); encoded != secret {
			variants = append(variants, encoded)
		}
	}

	sort.SliceStable(variants, func(left, right int) bool {
		return len(variants[left]) > len(variants[right])
	})

	for _, secret := range variants {
		result = strings.ReplaceAll(result, secret, redactedValue)
	}

	return result
}
