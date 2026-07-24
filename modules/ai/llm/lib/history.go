package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// History returns a copied array of visible text messages from a session.
func History(_ context.Context, value runtime.Value) (runtime.Value, error) {
	session, err := sessionValue(value)
	if err != nil {
		return runtime.None, err
	}

	messages := session.History()
	values := make([]runtime.Value, 0, len(messages))

	for _, message := range messages {
		values = append(values, messageValue(message))
	}

	return runtime.NewArrayWith(values...), nil
}

func messageValue(message core.Message) runtime.Value {
	return runtime.NewObjectWith(map[string]runtime.Value{
		"role":    runtime.NewString(string(message.Role)),
		"content": runtime.NewString(message.Content.Text),
	})
}
