package session

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/mafredri/cdp/protocol/target"
	"github.com/mafredri/cdp/rpcc"
)

func TestManagerCloseRetainsCleanupError(t *testing.T) {
	failure := errors.New("client cleanup failed")
	var calls atomic.Int32
	_, cancel := context.WithCancel(t.Context())
	manager := &Manager{
		cancel: cancel,
		clientsBySession: map[target.SessionID]*Client{"root": {
			closeFn: func() error {
				calls.Add(1)

				return failure
			},
		}},
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			if err := manager.Close(); !errors.Is(err, failure) {
				t.Errorf("Close lost cleanup error: %v", err)
			}
		})
	}

	wg.Wait()
	if calls.Load() != 1 {
		t.Fatalf("client closed %d times", calls.Load())
	}
}

func TestManagerCloseSkipsDetachOnClosedBrowser(t *testing.T) {
	local, remote := net.Pipe()
	defer remote.Close()
	conn, err := rpcc.DialContext(t.Context(), "", rpcc.WithDialer(func(context.Context, string) (io.ReadWriteCloser, error) {
		return local, nil
	}))
	if err != nil {
		t.Fatal(err)
	}

	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}

	var detached atomic.Bool
	manager := &Manager{
		cancel:      func() {},
		browserConn: conn,
		clientsBySession: map[target.SessionID]*Client{"root": {
			markDetachedFn: func() { detached.Store(true) },
			closeFn: func() error {
				if !detached.Load() {
					return errors.New("attempted remote detach on a closed browser")
				}

				return nil
			},
		}},
	}
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}

	if !detached.Load() {
		t.Fatal("client did not skip remote detach")
	}
}
