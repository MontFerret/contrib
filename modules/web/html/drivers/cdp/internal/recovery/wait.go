// Package recovery provides opt-in recovery for repeatable CDP operations.
package recovery

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/events"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Unlimited retries stale failures for the lifetime of the caller's context.
const Unlimited = -1

// Wait polls an explicitly repeatable operation using the normal polling interval.
// None means incomplete. Only recognized stale errors consume retries; zero
// disables recovery. invalidate must discard caller-owned state without RPCs so
// replacement acquisition happens on the next poll, never in the failing attempt.
func Wait(ctx context.Context, operation events.Function, invalidate func(), retries int) (runtime.Value, error) {
	return events.NewWaitTask(func(ctx context.Context) (runtime.Value, error) {
		out, err := operation(ctx)
		if err == nil {
			return out, nil
		}

		if ctx.Err() != nil {
			return runtime.None, ctx.Err()
		}

		if !eval.IsStaleError(err) {
			return runtime.None, err
		}

		invalidate()

		if retries == 0 {
			return runtime.None, err
		}

		if retries > 0 {
			retries--
		}

		return runtime.None, nil
	}, events.DefaultPolling).Run(ctx)
}
