package lib

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MontFerret/contrib/modules/security/oauth2/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type clientValue struct {
	target   *core.Client
	identity uint64
}

func newClientValue(client *core.Client) *clientValue {
	return &clientValue{
		target:   client.Clone(),
		identity: nextHostValueIdentity(),
	}
}

func (v *clientValue) client() *core.Client {
	if v == nil || v.target == nil {
		return nil
	}

	return v.target.Clone()
}

func (v *clientValue) String() string {
	if v == nil || v.target == nil {
		return runtime.None.String()
	}

	return v.target.String()
}

func (v *clientValue) Format(state fmt.State, verb rune) {
	_, _ = fmt.Fprint(state, v.String())
}

func (v *clientValue) Hash() uint64 {
	if v == nil {
		return runtime.None.Hash()
	}

	return v.identity
}

func (v *clientValue) Copy() runtime.Value {
	if v == nil {
		return runtime.None
	}

	return &clientValue{
		target:   v.target,
		identity: v.identity,
	}
}

func (v *clientValue) Get(_ context.Context, key runtime.Value) (runtime.Value, error) {
	if v == nil || v.target == nil {
		return runtime.None, nil
	}

	switch safeKey(key) {
	case "provider":
		return newProviderValue(v.target.Provider), nil
	case "clientID":
		return runtime.NewString(v.target.ClientID), nil
	case "authMethod":
		return runtime.NewString(string(v.target.AuthMethod)), nil
	case "hasClientSecret":
		return runtime.NewBoolean(v.target.ClientSecret != ""), nil
	default:
		return runtime.None, nil
	}
}

func (v *clientValue) Unwrap() any {
	if v == nil || v.target == nil {
		return nil
	}

	return v.safeProjection()
}

func (v *clientValue) MarshalJSON() ([]byte, error) {
	if v == nil || v.target == nil {
		return []byte("null"), nil
	}

	return json.Marshal(v.safeProjection())
}

func (v *clientValue) safeProjection() map[string]any {
	return map[string]any{
		"type":              "oauth2.Client",
		"provider":          newProviderValue(v.target.Provider).safeProjection(),
		"client_id":         v.target.ClientID,
		"auth_method":       string(v.target.AuthMethod),
		"has_client_secret": v.target.ClientSecret != "",
	}
}
