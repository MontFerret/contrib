package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MontFerret/contrib/modules/security/oauth2/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type tokenValue struct {
	target   *core.TokenSet
	clock    func() time.Time
	identity uint64
}

func newTokenValue(token *core.TokenSet) *tokenValue {
	return newTokenValueWithClock(token, time.Now)
}

func newTokenValueWithClock(token *core.TokenSet, clock func() time.Time) *tokenValue {
	return &tokenValue{
		target:   token.Clone(),
		clock:    clock,
		identity: nextHostValueIdentity(),
	}
}

func (v *tokenValue) token() *core.TokenSet {
	if v == nil || v.target == nil {
		return nil
	}

	return v.target.Clone()
}

func (v *tokenValue) now() time.Time {
	if v == nil || v.clock == nil {
		return time.Now()
	}

	return v.clock()
}

func (v *tokenValue) String() string {
	if v == nil || v.target == nil {
		return runtime.None.String()
	}

	return v.target.String()
}

func (v *tokenValue) Format(state fmt.State, verb rune) {
	_, _ = fmt.Fprint(state, v.String())
}

func (v *tokenValue) Hash() uint64 {
	if v == nil {
		return runtime.None.Hash()
	}

	return v.identity
}

func (v *tokenValue) Copy() runtime.Value {
	if v == nil {
		return runtime.None
	}

	return &tokenValue{
		target:   v.target,
		clock:    v.clock,
		identity: v.identity,
	}
}

func (v *tokenValue) Get(_ context.Context, key runtime.Value) (runtime.Value, error) {
	if v == nil || v.target == nil {
		return runtime.None, nil
	}

	switch safeKey(key) {
	case "tokenType":
		return runtime.NewString(v.target.TokenType), nil
	case "scopes":
		return stringArray(v.target.Scopes()), nil
	case "expiresAt":
		if v.target.ExpiresAt.IsZero() {
			return runtime.None, nil
		}

		return runtime.NewDateTime(v.target.ExpiresAt), nil
	case "expired":
		return runtime.NewBoolean(v.target.Expired(v.now(), 0)), nil
	case "validFor":
		remaining, known := v.target.ValidFor(v.now())
		if !known {
			return runtime.None, nil
		}

		return runtime.NewInt64(ceilMilliseconds(remaining)), nil
	case "hasRefreshToken":
		return runtime.NewBoolean(v.target.RefreshToken != ""), nil
	case "hasIDToken":
		return runtime.NewBoolean(v.target.IDToken != ""), nil
	default:
		return runtime.None, nil
	}
}

func (v *tokenValue) Unwrap() any {
	if v == nil || v.target == nil {
		return nil
	}

	return v.safeProjection()
}

func (v *tokenValue) MarshalJSON() ([]byte, error) {
	if v == nil || v.target == nil {
		return []byte("null"), nil
	}

	return json.Marshal(v.safeProjection())
}

func (v *tokenValue) safeProjection() map[string]any {
	projection := map[string]any{
		"type":       "oauth2.TokenSet",
		"token_type": v.target.TokenType,
	}

	if v.target.AccessToken != "" {
		projection["access_token"] = "<redacted>"
	}

	if v.target.RefreshToken != "" {
		projection["refresh_token"] = "<redacted>"
	}

	if v.target.IDToken != "" {
		projection["id_token"] = "<redacted>"
	}

	if v.target.Scope != "" {
		projection["scope"] = v.target.Scope
	}

	if !v.target.ExpiresAt.IsZero() {
		projection["expires_at"] = v.target.ExpiresAt.UTC().Format(time.RFC3339Nano)
	}

	return projection
}
