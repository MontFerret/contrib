package cdp

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"testing/synctest"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"
	"github.com/mafredri/cdp/rpcc"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/dom"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/events"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestWaitForDocumentReadyRecoversContext(t *testing.T) {
	stale := &rpcc.ResponseError{Code: -32000, Message: "Cannot find context with specified id"}
	tests := []struct {
		name           string
		creationErrs   []error
		evaluationErrs []error
		wantContexts   []cdpruntime.ExecutionContextID
		wantCreations  int
		wantPolls      int
	}{
		{name: "immediately ready", wantContexts: []cdpruntime.ExecutionContextID{1}, wantCreations: 1},
		{name: "not ready then ready", evaluationErrs: []error{nil}, wantContexts: []cdpruntime.ExecutionContextID{1, 1}, wantCreations: 1, wantPolls: 1},
		{name: "stale before first evaluation", evaluationErrs: []error{stale}, wantContexts: []cdpruntime.ExecutionContextID{1, 2}, wantCreations: 2, wantPolls: 1},
		{
			name:           "navigation between polls",
			evaluationErrs: []error{nil, fmt.Errorf("cdp.Runtime: CallFunctionOn: %w", stale)},
			wantContexts:   []cdpruntime.ExecutionContextID{1, 1, 2},
			wantCreations:  2,
			wantPolls:      2,
		},
		{name: "repeated navigation", evaluationErrs: []error{stale, stale}, wantContexts: []cdpruntime.ExecutionContextID{1, 2, 3}, wantCreations: 3, wantPolls: 2},
		{name: "initial creation interrupted", creationErrs: []error{stale}, wantContexts: []cdpruntime.ExecutionContextID{2}, wantCreations: 2, wantPolls: 1},
		{
			name:           "replacement creation interrupted",
			creationErrs:   []error{nil, stale},
			evaluationErrs: []error{stale},
			wantContexts:   []cdpruntime.ExecutionContextID{1, 3},
			wantCreations:  3,
			wantPolls:      2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx := t.Context()
				creations := 0
				var contexts []cdpruntime.ExecutionContextID
				pageAPI := &readinessPageAPI{createWorld: func(got context.Context, args *page.CreateIsolatedWorldArgs) (*page.CreateIsolatedWorldReply, error) {
					if got != ctx || args.FrameID != "child-frame" {
						t.Fatalf("context creation changed caller context or frame: %s", args.FrameID)
					}

					creations++
					if creations <= len(tt.creationErrs) && tt.creationErrs[creations-1] != nil {
						return nil, tt.creationErrs[creations-1]
					}

					return &page.CreateIsolatedWorldReply{ExecutionContextID: cdpruntime.ExecutionContextID(creations)}, nil
				}}
				runtimeAPI := &readinessRuntimeAPI{call: func(got context.Context, args *cdpruntime.CallFunctionOnArgs) (*cdpruntime.CallFunctionOnReply, error) {
					if got != ctx || args.ExecutionContextID == nil {
						t.Fatal("readiness evaluation lost its caller context or execution context")
					}

					contexts = append(contexts, *args.ExecutionContextID)
					if len(contexts) <= len(tt.evaluationErrs) {
						return readinessReply(false), tt.evaluationErrs[len(contexts)-1]
					}

					return readinessReply(true), nil
				}}
				client := &cdp.Client{Page: pageAPI, Runtime: runtimeAPI}
				// A different root client catches accidental fallback from the requested frame's session.
				p := NewHTMLPage(zerolog.Nop(), new(cdp.Client), nil, nil, nil)
				start := time.Now()

				if err := p.waitForDocumentReady(ctx, client, "child-frame"); err != nil {
					t.Fatalf("wait for readiness: %v", err)
				}

				if creations != tt.wantCreations || !slices.Equal(contexts, tt.wantContexts) {
					t.Fatalf("creations = %d, evaluated contexts = %v; want %d, %v", creations, contexts, tt.wantCreations, tt.wantContexts)
				}

				if elapsed := time.Since(start); elapsed != time.Duration(tt.wantPolls)*events.DefaultPolling {
					t.Fatalf("polling took %v, want %v", elapsed, time.Duration(tt.wantPolls)*events.DefaultPolling)
				}
			})
		})
	}
}

