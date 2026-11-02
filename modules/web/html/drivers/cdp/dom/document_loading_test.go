package dom

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

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDocumentLoadRecoversStaleContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := context.Background()
		start := time.Now()
		oldFrame := testFrameTree("frame", "https://old.example")
		oldFrame.Frame.LoaderID = "old-loader"
		newFrame := testFrameTree("frame", "https://new.example")
		newFrame.Frame.LoaderID = "new-loader"
		creates, calls, trees := 0, 0, 0
		client := &cdp.Client{
			Page: &loadingPageAPI{
				frameTree: func(context.Context) (*page.GetFrameTreeReply, error) {
					trees++
					if time.Since(start) != 200*time.Millisecond {
						t.Fatal("frame metadata was refreshed before the polling interval")
					}

					return &page.GetFrameTreeReply{FrameTree: newFrame}, nil
				},
				createWorld: func(got context.Context, args *page.CreateIsolatedWorldArgs) (*page.CreateIsolatedWorldReply, error) {
					creates++
					if got != ctx || args.FrameID != "frame" {
						t.Fatal("changed caller context or frame identity")
					}

					return &page.CreateIsolatedWorldReply{ExecutionContextID: cdpruntime.ExecutionContextID(creates)}, nil
				},
			},
			Runtime: &loadingRuntimeAPI{call: func(_ context.Context, args *cdpruntime.CallFunctionOnArgs) (*cdpruntime.CallFunctionOnReply, error) {
				calls++
				if args.ExecutionContextID == nil || int(*args.ExecutionContextID) != calls {
					t.Fatal("discarded execution context was reused")
				}

				if calls == 1 {
					return nil, fmt.Errorf("cdp.Runtime: CallFunctionOn: %w", &rpcc.ResponseError{Code: -32000, Message: "Cannot find context with specified id"})
				}

				return loadingRootReply(), nil
			}},
		}
		manager, err := New(zerolog.Nop(), client, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		doc, err := manager.LoadDocument(ctx, oldFrame)
		if err != nil {
			t.Fatal(err)
		}

		if creates != 2 || calls != 2 || trees != 1 || time.Since(start) != 200*time.Millisecond {
			t.Fatalf("creates=%d, calls=%d, trees=%d, elapsed=%s", creates, calls, trees, time.Since(start))
		}

		if doc.Frame().Frame.URL != newFrame.Frame.URL || doc.Frame().Frame.LoaderID != newFrame.Frame.LoaderID || doc.currentState().generation != 1 {
			t.Fatal("recovered document has stale metadata or an extra generation")
		}
	})
}

func BenchmarkLoadDocument(b *testing.B) {
	world := &page.CreateIsolatedWorldReply{ExecutionContextID: 1}
	reply := loadingRootReply()
	client := &cdp.Client{
		Page: &loadingPageAPI{createWorld: func(context.Context, *page.CreateIsolatedWorldArgs) (*page.CreateIsolatedWorldReply, error) {
			return world, nil
		}},
		Runtime: &loadingRuntimeAPI{call: func(context.Context, *cdpruntime.CallFunctionOnArgs) (*cdpruntime.CallFunctionOnReply, error) {
			return reply, nil
		}},
	}
	manager, err := New(zerolog.Nop(), client, nil, nil)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	frame := testFrameTree("frame", "https://example.com")
	b.ReportAllocs()

	for b.Loop() {
		if _, err := manager.LoadDocument(ctx, frame); err != nil {
			b.Fatal(err)
		}
	}
}

func loadingRootReply() *cdpruntime.CallFunctionOnReply {
	id := cdpruntime.RemoteObjectID("root")

	return &cdpruntime.CallFunctionOnReply{Result: cdpruntime.RemoteObject{Type: "object", ObjectID: &id}}
}

type loadingTrace struct {
	stages   []string
	times    []time.Duration
	contexts []cdpruntime.ExecutionContextID
}

