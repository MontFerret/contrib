package core

import (
	"context"
	"fmt"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type messageInput struct {
	Role    *string `json:"role"`
	Content *string `json:"content"`
}

// DecodeMessages validates the text-only v1 chat message format.
func DecodeMessages(ctx context.Context, value runtime.Value) ([]Message, error) {
	if value == nil || runtime.TypeNone.Is(value) {
		return nil, nil
	}

	input, err := decodeList[[]messageInput](ctx, value, "messages")
	if err != nil {
		return nil, err
	}

	messages := make([]Message, 0, len(input))
	for index, item := range input {
		if item.Role == nil {
			return nil, NewError(ErrInvalidOptions, fmt.Sprintf("messages[%d].role is required", index))
		}

		role := Role(*item.Role)
		switch role {
		case RoleSystem, RoleDeveloper, RoleUser, RoleAssistant:
		default:
			return nil, NewError(ErrInvalidOptions, fmt.Sprintf("messages[%d].role is unsupported", index))
		}

		if item.Content == nil {
			return nil, NewError(ErrInvalidOptions, fmt.Sprintf("messages[%d].content is required", index))
		}

		messages = append(messages, TextMessage(role, *item.Content))
	}

	return messages, nil
}

// TextMessage creates one provider-neutral text message.
func TextMessage(role Role, text string) Message {
	return Message{Role: role, Content: Content{Type: ContentText, Text: text}}
}
