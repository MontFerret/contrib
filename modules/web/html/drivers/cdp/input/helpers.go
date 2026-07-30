package input

import (
	"math/rand/v2"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const (
	minRndMouseDistance = 8.0
	maxRndMouseDistance = 24.0
)

func randomDuration(delay int) time.Duration {
	return time.Duration(runtime.Random2(float64(delay)))
}

func randomMouseDistance() float64 {
	return minRndMouseDistance + rand.Float64()*(maxRndMouseDistance-minRndMouseDistance)
}
