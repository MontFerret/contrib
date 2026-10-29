package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2"
	"github.com/MontFerret/ferret/v2/pkg/compiler"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

func TestExecQuerySettlesOutputAndOwnedResources(t *testing.T) {
	runErr := errors.New("execution failed")
	closeErr := errors.New("session cleanup failed")

	for _, test := range []struct {
		name     string
		runErr   error
		closeErr error
	}{
		{name: "success"},
		{name: "execution failure", runErr: runErr},
		{name: "cleanup failure", closeErr: closeErr},
		{name: "combined failure", runErr: runErr, closeErr: closeErr},
	} {
		t.Run(test.name, func(t *testing.T) {
			var closed []string
			engine, err := ferret.New(
				ferret.WithFunctionsRegistrar(func(ns runtime.Namespace) {
					if err := sdk.RegisterFunctions(ns, sdk.Func("RESULT", func(context.Context, ...runtime.Value) (runtime.Value, error) {
						return runtime.True, test.runErr
					})); err != nil {
						t.Fatal(err)
					}
				}),
				ferret.WithSessionCloseHook(func() error {
					closed = append(closed, "session")

					return test.closeErr
				}),
				ferret.WithPlanCloseHook(func() error {
					closed = append(closed, "plan")

					return nil
				}),
			)
			if err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() {
				if err := engine.Close(); err != nil {
					t.Error(err)
				}
			})
			var dst bytes.Buffer

			err = execQuery(t.Context(), engine, nil, ferret.NewSource("test", "RETURN RESULT()"), &dst)
			if test.runErr == nil && test.closeErr == nil {
				if err != nil || dst.String() != "true\n" {
					t.Fatalf("output = %q, error = %v", dst.String(), err)
				}
			} else if dst.Len() != 0 {
				t.Fatalf("failed query wrote %q", dst.String())
			}

			for _, cause := range []error{test.runErr, test.closeErr} {
				if cause != nil && !errors.Is(err, cause) {
					t.Fatalf("error %v does not preserve %v", err, cause)
				}
			}

			if strings.Join(closed, ",") != "session,plan" {
				t.Fatalf("unexpected cleanup order: %v", closed)
			}
		})
	}
}

func TestAnalyzeQuery(t *testing.T) {
	query := ferret.NewSource("test", "RETURN 42")
	var dst bytes.Buffer

	if err := analyzeQuery(t.Context(), &dst, query, compiler.Full); err != nil || dst.Len() == 0 {
		t.Fatalf("disassembly = %q, error = %v", dst.String(), err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	dst.Reset()

	if err := analyzeQuery(ctx, &dst, query, compiler.Full); !errors.Is(err, context.Canceled) || dst.Len() != 0 {
		t.Fatalf("canceled disassembly = %q, error = %v", dst.String(), err)
	}

	if err := analyzeQuery(t.Context(), &dst, query, compiler.OptimizationLevel(-1)); err == nil {
		t.Fatal("expected invalid optimization level")
	}
}
