package lib

import (
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestExactArityFunctionsUseFixedRegistries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		register  func(runtime.Namespace) error
		name      string
		namespace string
	}{
		{name: "module", namespace: "WEB::HTML", register: RegisterLib},
		{name: "legacy", register: RegisterLibLegacy},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ns := runtime.NewNamespace(tc.namespace)
			if err := tc.register(ns); err != nil {
				t.Fatalf("register functions: %v", err)
			}

			definitions := ns.Function()
			assertFixedArity(t, definitions.A1(), definitions.Var(), "DOWNLOAD")
			assertFixedArity(
				t,
				definitions.A2(),
				definitions.Var(),
				"COOKIE_GET",
				"ELEMENT",
				"ELEMENT_EXISTS",
				"ELEMENTS",
				"ELEMENTS_COUNT",
				"INNER_HTML_ALL",
				"INNER_TEXT_ALL",
				"PAGINATION",
			)
			assertFixedArity(t, definitions.A3(), definitions.Var(), "MOUSE")
		})
	}
}

func assertFixedArity[T runtime.FunctionConstraint](
	t *testing.T,
	fixed runtime.FnDef[T],
	variadic runtime.FnDef[runtime.Function],
	names ...string,
) {
	t.Helper()

	for _, name := range names {
		if !fixed.Has(name) {
			t.Errorf("%s is not registered with fixed arity", name)
		}
		if variadic.Has(name) {
			t.Errorf("%s remains registered as variadic", name)
		}
	}
}
