package openai

import (
	"context"
	"errors"
	"net"
	"strings"

	sdkopenai "github.com/openai/openai-go/v3"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
)

func normalizeError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return core.NewError(core.ErrTimeout, "provider request timed out")
	}

	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}

	var networkError net.Error
	if errors.As(err, &networkError) && networkError.Timeout() {
		return core.NewError(core.ErrTimeout, "provider request timed out")
	}

	var apiError *sdkopenai.Error
	if !errors.As(err, &apiError) {
		return core.NewError(core.ErrProvider, "provider request failed")
	}

	switch apiError.StatusCode {
	case 401, 403:
		return core.NewError(core.ErrAuth, "provider authentication failed")
	case 408:
		return core.NewError(core.ErrTimeout, "provider request timed out")
	case 429:
		return core.NewError(core.ErrRateLimit, "provider rate limit exceeded")
	}

	if isContextLimitCode(apiError.Code) || isContextLimitCode(apiError.Type) {
		return core.NewError(core.ErrContextLimit, "provider context limit exceeded")
	}

	return core.NewError(core.ErrProvider, "provider request failed")
}

func isContextLimitCode(code string) bool {
	return strings.EqualFold(strings.TrimSpace(code), "context_length_exceeded")
}
