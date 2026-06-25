package core

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDSNFromURI(t *testing.T) {
	t.Parallel()

	dsn, err := OpenOptions{URI: stringPtr(" postgres://ferret:secret@localhost:5432/ferret?sslmode=disable ")}.dsn()
	if err != nil {
		t.Fatalf("unexpected dsn error: %v", err)
	}

	if dsn != "postgres://ferret:secret@localhost:5432/ferret?sslmode=disable" {
		t.Fatalf("unexpected dsn: %q", dsn)
	}
}

func TestDSNFromStructuredFields(t *testing.T) {
	t.Parallel()

	dsn, err := OpenOptions{
		Host:     stringPtr("localhost"),
		Port:     intPtr(15432),
		Database: stringPtr("ferret"),
		User:     stringPtr("ferret"),
		Password: stringPtr("s e/c:r@e?t"),
		SSLMode:  stringPtr("disable"),
	}.dsn()
	if err != nil {
		t.Fatalf("unexpected dsn error: %v", err)
	}

	expected := "postgres://ferret:s%20e%2Fc%3Ar%40e%3Ft@localhost:15432/ferret?sslmode=disable"
	if dsn != expected {
		t.Fatalf("unexpected dsn: got %q, want %q", dsn, expected)
	}
}

func TestDSNFromStructuredFieldsDefaultsPort(t *testing.T) {
	t.Parallel()

	dsn, err := OpenOptions{
		Host:     stringPtr("localhost"),
		Database: stringPtr("ferret"),
		User:     stringPtr("ferret"),
	}.dsn()
	if err != nil {
		t.Fatalf("unexpected dsn error: %v", err)
	}

	expected := "postgres://ferret@localhost:5432/ferret"
	if dsn != expected {
		t.Fatalf("unexpected dsn: got %q, want %q", dsn, expected)
	}
}

func TestOpenValidation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		options OpenOptions
		name    string
		want    string
	}{
		{
			name: "missing source",
			want: "exactly one of uri or structured connection fields must be provided",
		},
		{
			name: "uri and fields",
			options: OpenOptions{
				URI:  stringPtr("postgres://localhost/db"),
				Host: stringPtr("localhost"),
			},
			want: "exactly one of uri or structured connection fields must be provided",
		},
		{
			name: "missing structured required field",
			options: OpenOptions{
				Host: stringPtr("localhost"),
				User: stringPtr("ferret"),
			},
			want: "host, database, and user are required when uri is not provided",
		},
		{
			name: "invalid port",
			options: OpenOptions{
				Host:     stringPtr("localhost"),
				Port:     intPtr(0),
				Database: stringPtr("ferret"),
				User:     stringPtr("ferret"),
			},
			want: "port must be greater than 0",
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := tt.options.dsn()
			assertErrorContains(t, err, tt.want)
		})
	}
}

func TestDecodeOpenOptions(t *testing.T) {
	t.Parallel()

	opts, err := DecodeOpenOptions(runtime.NewObjectWith(map[string]runtime.Value{
		"host":     runtime.NewString("localhost"),
		"port":     runtime.NewInt(15432),
		"database": runtime.NewString("ferret"),
		"user":     runtime.NewString("ferret"),
		"password": runtime.NewString("secret"),
		"sslMode":  runtime.NewString("disable"),
	}))
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}

	dsn, err := opts.dsn()
	if err != nil {
		t.Fatalf("unexpected dsn error: %v", err)
	}

	expected := "postgres://ferret:secret@localhost:15432/ferret?sslmode=disable"
	if dsn != expected {
		t.Fatalf("unexpected dsn: got %q, want %q", dsn, expected)
	}

	_, err = DecodeOpenOptions(runtime.NewString("invalid"))
	assertErrorContains(t, err, "DB::POSTGRES OPEN failed")
}

func TestOpenWrapsValidationErrors(t *testing.T) {
	t.Parallel()

	_, err := Open(context.Background(), OpenOptions{})
	assertErrorContains(t, err, "DB::POSTGRES OPEN failed")
	assertErrorContains(t, err, "exactly one of uri or structured connection fields must be provided")
}
