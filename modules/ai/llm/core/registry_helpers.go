package core

import (
	"reflect"
	"strings"
)

func validateFactory(factory ProviderFactory) (string, error) {
	if factory == nil || isNilFactory(factory) {
		return "", NewError(ErrInvalidOptions, "provider factory must not be nil")
	}

	name := strings.ToLower(strings.TrimSpace(factory.Name()))
	if name == "" {
		return "", NewError(ErrInvalidOptions, "provider name must not be blank")
	}

	return name, nil
}

func isNilFactory(factory ProviderFactory) bool {
	value := reflect.ValueOf(factory)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