func TestWaitForDocumentReadyPreservesErrors(t *testing.T) {
	stale := &rpcc.ResponseError{Code: -32000, Message: "Cannot find context with specified id"}
	description := "Error: Cannot find context with specified id"
	tests := []struct {
		err   error
		reply *cdpruntime.CallFunctionOnReply
		name  string
	}{
		{name: "unrelated protocol error", err: &rpcc.ResponseError{Code: -32000, Message: "Some other protocol error"}},
		{name: "missing frame", err: &rpcc.ResponseError{Code: -32000, Message: "No frame for given id found"}},
		{name: "wrong code", err: &rpcc.ResponseError{Code: -32602, Message: stale.Message}},
		{name: "plain matching text", err: errors.New(stale.Message)},
		{name: "canceled RPC", err: context.Canceled},
		{name: "expired RPC", err: context.DeadlineExceeded},
		{
			name: "JavaScript exception",
			err:  runtime.ErrUnexpected,
			reply: &cdpruntime.CallFunctionOnReply{ExceptionDetails: &cdpruntime.ExceptionDetails{
				Exception: &cdpruntime.RemoteObject{Description: &description},
			}},
		},
	}

	for _, tt := range tests {
		for _, phase := range []string{"creation", "evaluation", "replacement"} {
			if tt.reply != nil && phase != "evaluation" {
				continue
			}

			t.Run(tt.name+"/"+phase, func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(t.Context(), time.Second)
					defer cancel()

					creations, evaluations := 0, 0
					pageAPI := &readinessPageAPI{createWorld: func(context.Context, *page.CreateIsolatedWorldArgs) (*page.CreateIsolatedWorldReply, error) {
						creations++
						if phase == "creation" || (phase == "replacement" && creations > 1) {
							return nil, tt.err
						}

						return &page.CreateIsolatedWorldReply{ExecutionContextID: 1}, nil
					}}
					runtimeAPI := &readinessRuntimeAPI{call: func(context.Context, *cdpruntime.CallFunctionOnArgs) (*cdpruntime.CallFunctionOnReply, error) {
						evaluations++
						if phase == "replacement" {
							return nil, stale
						}

						if tt.reply != nil {
							return tt.reply, nil
						}

						return nil, tt.err
					}}
					client := &cdp.Client{Page: pageAPI, Runtime: runtimeAPI}
					p := NewHTMLPage(zerolog.Nop(), client, nil, nil, nil)

					err := p.waitForDocumentReady(ctx, client, "frame")
					if !errors.Is(err, tt.err) {
						t.Fatalf("readiness error = %v, want %v", err, tt.err)
					}

					wantCreations, wantEvaluations := 1, 1
					if phase == "creation" {
						wantEvaluations = 0
					}

					if phase == "replacement" {
						wantCreations = 2
					}

					if creations != wantCreations || evaluations != wantEvaluations {
						t.Fatalf("creation/evaluation calls = %d/%d, want %d/%d", creations, evaluations, wantCreations, wantEvaluations)
					}
				})
			})
		}
	}
}

func TestWaitForDocumentReadyCancellation(t *testing.T) {
	for _, phase := range []string{"before starting", "during creation", "during evaluation", "between polls"} {
		t.Run(phase, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()

				creations, evaluations := 0, 0
				entered := make(chan struct{})
				pageAPI := &readinessPageAPI{createWorld: func(ctx context.Context, _ *page.CreateIsolatedWorldArgs) (*page.CreateIsolatedWorldReply, error) {
					creations++
					if phase == "during creation" {
						close(entered)
						<-ctx.Done()

						return nil, ctx.Err()
					}

					return &page.CreateIsolatedWorldReply{ExecutionContextID: 1}, nil
				}}
				runtimeAPI := &readinessRuntimeAPI{call: func(ctx context.Context, _ *cdpruntime.CallFunctionOnArgs) (*cdpruntime.CallFunctionOnReply, error) {
					evaluations++
					close(entered)
					if phase == "during evaluation" {
						<-ctx.Done()

						return nil, ctx.Err()
					}

					return nil, &rpcc.ResponseError{Code: -32000, Message: "Execution context was destroyed."}
				}}
				client := &cdp.Client{Page: pageAPI, Runtime: runtimeAPI}
				p := NewHTMLPage(zerolog.Nop(), client, nil, nil, nil)

				if phase == "before starting" {
					cancel()
				}

				result := make(chan error, 1)
				go func() {
					result <- p.waitForDocumentReady(ctx, client, "frame")
				}()

				if phase != "before starting" {
					<-entered
					synctest.Wait()
					cancel()
				}

				synctest.Wait()

				select {
				case err := <-result:
					if !errors.Is(err, context.Canceled) {
						t.Fatalf("readiness error = %v, want cancellation", err)
					}
				default:
					t.Error("readiness did not return immediately on cancellation")
					<-result
				}

				wantCreations, wantEvaluations := 1, 1
				if phase == "before starting" {
					wantCreations, wantEvaluations = 0, 0
				} else if phase == "during creation" {
					wantEvaluations = 0
				}

				if creations != wantCreations || evaluations != wantEvaluations {
					t.Fatalf("creation/evaluation calls = %d/%d, want %d/%d", creations, evaluations, wantCreations, wantEvaluations)
				}
			})
		})
	}
}

