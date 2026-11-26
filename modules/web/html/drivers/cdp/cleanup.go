package cdp

import (
	"context"
	"errors"
	"time"

	"github.com/mafredri/cdp/rpcc"
)

func closeConnection(conn *rpcc.Conn) error {
	if conn == nil || conn.Context().Err() != nil {
		return nil
	}

	err := conn.Close()
	if errors.Is(err, rpcc.ErrConnClosing) {
		return nil
	}

	return err
}
func defaultCleanupContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), policyCleanupTimeout)
}

func primaryCleanupContext(ctx context.Context) (context.Context, context.CancelFunc) {
	budget := policyCleanupTimeout
	if deadline, ok := ctx.Deadline(); ok {
		budget = min(budget, time.Until(deadline))
	}

	return context.WithTimeout(ctx, budget/2)
}
