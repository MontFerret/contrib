package network

import (
	"context"
	"errors"
	"time"

	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/rpcc"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func sendNetworkMessage(
	ctx context.Context,
	done <-chan struct{},
	out chan<- runtime.Message,
	message runtime.Message,
) bool {
	select {
	case <-ctx.Done():
		return false
	case <-done:
		return false
	case out <- message:
		return true
	}
}

func stopIdleTimer(timer *time.Timer) {
	if timer == nil {
		return
	}

	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}

func closeNetworkStream(stream rpcc.Stream) error {
	if stream == nil {
		return nil
	}

	return stream.Close()
}

func frameIDString(frameID *page.FrameID) string {
	if frameID == nil {
		return ""
	}

	return string(*frameID)
}

func boolPtrValue(value *bool) bool {
	return value != nil && *value
}

// Reserve one terminal message when the page goes away. No producer is left
// when this runs, so an abandoned reader cannot block subscription cleanup.
func closePageMessages(pageCtx context.Context, out chan runtime.Message) {
	if err := context.Cause(pageCtx); errors.Is(err, rpcc.ErrConnClosing) {
		select {
		case <-out:
		default:
		}

		out <- runtime.NewErrorMessage(err)
	}

	close(out)
}