func newLoadingTestManager(t *testing.T, beforeRPC func(context.Context, string) error) (*Manager, *loadingPageAPI, *loadingRuntimeAPI, *loadingTrace) {
	t.Helper()

	trace := new(loadingTrace)
	start := time.Now()
	record := func(ctx context.Context, stage string) error {
		trace.stages = append(trace.stages, stage)
		trace.times = append(trace.times, time.Since(start))

		if beforeRPC != nil {
			return beforeRPC(ctx, stage)
		}

		return nil
	}
	worldID := cdpruntime.ExecutionContextID(0)
	pageAPI := &loadingPageAPI{
		frameTree: func(ctx context.Context) (*page.GetFrameTreeReply, error) {
			if err := record(ctx, "tree"); err != nil {
				return nil, err
			}

			return &page.GetFrameTreeReply{FrameTree: testFrameTree("frame", "https://new.example")}, nil
		},
		createWorld: func(ctx context.Context, _ *page.CreateIsolatedWorldArgs) (*page.CreateIsolatedWorldReply, error) {
			worldID++
			if err := record(ctx, "create"); err != nil {
				return nil, err
			}

			return &page.CreateIsolatedWorldReply{ExecutionContextID: worldID}, nil
		},
	}
	runtimeAPI := &loadingRuntimeAPI{call: func(ctx context.Context, args *cdpruntime.CallFunctionOnArgs) (*cdpruntime.CallFunctionOnReply, error) {
		trace.contexts = append(trace.contexts, *args.ExecutionContextID)
		if err := record(ctx, "eval"); err != nil {
			return nil, err
		}

		return loadingRootReply(), nil
	}}
	manager, err := New(zerolog.Nop(), &cdp.Client{Page: pageAPI, Runtime: runtimeAPI}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	return manager, pageAPI, runtimeAPI, trace
}

func TestDocumentLoadRetryBudget(t *testing.T) {
	stale := &rpcc.ResponseError{Code: -32000, Message: "Cannot find context with specified id"}
	for _, tt := range []struct {
		name         string
		failStage    string
		failures     int
		wantStages   []string
		wantContexts []cdpruntime.ExecutionContextID
	}{
		{name: "healthy", wantStages: []string{"create", "eval"}, wantContexts: []cdpruntime.ExecutionContextID{1}},
		{name: "creation recovered", failStage: "create", failures: 1, wantStages: []string{"create", "tree", "create", "eval"}, wantContexts: []cdpruntime.ExecutionContextID{2}},
		{name: "creation exhausted", failStage: "create", failures: 2, wantStages: []string{"create", "tree", "create"}},
		{name: "evaluation exhausted", failStage: "eval", failures: 2, wantStages: []string{"create", "eval", "tree", "create", "eval"}, wantContexts: []cdpruntime.ExecutionContextID{1, 2}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				failures := tt.failures
				manager, _, _, trace := newLoadingTestManager(t, func(_ context.Context, stage string) error {
					if stage == tt.failStage && failures > 0 {
						failures--

						return stale
					}

					return nil
				})
				start := time.Now()
				doc, err := manager.LoadDocument(t.Context(), testFrameTree("frame", "old"))
				if tt.failures == 2 {
					if doc != nil || !errors.Is(err, stale) {
						t.Fatalf("document=%v, error=%v", doc, err)
					}
				} else if err != nil || doc.currentState().generation != 1 {
					t.Fatalf("load: %v", err)
				}

				if !slices.Equal(trace.stages, tt.wantStages) || !slices.Equal(trace.contexts, tt.wantContexts) {
					t.Fatalf("stages=%v, contexts=%v", trace.stages, trace.contexts)
				}

				elapsed := time.Duration(0)
				if tt.failures > 0 {
					elapsed = 200 * time.Millisecond
				}

				if time.Since(start) != elapsed {
					t.Fatalf("elapsed=%s", time.Since(start))
				}

				for i, stage := range trace.stages {
					if stage == "tree" && trace.times[i] != 200*time.Millisecond {
						t.Fatalf("retry started at %s", trace.times[i])
					}
				}
			})
		})
	}
}

