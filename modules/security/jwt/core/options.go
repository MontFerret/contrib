package core

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// VerifyOptions configures SECURITY::JWT::VERIFY.
type VerifyOptions struct {
	Algorithms []string `json:"algorithms"`
	Issuer     string   `json:"issuer"`
	Audience   string   `json:"audience"`
	Subject    string   `json:"subject"`
	Required   []string `json:"required"`
	Leeway     int64    `json:"leeway"`
	Now        int64    `json:"now"`
	MaxAge     int64    `json:"max_age"`
}

// SignOptions configures SECURITY::JWT::SIGN.
type SignOptions struct {
	Audience  any            `json:"audience"`
	Header    map[string]any `json:"header"`
	Algorithm string         `json:"algorithm"`
	Issuer    string         `json:"issuer"`
	Subject   string         `json:"subject"`
	ExpiresIn int64          `json:"expires_in"`
	NotBefore int64          `json:"not_before"`
	IssuedAt  bool           `json:"issued_at"`
}

// DecodeVerifyOptions decodes VERIFY options from a Ferret map.
func DecodeVerifyOptions(ctx context.Context, value runtime.Map) (VerifyOptions, error) {
	var opts VerifyOptions
	if err := sdk.Decode(ctx, value, &opts, sdk.DisallowUnknownFields()); err != nil {
		return VerifyOptions{}, wrapError(ErrInvalidToken, "invalid verify options", err)
	}

	return opts, nil
}

// DecodeSignOptions decodes SIGN options from a Ferret map.
func DecodeSignOptions(ctx context.Context, value runtime.Map) (SignOptions, error) {
	var opts SignOptions
	if err := sdk.Decode(ctx, value, &opts, sdk.DisallowUnknownFields()); err != nil {
		return SignOptions{}, wrapError(ErrInvalidToken, "invalid sign options", err)
	}

	return opts, nil
}

func (o VerifyOptions) validate() error {
	if len(o.Algorithms) == 0 {
		return newError(ErrInvalidToken, "algorithms option is required")
	}

	for _, alg := range o.Algorithms {
		if alg == "none" {
			return newError(ErrUnexpectedAlgorithm, "alg none is not allowed")
		}
	}

	return nil
}

func (o SignOptions) validate() error {
	if o.Algorithm == "" {
		return newError(ErrInvalidToken, "algorithm option is required")
	}

	if o.Algorithm == "none" {
		return newError(ErrUnexpectedAlgorithm, "alg none is not allowed")
	}

	return nil
}
