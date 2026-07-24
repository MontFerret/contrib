package core

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

const redactedValue = "<redacted>"

type (
	// TokenSet represents a token endpoint response.
	TokenSet struct {
		ExpiresAt    time.Time
		Extra        map[string]any
		AccessToken  string
		TokenType    string
		RefreshToken string
		Scope        string
		IDToken      string
		ExpiresIn    time.Duration
	}

	safeTokenSet struct {
		Type         string `json:"type" msgpack:"type"`
		AccessToken  string `json:"access_token,omitempty" msgpack:"access_token,omitempty"`
		RefreshToken string `json:"refresh_token,omitempty" msgpack:"refresh_token,omitempty"`
		IDToken      string `json:"id_token,omitempty" msgpack:"id_token,omitempty"`
		TokenType    string `json:"token_type,omitempty" msgpack:"token_type,omitempty"`
		Scope        string `json:"scope,omitempty" msgpack:"scope,omitempty"`
		ExpiresAt    string `json:"expires_at,omitempty" msgpack:"expires_at,omitempty"`
	}
)

// Clone returns an independent copy of the token set.
func (t *TokenSet) Clone() *TokenSet {
	if t == nil {
		return nil
	}

	cloned := *t
	cloned.Extra = cloneTokenExtra(t.Extra)

	return &cloned
}

// Scopes returns the response scope split on ASCII spaces.
func (t TokenSet) Scopes() []string {
	if t.Scope == "" {
		return nil
	}

	parts := strings.Split(t.Scope, " ")
	scopes := make([]string, 0, len(parts))

	for _, part := range parts {
		if part != "" {
			scopes = append(scopes, part)
		}
	}

	return scopes
}

// Expired reports whether the token is expired at now after applying skew.
// A token with no known expiry is not considered expired.
func (t TokenSet) Expired(now time.Time, skew time.Duration) bool {
	if t.ExpiresAt.IsZero() {
		return false
	}
	if skew < 0 {
		skew = 0
	}

	return !t.ExpiresAt.After(now.Add(skew))
}

// ValidFor returns the remaining lifetime and whether expiry is known.
func (t TokenSet) ValidFor(now time.Time) (time.Duration, bool) {
	if t.ExpiresAt.IsZero() {
		return 0, false
	}

	remaining := t.ExpiresAt.Sub(now)
	if remaining < 0 {
		remaining = 0
	}

	return remaining, true
}

// String returns a secret-free representation.
func (t TokenSet) String() string {
	value, err := t.MarshalJSON()
	if err != nil {
		return `{"type":"oauth2.TokenSet"}`
	}

	return string(value)
}

// Format implements fmt.Formatter without exposing token material.
func (t TokenSet) Format(state fmt.State, verb rune) {
	value := t.String()
	if verb == 'q' {
		value = fmt.Sprintf("%q", value)
	}

	_, _ = io.WriteString(state, value)
}

// MarshalJSON serializes a redacted, stable token view.
func (t TokenSet) MarshalJSON() ([]byte, error) {
	return marshalSafeJSON(t.safeProjection())
}

// MarshalMsgpack serializes a redacted token view.
func (t TokenSet) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal(t.safeProjection())
}

func (t TokenSet) safeProjection() safeTokenSet {
	safe := safeTokenSet{
		Type:      "oauth2.TokenSet",
		TokenType: t.TokenType,
		Scope:     t.Scope,
	}

	if t.AccessToken != "" {
		safe.AccessToken = redactedValue
	}
	if t.RefreshToken != "" {
		safe.RefreshToken = redactedValue
	}
	if t.IDToken != "" {
		safe.IDToken = redactedValue
	}
	if !t.ExpiresAt.IsZero() {
		safe.ExpiresAt = t.ExpiresAt.UTC().Format(time.RFC3339Nano)
	}

	return safe
}
