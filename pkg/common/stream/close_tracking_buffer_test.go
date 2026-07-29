package stream

import "bytes"

type closeTrackingBuffer struct {
	bytes.Buffer
	closed bool
}

func (b *closeTrackingBuffer) Close() error {
	b.closed = true

	return nil
}
