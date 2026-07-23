package core

import (
	"context"
	"fmt"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// DecodeMessages validates the text-only v1 chat message format.
func DecodeMessages(ctx context.Context, value runtime.Value) ([]Message, error) {
	if value == nil || runtime.TypeNone.Is(value) {
		return nil, nil
	}

	list, ok := value.(runtime.List)
	if !ok {
		return nil, NewError(ErrInvalidOptions, "messages must be an array")
	}

	messages := make([]Message, 0)
	err := list.ForEach(ctx, func(ctx context.Context, value runtime.Value, index runtime.Int) (runtime.Boolean, error) {
		fields, err := optionValues(ctx, value, fmt.Sprintf("messages[%d]", index))
		if err != nil {
			return runtime.False, err
		}

		if err := rejectUnknown(fields, map[string]struct{}{"role": {}, "content": {}}, fmt.Sprintf("messages[%d]", index)); err != nil {
			return runtime.False, err
		}

		roleText, found, err := stringOption(fields, "role", fmt.Sprintf("messages[%d]", index))
		if err != nil {
			return runtime.False, err
		}

		if !found {
			return runtime.False, NewError(ErrInvalidOptions, fmt.Sprintf("messages[%d].role is required", index))
		}

		role := Role(roleText)
		switch role {
		case RoleSystem, RoleDeveloper, RoleUser, RoleAssistant:
		default:
			return runtime.False, NewError(ErrInvalidOptions, fmt.Sprintf("messages[%d].role is unsupported", index))
		}

		content, found, err := stringOption(fields, "content", fmt.Sprintf("messages[%d]", index))
		if err != nil {
			return runtime.False, err
		}

		if !found {
			return runtime.False, NewError(ErrInvalidOptions, fmt.Sprintf("messages[%d].content is required", index))
		}

		messages = append(messages, TextMessage(role, content))

		return runtime.True, nil
	})

	if err != nil {
		return nil, err
	}

	return messages, nil
}

// TextMessage creates one provider-neutral text message.
func TextMessage(role Role, text string) Message {
	return Message{Role: role, Content: Content{Type: ContentText, Text: text}}
}
