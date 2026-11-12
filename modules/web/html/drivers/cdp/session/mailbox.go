package session

import "sync"

// A CDP reader must never wait for another session to consume its messages.
type mailbox[T any] struct {
	mu    sync.Mutex
	ready *sync.Cond
	items []T
	err   error
}

func newMailbox[T any]() *mailbox[T] {
	q := new(mailbox[T])
	q.ready = sync.NewCond(&q.mu)

	return q
}

func (q *mailbox[T]) put(item T) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.err == nil {
		q.items = append(q.items, item)
		q.ready.Signal()
	}
}

func (q *mailbox[T]) next() (T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.items) == 0 && q.err == nil {
		q.ready.Wait()
	}

	var zero T

	if len(q.items) == 0 {
		return zero, q.err
	}

	item := q.items[0]
	q.items[0] = zero
	q.items = q.items[1:]

	return item, nil
}

func (q *mailbox[T]) stop(err error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.err == nil {
		q.err = err
		q.ready.Broadcast()
	}
}

func (q *mailbox[T]) clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.items = nil
}
