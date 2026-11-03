package network

import (
	"sort"
	"sync"

	"github.com/mafredri/cdp/protocol/target"

	cdpsession "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/session"
)

type subscriptionTestSessions struct {
	clients       map[target.SessionID]*cdpsession.Client
	listeners     map[cdpsession.ListenerID]cdpsession.Listener
	next          cdpsession.ListenerID
	afterSnapshot func()
	mu            sync.Mutex
}

func newSubscriptionTestSessions(clients ...*cdpsession.Client) *subscriptionTestSessions {
	s := &subscriptionTestSessions{
		clients:   make(map[target.SessionID]*cdpsession.Client),
		listeners: make(map[cdpsession.ListenerID]cdpsession.Listener),
	}

	for _, client := range clients {
		s.clients[client.ID] = client
	}

	return s
}

func (s *subscriptionTestSessions) Snapshot() []*cdpsession.Client {
	s.mu.Lock()

	clients := make([]*cdpsession.Client, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}

	sort.Slice(clients, func(i, j int) bool { return clients[i].ID < clients[j].ID })
	after := s.afterSnapshot
	s.afterSnapshot = nil
	s.mu.Unlock()
	if after != nil {
		after()
	}

	return clients
}

func (s *subscriptionTestSessions) AddListener(listener cdpsession.Listener) cdpsession.ListenerID {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	s.listeners[s.next] = listener

	return s.next
}

func (s *subscriptionTestSessions) RemoveListener(id cdpsession.ListenerID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.listeners, id)
}

func (s *subscriptionTestSessions) emit(kind cdpsession.EventKind, client *cdpsession.Client) {
	s.mu.Lock()
	if kind == cdpsession.EventAttached {
		s.clients[client.ID] = client
	} else {
		delete(s.clients, client.ID)
	}

	listeners := make([]cdpsession.Listener, 0, len(s.listeners))
	for _, listener := range s.listeners {
		listeners = append(listeners, listener)
	}

	s.mu.Unlock()

	for _, listener := range listeners {
		listener(cdpsession.Event{Kind: kind, Client: client})
	}
}

func (s *subscriptionTestSessions) listenerCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.listeners)
}
