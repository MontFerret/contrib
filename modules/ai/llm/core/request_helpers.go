package core

import "strings"

func joinInstructions(parts ...string) string {
	nonempty := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			nonempty = append(nonempty, part)
		}
	}

	return strings.Join(nonempty, "\n\n")
}

func copyMessages(messages []Message) []Message {
	if len(messages) == 0 {
		return nil
	}

	return append([]Message(nil), messages...)
}

func validateLabels(labels []string) error {
	if len(labels) == 0 {
		return NewError(ErrInvalidOptions, "labels must not be empty")
	}

	seen := make(map[string]struct{}, len(labels))

	for _, label := range labels {
		if label == "" {
			return NewError(ErrInvalidOptions, "labels must not contain empty strings")
		}

		if _, exists := seen[label]; exists {
			return NewError(ErrInvalidOptions, "labels must be unique")
		}

		seen[label] = struct{}{}
	}

	return nil
}
