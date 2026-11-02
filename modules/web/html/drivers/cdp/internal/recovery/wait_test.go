package recovery

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/mafredri/cdp/rpcc"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestWaitRetryBudget(t *testing.T) {
	stale := fmt.Errorf("wrapped: %w", &rpcc.ResponseError{Code: -32000, Message: "Cannot find context with specified id"})

	for _, retries := range []int{0, 1, Unlimited} {
		t.Run(fmt.Sprint(retries), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				calls, invalidations := 0, 0
				start := time.Now()
				out, err := Wait(t.Context(), func(context.Context) (runtime.Value, error) {
					calls++
					if calls == 1 {
						return runtime.None, nil
					}

					if calls <= 3 {
						return runtime.None, stale
					}

					return runtime.False, nil
				}, func() { invalidations++ }, retries)

				wantCalls, wantInvalidations := retries+2, retries+1
				if retries == Unlimited {
					wantCalls, wantInvalidations = 4, 2
					if err != nil || out != runtime.False {
						t.Fatalf("result=%v, error=%v", out, err)
					}
				} else if err != stale || out != runtime.None {
					t.Fatalf("exhaustion lost the original error: %v", err)
				}

				if calls != wantCalls || invalidations != wantInvalidations || time.Since(start) != time.Duration(wantCalls-1)*200*time.Millisecond {
					t.Fatalf("calls=%d, invalidations=%d, elapsed=%s", calls, invalidations, time.Since(start))
				}
			})
		})
	}
}

func TestWaitTerminalError(t *testing.T) {
	terminal := &rpcc.ResponseError{Code: -32000, Message: "No frame for given id found"}
	out, err := Wait(t.Context(), func(context.Context) (runtime.Value, error) {
		return runtime.True, terminal
	}, func() { t.Fatal("terminal error invalidated state") }, Unlimited)
	if err != terminal || out != runtime.None {
		t.Fatalf("result=%v, error=%v", out, err)
	}
}

func TestWaitCancellation(t *testing.T) {
	for _, phase := range []string{"before", "during", "delay", "deadline"} {
		t.Run(phase, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()

				want := context.Canceled
				if phase == "deadline" {
					var stop context.CancelFunc
					ctx, stop = context.WithTimeout(ctx, 450*time.Millisecond)
					defer stop()
					want = context.DeadlineExceeded
				}

				if phase == "before" {
					cancel()
				}

				calls := 0
				entered := make(chan struct{}, 4)
				finished := make(chan error, 1)
				start := time.Now()

				go func() {
					_, err := Wait(ctx, func(context.Context) (runtime.Value, error) {
						calls++
						entered <- struct{}{}
						if phase == "during" {
							<-ctx.Done()
						}

						return runtime.None, &rpcc.ResponseError{Code: -32000, Message: "Execution context was destroyed"}
					}, func() {}, Unlimited)
					finished <- err
				}()

				if phase == "during" || phase == "delay" {
					<-entered
					synctest.Wait()
					cancel()
				}

				if err := <-finished; !errors.Is(err, want) {
					t.Fatalf("error=%v, want %v", err, want)
				}

				wantCalls, elapsed := 1, time.Duration(0)
				if phase == "before" {
					wantCalls = 0
				}

				if phase == "deadline" {
					wantCalls, elapsed = 3, 450*time.Millisecond
				}

				if calls != wantCalls || time.Since(start) != elapsed {
					t.Fatalf("calls=%d, elapsed=%s", calls, time.Since(start))
				}
			})
		})
	}
}
