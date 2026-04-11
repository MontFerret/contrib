package dom

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	jsoncodec "github.com/MontFerret/ferret/v2/pkg/encoding/json"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"
	"github.com/mafredri/cdp/rpcc"
)

func closeDOMEventResources(
	api domBindingRuntime,
	bindingName string,
	detach domEventSubscription,
	stream cdpruntime.BindingCalledClient,
	initial error,
) error {
	var errs []error

	if initial != nil {
		errs = append(errs, initial)
	}

	closeCtx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(drivers.DefaultWaitTimeout)*time.Millisecond,
	)
	defer cancel()

	if detach != nil {
		if err := detach(closeCtx, bindingName); err != nil && !isIgnorableDOMEventCloseError(err) {
			errs = append(errs, err)
		}
	}

	if err := api.RemoveBinding(closeCtx, cdpruntime.NewRemoveBindingArgs(bindingName)); err != nil && !isIgnorableDOMEventCloseError(err) {
		errs = append(errs, err)
	}

	if stream != nil {
		if err := stream.Close(); err != nil && !isIgnorableDOMEventCloseError(err) {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

func decodeDOMEventPayload(payload string) (runtime.Value, error) {
	return jsoncodec.Default.Decode([]byte(payload))
}

func newDOMBindingName() (string, error) {
	buf := make([]byte, 8)

	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return "__ferret_dom_event_" + hex.EncodeToString(buf), nil
}

func isIgnorableDOMEventCloseError(err error) bool {
	if err == nil {
		return false
	}

	switch {
	case errors.Is(err, context.Canceled):
		return true
	case errors.Is(err, context.DeadlineExceeded):
		return true
	case errors.Is(err, io.EOF):
		return true
	case errors.Is(err, rpcc.ErrConnClosing):
		return true
	}

	msg := strings.ToLower(err.Error())

	for _, fragment := range []string{
		"execution context was destroyed",
		"cannot find context with specified id",
		"cannot find object with given id",
		"cannot find object with id",
		"inspected target navigated or closed",
		"session closed",
		"target closed",
		"use of closed network connection",
	} {
		if strings.Contains(msg, fragment) {
			return true
		}
	}

	return false
}
