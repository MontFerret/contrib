package session

type (
	EventKind int

	Event struct {
		Client    *Client
		Kind      EventKind
		Err       error
		Operation string
	}

	Listener func(Event)

	ListenerID int64
)

const (
	EventAttached EventKind = iota + 1
	EventDetached
	EventFailed
)
