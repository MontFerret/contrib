package core

import (
	"encoding/xml"
	"io"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type decodeEventKind uint8

const (
	decodeEventStart decodeEventKind = iota + 1
	decodeEventEnd
	decodeEventText
)

type decodeEvent struct {
	attrs *runtime.Object
	name  string
	text  string
	kind  decodeEventKind
}

type decodeCursor struct {
	decoder  *xml.Decoder
	stack    []string
	rootSeen bool
}

func newDecodeCursor(data runtime.String) *decodeCursor {
	return &decodeCursor{
		decoder: xml.NewDecoder(strings.NewReader(data.String())),
	}
}

func (c *decodeCursor) Next() (decodeEvent, error) {
	for {
		token, err := c.decoder.RawToken()
		if err != nil {
			if err == io.EOF {
				if len(c.stack) > 0 {
					return decodeEvent{}, newXMLErrorf("unexpected EOF while closing %q", c.stack[len(c.stack)-1])
				}

				if !c.rootSeen {
					return decodeEvent{}, newXMLError("document has no root element")
				}

				return decodeEvent{}, io.EOF
			}

			return decodeEvent{}, wrapXMLError(err, "failed to decode XML data")
		}

		switch typed := token.(type) {
		case xml.StartElement:
			if c.rootSeen && len(c.stack) == 0 {
				return decodeEvent{}, newXMLError("multiple root elements are not supported")
			}

			name := xmlNameToString(typed.Name)
			c.stack = append(c.stack, name)
			c.rootSeen = true

			return decodeEvent{
				kind:  decodeEventStart,
				name:  name,
				attrs: newAttrsObject(typed.Attr),
			}, nil
		case xml.EndElement:
			name := xmlNameToString(typed.Name)
			if len(c.stack) == 0 {
				return decodeEvent{}, newXMLErrorf("unexpected closing tag %q", name)
			}

			expected := c.stack[len(c.stack)-1]
			if expected != name {
				return decodeEvent{}, newXMLErrorf("mismatched closing tag: expected %q but got %q", expected, name)
			}

			c.stack = c.stack[:len(c.stack)-1]

			return decodeEvent{
				kind: decodeEventEnd,
				name: name,
			}, nil
		case xml.CharData:
			text := string(typed)
			if len(c.stack) == 0 {
				if strings.TrimSpace(text) == "" {
					continue
				}

				return decodeEvent{}, newXMLError("text outside root element is not supported")
			}

			return decodeEvent{
				kind: decodeEventText,
				text: text,
			}, nil
		case xml.Comment, xml.Directive, xml.ProcInst:
			continue
		}
	}
}
