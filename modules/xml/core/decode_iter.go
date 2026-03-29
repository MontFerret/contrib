package core

import (
	"context"
	"encoding/xml"
	"io"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// DecodeIterator iterates over normalized XML events.
type DecodeIterator struct {
	decoder  *xml.Decoder
	stack    []string
	eventNum runtime.Int
	rootSeen bool
	done     bool
}

// NewDecodeIterator returns an iterator over normalized XML events.
func NewDecodeIterator(data runtime.String) (*DecodeIterator, error) {
	return &DecodeIterator{
		decoder: xml.NewDecoder(strings.NewReader(data.String())),
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
		token, err := d.decoder.RawToken()
		if err != nil {
			d.done = true

			if err == io.EOF {
				if len(d.stack) > 0 {
					return runtime.None, runtime.None, newXMLErrorf("unexpected EOF while closing %q", d.stack[len(d.stack)-1])
				}

				if !d.rootSeen {
					return runtime.None, runtime.None, newXMLError("document has no root element")
				}

				return runtime.None, runtime.None, io.EOF
			}

			return runtime.None, runtime.None, wrapXMLError(err, "failed to decode XML data")
		}

		switch typed := token.(type) {
		case xml.StartElement:
			if d.rootSeen && len(d.stack) == 0 {
				d.done = true
				return runtime.None, runtime.None, newXMLError("multiple root elements are not supported")
			}

			name := xmlNameToString(typed.Name)
			d.stack = append(d.stack, name)
			d.rootSeen = true
			d.eventNum++

			return newStartElementEvent(name, newAttrsObject(typed.Attr)), d.eventNum, nil
		case xml.EndElement:
			name := xmlNameToString(typed.Name)
			if len(d.stack) == 0 {
				d.done = true
				return runtime.None, runtime.None, newXMLErrorf("unexpected closing tag %q", name)
			}

			expected := d.stack[len(d.stack)-1]
			if expected != name {
				d.done = true
				return runtime.None, runtime.None, newXMLErrorf("mismatched closing tag: expected %q but got %q", expected, name)
			}

			d.stack = d.stack[:len(d.stack)-1]
			d.eventNum++

			return newEndElementEvent(name), d.eventNum, nil
		case xml.CharData:
			text := string(typed)
			if len(d.stack) == 0 {
				if strings.TrimSpace(text) == "" {
					continue
				}

				d.done = true
				return runtime.None, runtime.None, newXMLError("text outside root element is not supported")
			}

			d.eventNum++

			return newTextNode(text), d.eventNum, nil
		case xml.Comment, xml.Directive, xml.ProcInst:
			continue
		}
	}
}

// Close stops the iterator.
func (d *DecodeIterator) Close() error {
	d.done = true

	return nil
}
