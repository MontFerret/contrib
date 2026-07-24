package core

import (
	"fmt"
	"reflect"
)

// ExecutorOption configures an Executor.
type ExecutorOption func(*Executor) error

// WithClock sets the clock used to calculate token expiry.
func WithClock(clock Clock) ExecutorOption {
	return func(executor *Executor) error {
		if tokenDependencyIsNil(clock) {
			return fmt.Errorf("clock is required")
		}

		executor.clock = clock

		return nil
	}
}

func tokenDependencyIsNil(value any) bool {
	if value == nil {
		return true
	}

	reflected := reflect.ValueOf(value)

	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
