package events

import (
	"context"
	"time"

	"github.com/MontFerret/contrib/modules/html/drivers/cdp/eval"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type (
	Function func(ctx context.Context) (runtime.Value, error)

	WaitTask struct {
		fun     Function
		polling time.Duration
	}
)

const DefaultPolling = time.Millisecond * time.Duration(200)

func NewWaitTask(
	fun Function,
	polling time.Duration,
) *WaitTask {
	return &WaitTask{
		fun,
		polling,
	}
}

func (task *WaitTask) Run(ctx context.Context) (runtime.Value, error) {
	for {
		if ctx.Err() != nil {
			return runtime.None, ctx.Err()
		}

		out, err := task.fun(ctx)

		// expression failed
		// terminating
		if err != nil {
			return runtime.None, err
		}

		// output is not empty
		// terminating
		if out != runtime.None {
			return out, nil
		}

		// Nothing yet, let's wait before the next try
		<-time.After(task.polling)
	}
}

func NewEvalWaitTask(
	ec *eval.Runtime,
	fn *eval.Function,
	polling time.Duration,
) *WaitTask {
	return NewWaitTask(
		func(ctx context.Context) (runtime.Value, error) {
			return ec.EvalValue(ctx, fn)
		},
		polling,
	)
}