func TestDocumentLoadTerminalErrors(t *testing.T) {
	for _, want := range []error{
		&rpcc.ResponseError{Code: -32000, Message: "unrelated protocol failure"},
		&rpcc.ResponseError{Code: -32000, Message: "No frame for given id found"},
		&rpcc.ResponseError{Code: -32602, Message: "Cannot find context with specified id"},
		errors.New("Cannot find context with specified id"),
		context.Canceled,
		context.DeadlineExceeded,
	} {
		for _, phase := range []string{"create", "eval", "tree"} {
			t.Run(want.Error()+"/"+phase, func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					manager, _, _, trace := newLoadingTestManager(t, func(_ context.Context, stage string) error {
						if stage == phase {
							return want
						}

						if phase == "tree" && stage == "eval" {
							return &rpcc.ResponseError{Code: -32000, Message: "Execution context was destroyed"}
						}

						return nil
					})

					doc, err := manager.LoadDocument(t.Context(), testFrameTree("frame", "old"))
					if doc != nil || !errors.Is(err, want) {
						t.Fatalf("document=%v, error=%v, want %v", doc, err, want)
					}

					wantCalls := map[string]int{"create": 1, "eval": 2, "tree": 3}[phase]
					if len(trace.stages) != wantCalls {
						t.Fatalf("unexpected RPCs: %v", trace.stages)
					}
				})
			})
		}
	}
}

func TestDocumentLoadJavaScriptException(t *testing.T) {
	manager, _, runtimeAPI, trace := newLoadingTestManager(t, nil)
	description := "Error: Cannot find context with specified id"
	calls := 0
	runtimeAPI.call = func(context.Context, *cdpruntime.CallFunctionOnArgs) (*cdpruntime.CallFunctionOnReply, error) {
		calls++

		return &cdpruntime.CallFunctionOnReply{ExceptionDetails: &cdpruntime.ExceptionDetails{
			Exception: &cdpruntime.RemoteObject{Description: &description},
		}}, nil
	}

	doc, err := manager.LoadDocument(t.Context(), testFrameTree("frame", "old"))
	if doc != nil || !errors.Is(err, runtime.ErrUnexpected) || calls != 1 || len(trace.stages) != 1 {
		t.Fatalf("document=%v, error=%v, calls=%d, stages=%v", doc, err, calls, trace.stages)
	}
}

func TestDocumentLoadCancellation(t *testing.T) {
	for _, phase := range []string{"before", "create", "eval", "tree", "after create", "after tree", "delay", "deadline"} {
		t.Run(phase, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()

				want := context.Canceled
				if phase == "deadline" {
					var stop context.CancelFunc
					ctx, stop = context.WithTimeout(ctx, 100*time.Millisecond)
					defer stop()
					want = context.DeadlineExceeded
				}

				deadline, hasDeadline := ctx.Deadline()
				if phase == "before" {
					cancel()
				}

				manager, _, _, trace := newLoadingTestManager(t, func(got context.Context, stage string) error {
					gotDeadline, gotHasDeadline := got.Deadline()
					if got != ctx || hasDeadline != gotHasDeadline || !gotDeadline.Equal(deadline) {
						t.Fatal("caller context or deadline changed")
					}

					if phase == "after "+stage {
						cancel()

						return nil
					}

					if stage == phase {
						go cancel()
						<-got.Done()

						return got.Err()
					}

					if stage == "eval" {
						if phase == "delay" {
							go func() {
								time.Sleep(50 * time.Millisecond)
								cancel()
							}()
						}

						return &rpcc.ResponseError{Code: -32000, Message: "Execution context was destroyed"}
					}

					return nil
				})
				start := time.Now()

				doc, err := manager.LoadDocument(ctx, testFrameTree("frame", "old"))
				if doc != nil || !errors.Is(err, want) {
					t.Fatalf("document=%v, error=%v", doc, err)
				}

				wantCalls := map[string]int{"before": 0, "create": 1, "after create": 1, "eval": 2, "tree": 3, "after tree": 3, "delay": 2, "deadline": 2}[phase]
				if len(trace.stages) != wantCalls {
					t.Fatalf("unexpected RPCs after cancellation: %v", trace.stages)
				}

				elapsed := map[string]time.Duration{"tree": 200 * time.Millisecond, "after tree": 200 * time.Millisecond, "delay": 50 * time.Millisecond, "deadline": 100 * time.Millisecond}[phase]
				if time.Since(start) != elapsed {
					t.Fatalf("cancellation took %s, want %s", time.Since(start), elapsed)
				}
			})
		})
	}
}
