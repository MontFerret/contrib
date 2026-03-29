package core

import (
	"context"
	"io"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// DecodeIterator iterates over normalized XML events.
type DecodeIterator struct {
	cursor   *decodeCursor
	eventNum runtime.Int
	done     bool
}

// NewDecodeIterator returns an iterator over normalized XML events.
func NewDecodeIterator(data runtime.String) (*DecodeIterator, error) {
	return &DecodeIterator{
		cursor: newDecodeCursor(data),
	}, nil
}

// Iterate returns the iterator itself.
func (d *DecodeIterator) Iterate(_ context.Context) (runtime.Iterator, error) {
	return d, nil
}

// Next returns the next XML event and its 1-based event number.
func (d *DecodeIterator) Next(_ context.Context) (runtime.Value, runtime.Value, error) {
	if d.done {
		return runtime.None, runtime.None, io.EOF
	}

	for {
		event, err := d.cursor.Next()
		if err != nil {
			d.done = true

			return runtime.None, runtime.None, err
		}

		switch event.kind {
		case decodeEventStart:
			d.eventNum++
			return newStartElementEvent(event.name, event.attrs), d.eventNum, nil
		case decodeEventEnd:
			d.eventNum++
			return newEndElementEvent(event.name), d.eventNum, nil
		case decodeEventText:
			d.eventNum++
			return newTextNode(event.text), d.eventNum, nil
		}
	}
}

// Close stops the iterator.
func (d *DecodeIterator) Close() error {
	d.done = true

	return nil
}
