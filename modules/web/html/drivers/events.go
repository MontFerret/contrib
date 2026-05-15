package drivers

const (
	NavigationEvent = "navigation"
	RequestEvent    = "request"
	ResponseEvent   = "response"

	NetworkRequestStartedEvent   = "network.request_started"
	NetworkResponseReceivedEvent = "network.response_received"
	NetworkRequestFinishedEvent  = "network.request_finished"
	NetworkRequestFailedEvent    = "network.request_failed"
	NetworkIdleEvent             = "network.idle"

	DispatchClickEvent       = "click"
	DispatchDoubleClickEvent = "dblclick"
	DispatchMouseDownEvent   = "mousedown"
	DispatchMouseUpEvent     = "mouseup"
	DispatchMouseOverEvent   = "mouseover"
	DispatchMouseOutEvent    = "mouseout"
	DispatchMouseMoveEvent   = "mousemove"
	DispatchKeyDownEvent     = "keydown"
	DispatchKeyUpEvent       = "keyup"
	DispatchKeyPressEvent    = "keypress"
	DispatchPressEvent       = "press"
	DispatchTypeEvent        = "type"
	DispatchInputEvent       = "input"
	DispatchChangeEvent      = "change"
	DispatchSubmitEvent      = "submit"
	DispatchResetEvent       = "reset"
	DispatchFocusEvent       = "focus"
	DispatchBlurEvent        = "blur"
	DispatchCheckEvent       = "check"
	DispatchUncheckEvent     = "uncheck"
	DispatchToggleEvent      = "toggle"
	DispatchScrollEvent      = "scroll"
)

func IsNetworkEvent(name string) bool {
	switch name {
	case NetworkRequestStartedEvent,
		NetworkResponseReceivedEvent,
		NetworkRequestFinishedEvent,
		NetworkRequestFailedEvent,
		NetworkIdleEvent:
		return true
	default:
		return false
	}
}

func IsDispatchEvent(name string) bool {
	switch name {
	case DispatchClickEvent,
		DispatchDoubleClickEvent,
		DispatchMouseDownEvent,
		DispatchMouseUpEvent,
		DispatchMouseOverEvent,
		DispatchMouseOutEvent,
		DispatchMouseMoveEvent,
		DispatchKeyDownEvent,
		DispatchKeyUpEvent,
		DispatchKeyPressEvent,
		DispatchPressEvent,
		DispatchTypeEvent,
		DispatchInputEvent,
		DispatchChangeEvent,
		DispatchSubmitEvent,
		DispatchResetEvent,
		DispatchFocusEvent,
		DispatchBlurEvent,
		DispatchCheckEvent,
		DispatchUncheckEvent,
		DispatchToggleEvent,
		DispatchScrollEvent:
		return true
	default:
		return false
	}
}
