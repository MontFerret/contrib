package input

import (
	"context"

	"github.com/mafredri/cdp"
	cdpinput "github.com/mafredri/cdp/protocol/input"
)

type mouseDispatchInput struct {
	cdp.Input
	err   error
	calls []*cdpinput.DispatchMouseEventArgs
}

func newMouseDispatchInput(err error) *mouseDispatchInput {
	return &mouseDispatchInput{err: err}
}

func (i *mouseDispatchInput) DispatchMouseEvent(_ context.Context, args *cdpinput.DispatchMouseEventArgs) error {
	i.calls = append(i.calls, args)

	return i.err
}
