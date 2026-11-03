package network

import (
	"context"
	"sync/atomic"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type subscriptionTestStream struct {
	messages chan runtime.Message
	closes   atomic.Int32
}

func newSubscriptionTestStream() *subscriptionTestStream {
	return &subscriptionTestStream{messages: make(chan runtime.Message, 4)}
}

func (s *subscriptionTestStream) Read(context.Context) <-chan runtime.Message {
	return s.messages
}

func (s *subscriptionTestStream) Close() error {
	if s.closes.Add(1) == 1 {
		close(s.messages)
	}

	return nil
}
