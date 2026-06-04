package core

import "sync/atomic"

var nextResourceID atomic.Uint64

func newResourceID() uint64 {
	return nextResourceID.Add(1)
}
