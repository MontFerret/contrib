package lib

import (
	"slices"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestRegisterLibRegistersCompleteNamespace(t *testing.T) {
	library := runtime.NewLibrary()
	if err := RegisterLib(library.Namespace("SECURITY").Namespace("OAUTH2")); err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}

	functions, err := library.Build()
	if err != nil {
		t.Fatalf("unexpected library build error: %v", err)
	}

	expected := []string{
		"SECURITY::OAUTH2::ACCESS_TOKEN",
		"SECURITY::OAUTH2::AUTH_HEADER",
		"SECURITY::OAUTH2::CLIENT",
		"SECURITY::OAUTH2::CLIENT_CREDENTIALS",
		"SECURITY::OAUTH2::DISCOVER",
		"SECURITY::OAUTH2::EXPIRED",
		"SECURITY::OAUTH2::EXPIRES_AT",
		"SECURITY::OAUTH2::ID_TOKEN",
		"SECURITY::OAUTH2::PROVIDER",
		"SECURITY::OAUTH2::REFRESH",
		"SECURITY::OAUTH2::REFRESH_TOKEN",
		"SECURITY::OAUTH2::SCOPES",
		"SECURITY::OAUTH2::TOKEN",
		"SECURITY::OAUTH2::TOKEN_TYPE",
		"SECURITY::OAUTH2::VALID_FOR",
	}

	actual := functions.List()
	slices.Sort(actual)
	slices.Sort(expected)

	if functions.Size() != len(expected) {
		t.Fatalf("registered %d functions, want %d", functions.Size(), len(expected))
	}
	if !slices.Equal(actual, expected) {
		t.Fatalf("registered names = %v, want %v", actual, expected)
	}
}

func TestRegisterLibIsAtomicOnDuplicate(t *testing.T) {
	library := runtime.NewLibrary()
	namespace := library.Namespace("SECURITY").Namespace("OAUTH2")

	namespace.Function().Var().Add("TOKEN", Token)
	if err := RegisterLib(namespace); err == nil {
		t.Fatal("expected duplicate registration error")
	}

	if library.Size() != 1 {
		t.Fatalf("registration was not atomic: size = %d, want 1", library.Size())
	}
}
