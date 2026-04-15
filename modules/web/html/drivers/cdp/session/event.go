package session

type (
	EventKind int

	Event struct {
		Client *Client
		Kind   EventKind
	}

	Listener func(Event)

	ListenerID int64
)

const (
	EventAttached EventKind = iota + 1
	EventDetached
)
