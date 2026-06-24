package core

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestRuntimeMapToClaims(t *testing.T) {
	ctx := context.Background()
	m := runtime.NewObjectWith(map[string]runtime.Value{
		"sub": runtime.NewString("user123"),
	})

	claims, err := runtimeMapToClaims(ctx, m)
	if err != nil {
		t.Fatalf("runtimeMapToClaims() error = %v", err)
	}

	if claims["sub"].(runtime.String).String() != "user123" {
		t.Errorf("claims[\"sub\"] = %v (%T), want \"user123\"", claims["sub"], claims["sub"])
	}
}

func TestClaimAsInt64(t *testing.T) {
	tests := []struct {
		in    any
		name  string
		want  int64
		wantB bool
	}{
		{name: "float64", in: 123.0, want: 123, wantB: true},
		{name: "int64", in: int64(456), want: 456, wantB: true},
		{name: "int", in: 789, want: 789, wantB: true},
		{name: "string", in: "123", want: 0, wantB: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotB := claimAsInt64(tt.in)
			if got != tt.want || gotB != tt.wantB {
				t.Errorf("claimAsInt64() = (%v, %v), want (%v, %v)", got, gotB, tt.want, tt.wantB)
			}
		})
	}
}

func TestAudienceMatches(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected string
		want     bool
	}{
		{"empty expected", "any", "", true},
		{"string match", "aud1", "aud1", true},
		{"string mismatch", "aud1", "aud2", false},
		{"array match", []any{"aud1", "aud2"}, "aud2", true},
		{"array mismatch", []any{"aud1", "aud2"}, "aud3", false},
		{"wrong type", 123, "aud1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := audienceMatches(tt.actual, tt.expected); got != tt.want {
				t.Errorf("audienceMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}
