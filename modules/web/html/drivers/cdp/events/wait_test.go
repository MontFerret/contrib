package events

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestWaitTaskPreservesResultsAndErrors(t *testing.T) {
	wantErr := errors.New("evaluation failed")
	tests := []struct {
		value runtime.Value
		err   error
		name  string
		polls int
	}{
		{name: "immediate result", value: runtime.True},
		{name: "false is a result", value: runtime.False},
		{name: "result after polling", value: runtime.NewString("ready"), polls: 2},
		{name: "terminal error", value: runtime.None, err: wantErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				calls := 0
				task := NewWaitTask(func(context.Context) (runtime.Value, error) {
					calls++
					if calls <= tt.polls {
						return runtime.None, nil
					}

					return tt.value, tt.err
				}, DefaultPolling)
				start := time.Now()
				value, err := task.Run(t.Context())
				if value != tt.value || !errors.Is(err, tt.err) {
					t.Fatalf("Run() = %v, %v; want %v, %v", value, err, tt.value, tt.err)
				}

				if calls != tt.polls+1 || time.Since(start) != time.Duration(tt.polls)*DefaultPolling {
					t.Fatalf("calls = %d, elapsed = %v; want %d polls", calls, time.Since(start), tt.polls)
				}
			})
		})
	}
}

func TestWaitTaskCancellationInterruptsPolling(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		calls := 0
		task := NewWaitTask(func(context.Context) (runtime.Value, error) {
			calls++

			return runtime.None, nil
		}, time.Hour)
		done := make(chan error, 1)
		go func() {
			_, err := task.Run(ctx)
			done <- err
		}()
		synctest.Wait()
		cancel()
		synctest.Wait()

		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("Run() error = %v, want cancellation", err)
			}
		default:
			t.Error("cancellation did not interrupt the polling timer")
			<-done
		}

		if calls != 1 {
			t.Fatalf("evaluation calls = %d, want 1", calls)
		}
	})
}

func TestWaitTaskCanceledBeforeStarting(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	task := NewWaitTask(func(context.Context) (runtime.Value, error) {
		t.Fatal("called evaluator with canceled context")

		return runtime.None, nil
	}, DefaultPolling)

	value, err := task.Run(ctx)
	if value != runtime.None || !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() = %v, %v; want None and cancellation", value, err)
	}
}
