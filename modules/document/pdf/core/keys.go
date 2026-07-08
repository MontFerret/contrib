package core

import "github.com/MontFerret/ferret/v2/pkg/runtime"

func isEmptyKey(key runtime.Value) bool {
	return key == nil || key == runtime.None
}
