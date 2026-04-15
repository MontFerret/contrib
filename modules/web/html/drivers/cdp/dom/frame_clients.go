package dom

import (
	"sync"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
)

type AtomicFrameClientCollection struct {
	value map[page.FrameID]*cdp.Client
	mu    sync.Mutex
}

func NewAtomicFrameClientCollection() *AtomicFrameClientCollection {
	return &AtomicFrameClientCollection{
		value: make(map[page.FrameID]*cdp.Client),
	}
}

func (fc *AtomicFrameClientCollection) Get(key page.FrameID) (*cdp.Client, bool) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	client, ok := fc.value[key]

	return client, ok
}

func (fc *AtomicFrameClientCollection) Set(key page.FrameID, client *cdp.Client) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	fc.value[key] = client
}

func (fc *AtomicFrameClientCollection) Remove(key page.FrameID) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	delete(fc.value, key)
}

func (fc *AtomicFrameClientCollection) Retain(ids map[page.FrameID]struct{}) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	for id := range fc.value {
		if _, ok := ids[id]; ok {
			continue
		}

		delete(fc.value, id)
	}
}