func TestWaitForDocumentReadyKeepsDeadline(t *testing.T) {
	for _, phase := range []string{"creation", "evaluation"} {
		t.Run(phase, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				start := time.Now()
				ctx, cancel := context.WithTimeout(t.Context(), 450*time.Millisecond)
				defer cancel()

				stale := &rpcc.ResponseError{Code: -32000, Message: "Cannot find context with specified id"}
				var attempts []time.Duration
				pageAPI := &readinessPageAPI{createWorld: func(got context.Context, _ *page.CreateIsolatedWorldArgs) (*page.CreateIsolatedWorldReply, error) {
					if got != ctx {
						t.Fatal("context creation replaced the caller context")
					}

					attempts = append(attempts, time.Since(start))
					if phase == "creation" {
						return nil, stale
					}

					return &page.CreateIsolatedWorldReply{ExecutionContextID: cdpruntime.ExecutionContextID(len(attempts))}, nil
				}}
				runtimeAPI := &readinessRuntimeAPI{call: func(context.Context, *cdpruntime.CallFunctionOnArgs) (*cdpruntime.CallFunctionOnReply, error) {
					return nil, stale
				}}
				client := &cdp.Client{Page: pageAPI, Runtime: runtimeAPI}
				p := NewHTMLPage(zerolog.Nop(), client, nil, nil, nil)

				err := p.waitForDocumentReady(ctx, client, "frame")
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("readiness error = %v, want deadline", err)
				}

				wantAttempts := []time.Duration{0, 200 * time.Millisecond, 400 * time.Millisecond}
				if !slices.Equal(attempts, wantAttempts) || time.Since(start) != 450*time.Millisecond {
					t.Fatalf("attempts = %v, elapsed = %v; want %v and 450ms", attempts, time.Since(start), wantAttempts)
				}
			})
		})
	}
}

func TestLoadNavigatedMainFrameRunsInitScriptOnceAfterRecovery(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var operations []string
		creations := 0
		pageAPI := &readinessPageAPI{
			frameTree: page.FrameTree{Frame: page.Frame{ID: "frame"}},
			createWorld: func(context.Context, *page.CreateIsolatedWorldArgs) (*page.CreateIsolatedWorldReply, error) {
				creations++

				return &page.CreateIsolatedWorldReply{ExecutionContextID: cdpruntime.ExecutionContextID(creations)}, nil
			},
		}
		runtimeAPI := &readinessRuntimeAPI{
			call: func(_ context.Context, args *cdpruntime.CallFunctionOnArgs) (*cdpruntime.CallFunctionOnReply, error) {
				switch *args.ExecutionContextID {
				case 1:
					operations = append(operations, "stale")

					return nil, &rpcc.ResponseError{Code: -32000, Message: "Cannot find context with specified id"}
				case 2:
					operations = append(operations, "ready")

					return readinessReply(true), nil
				default:
					operations = append(operations, "load document")
					id := cdpruntime.RemoteObjectID("document")

					return &cdpruntime.CallFunctionOnReply{Result: cdpruntime.RemoteObject{Type: "object", ObjectID: &id}}, nil
				}
			},
			evaluate: func(_ context.Context, args *cdpruntime.EvaluateArgs) (*cdpruntime.EvaluateReply, error) {
				if args.Expression != "window.initCount = (window.initCount || 0) + 1" {
					t.Fatalf("unexpected init script: %s", args.Expression)
				}

				operations = append(operations, "init script")

				return &cdpruntime.EvaluateReply{}, nil
			},
		}
		client := &cdp.Client{Page: pageAPI, Runtime: runtimeAPI}
		manager, err := dom.New(zerolog.Nop(), client, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() { _ = manager.Close() })
		p := NewHTMLPage(zerolog.Nop(), client, nil, nil, manager)
		p.initScript = &drivers.InitScript{
			Source: "window.initCount = (window.initCount || 0) + 1",
			Timing: drivers.InitScriptAfterNavigation,
		}

		if err := p.loadNavigatedMainFrame(t.Context()); err != nil {
			t.Fatalf("load navigated frame: %v", err)
		}

		want := []string{"stale", "ready", "init script", "load document"}
		if !slices.Equal(operations, want) {
			t.Fatalf("operations = %v, want %v", operations, want)
		}

		if manager.GetMainFrame() == nil {
			t.Fatal("ready document was not installed")
		}
	})
}

func BenchmarkWaitForDocumentReady(b *testing.B) {
	world := &page.CreateIsolatedWorldReply{ExecutionContextID: 1}
	reply := readinessReply(true)
	client := &cdp.Client{
		Page: &readinessPageAPI{createWorld: func(context.Context, *page.CreateIsolatedWorldArgs) (*page.CreateIsolatedWorldReply, error) {
			return world, nil
		}},
		Runtime: &readinessRuntimeAPI{call: func(context.Context, *cdpruntime.CallFunctionOnArgs) (*cdpruntime.CallFunctionOnReply, error) {
			return reply, nil
		}},
	}
	p := NewHTMLPage(zerolog.Nop(), client, nil, nil, nil)
	ctx := context.Background()
	b.ReportAllocs()

	for b.Loop() {
		if err := p.waitForDocumentReady(ctx, client, "frame"); err != nil {
			b.Fatal(err)
		}
	}
}

func readinessReply(ready bool) *cdpruntime.CallFunctionOnReply {
	if ready {
		return &cdpruntime.CallFunctionOnReply{Result: cdpruntime.RemoteObject{Type: "boolean", Value: []byte("true")}}
	}

	return &cdpruntime.CallFunctionOnReply{Result: cdpruntime.RemoteObject{Type: "object", Value: []byte("null")}}
}
