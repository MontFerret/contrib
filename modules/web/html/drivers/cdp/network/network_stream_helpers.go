package network

import (
	"context"
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
