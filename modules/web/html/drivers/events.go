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
