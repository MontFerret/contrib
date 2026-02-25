package input

import (
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func randomDuration(delay int) time.Duration {
	return time.Duration(runtime.Random2(float64(delay)))
}
