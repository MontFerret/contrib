package lazy

import (
	"context"
	"sync"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type (
	ValueFactory func(ctx context.Context) (runtime.Value, error)

	Value struct {
		value   runtime.Value
		err     error
		factory ValueFactory
		mu      sync.Mutex
		ready   bool
	}
)

func New(factory ValueFactory) *Value {
	return &Value{
		value:   runtime.None,
		factory: factory,
	}
}

func (lv *Value) Ready() bool {
	lv.mu.Lock()
	defer lv.mu.Unlock()

	return lv.ready
}

func (lv *Value) Read(ctx context.Context) (runtime.Value, error) {
	lv.mu.Lock()
	defer lv.mu.Unlock()

	if !lv.ready {
		lv.load(ctx)
	}

	return lv.value, lv.err
}

func (lv *Value) Mutate(ctx context.Context, mutator func(v runtime.Value, err error)) {
	lv.mu.Lock()
	defer lv.mu.Unlock()

	if !lv.ready {
		lv.load(ctx)
	}

	mutator(lv.value, lv.err)
}

func (lv *Value) MutateIfReady(mutator func(v runtime.Value, err error)) {
	lv.mu.Lock()
	defer lv.mu.Unlock()

	if lv.ready {
		mutator(lv.value, lv.err)
	}
}

func (lv *Value) Reload(ctx context.Context) {
	lv.mu.Lock()
	defer lv.mu.Unlock()

	lv.reset()
	lv.load(ctx)
}

func (lv *Value) Reset() {
	lv.mu.Lock()
	defer lv.mu.Unlock()

	lv.reset()
}

func (lv *Value) reset() {
	lv.ready = false
	lv.value = runtime.None
	lv.err = nil
}

func (lv *Value) load(ctx context.Context) {
	val, err := lv.factory(ctx)
	if err == nil {
		lv.value = val
		lv.err = nil
	} else {
		lv.value = runtime.None
		lv.err = err
	}

	lv.ready = true
}
