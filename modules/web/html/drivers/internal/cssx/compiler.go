package cssx

import (
	"github.com/MontFerret/cssx"
)

func Compile(input string) (cssx.Pipeline, error) {
	return cssx.Compile(input)
}
