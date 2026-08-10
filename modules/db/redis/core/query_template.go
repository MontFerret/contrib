package core

import (
	"fmt"
	"strings"
)

type (
	queryTemplateToken struct {
		parts []queryTemplatePart
	}

	queryTemplatePart struct {
		literal string
		binding string
		spread  bool
	}
)

func parseQueryTemplate(input string) ([]queryTemplateToken, error) {
	tokens := make([]queryTemplateToken, 0)

	for offset := 0; ; {
		for offset < len(input) && isTemplateWhitespace(input[offset]) {
			offset++
		}
		if offset == len(input) {
			break
		}

		token, next, err := parseQueryTemplateToken(input, offset)
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, token)
		offset = next
	}

	if len(tokens) == 0 {
		return nil, fmt.Errorf("redis query must contain a command")
	}

	return tokens, nil
}

func parseQueryTemplateToken(input string, offset int) (queryTemplateToken, int, error) {
	if input[offset] == '\'' || input[offset] == '"' {
		return parseQuotedQueryTemplateToken(input, offset, input[offset])
	}

	parts := make([]queryTemplatePart, 0, 1)
	literal := strings.Builder{}

	for offset < len(input) && !isTemplateWhitespace(input[offset]) {
		switch input[offset] {
		case '\'', '"':
			return queryTemplateToken{}, 0, fmt.Errorf("quoted Redis argument must start with its quote")
		case '\\':
			if offset+1 < len(input) && input[offset+1] == '$' {
				literal.WriteByte('$')
				offset += 2
				continue
			}
		case '$':
			part, next, ok := parseBindingPart(input, offset)
			if ok {
				appendTemplateLiteral(&parts, &literal)
				parts = append(parts, part)
				offset = next
				continue
			}
		}

		literal.WriteByte(input[offset])
		offset++
	}

	appendTemplateLiteral(&parts, &literal)
	token := queryTemplateToken{parts: parts}
	if err := validateSpreadToken(token); err != nil {
		return queryTemplateToken{}, 0, err
	}

	return token, offset, nil
}

func parseQuotedQueryTemplateToken(input string, offset int, quote byte) (queryTemplateToken, int, error) {
	parts := make([]queryTemplatePart, 0, 1)
	literal := strings.Builder{}
	offset++

	for offset < len(input) {
		if input[offset] == quote {
			offset++
			if offset < len(input) && !isTemplateWhitespace(input[offset]) {
				return queryTemplateToken{}, 0, fmt.Errorf("quoted Redis argument must be followed by whitespace or the end of the query")
			}

			appendTemplateLiteral(&parts, &literal)
			if len(parts) == 0 {
				parts = append(parts, queryTemplatePart{literal: ""})
			}

			token := queryTemplateToken{parts: parts}
			if err := validateSpreadToken(token); err != nil {
				return queryTemplateToken{}, 0, err
			}

			return token, offset, nil
		}

		if input[offset] == '\\' {
			escaped, next, err := parseQuotedEscape(input, offset, quote)
			if err != nil {
				return queryTemplateToken{}, 0, err
			}

			literal.WriteString(escaped)
			offset = next
			continue
		}

		if input[offset] == '$' {
			part, next, ok := parseBindingPart(input, offset)
			if ok {
				appendTemplateLiteral(&parts, &literal)
				parts = append(parts, part)
				offset = next
				continue
			}
		}

		literal.WriteByte(input[offset])
		offset++
	}

	return queryTemplateToken{}, 0, fmt.Errorf("unterminated quoted Redis argument")
}

func parseQuotedEscape(input string, offset int, quote byte) (string, int, error) {
	if offset+1 >= len(input) {
		return "", 0, fmt.Errorf("incomplete escape in quoted Redis argument")
	}

	escaped := input[offset+1]
	if quote == '\'' {
		switch escaped {
		case '\'', '\\', '$':
			return string(escaped), offset + 2, nil
		default:
			return "", 0, fmt.Errorf("unsupported escape \\%c in single-quoted Redis argument", escaped)
		}
	}

	switch escaped {
	case '"', '\\', '$':
		return string(escaped), offset + 2, nil
	case 'n':
		return "\n", offset + 2, nil
	case 'r':
		return "\r", offset + 2, nil
	case 't':
		return "\t", offset + 2, nil
	case 'b':
		return "\b", offset + 2, nil
	case 'a':
		return "\a", offset + 2, nil
	case 'x':
		if offset+3 >= len(input) || !isHexDigit(input[offset+2]) || !isHexDigit(input[offset+3]) {
			return "", 0, fmt.Errorf("double-quoted Redis \\x escape requires two hexadecimal digits")
		}

		value := hexDigitValue(input[offset+2])*16 + hexDigitValue(input[offset+3])

		return string([]byte{value}), offset + 4, nil
	default:
		return "", 0, fmt.Errorf("unsupported escape \\%c in double-quoted Redis argument", escaped)
	}
}

func parseBindingPart(input string, offset int) (queryTemplatePart, int, bool) {
	if offset+1 >= len(input) || !isBindingIdentifierStart(input[offset+1]) {
		return queryTemplatePart{}, offset, false
	}

	end := offset + 2
	for end < len(input) && isBindingIdentifierPart(input[end]) {
		end++
	}

	part := queryTemplatePart{binding: input[offset+1 : end]}
	if end+3 <= len(input) && input[end:end+3] == "..." {
		part.spread = true
		end += 3
	}

	return part, end, true
}

func appendTemplateLiteral(parts *[]queryTemplatePart, literal *strings.Builder) {
	if literal.Len() == 0 {
		return
	}

	*parts = append(*parts, queryTemplatePart{literal: literal.String()})
	literal.Reset()
}

func validateSpreadToken(token queryTemplateToken) error {
	for _, part := range token.parts {
		if part.spread && (len(token.parts) != 1 || part.binding == "") {
			return fmt.Errorf("spread binding $%s... must occupy the entire Redis argument", part.binding)
		}
	}

	return nil
}

func isTemplateWhitespace(value byte) bool {
	switch value {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	default:
		return false
	}
}

func isBindingIdentifierStart(value byte) bool {
	return value == '_' || value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z'
}

func isBindingIdentifierPart(value byte) bool {
	return isBindingIdentifierStart(value) || value >= '0' && value <= '9'
}

func isHexDigit(value byte) bool {
	return value >= '0' && value <= '9' || value >= 'a' && value <= 'f' || value >= 'A' && value <= 'F'
}

func hexDigitValue(value byte) byte {
	switch {
	case value >= '0' && value <= '9':
		return value - '0'
	case value >= 'a' && value <= 'f':
		return value - 'a' + 10
	default:
		return value - 'A' + 10
	}
}
