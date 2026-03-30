package core

import (
	"context"
	"errors"
	"io"
	"strconv"
	"strings"

	xmlcore "github.com/MontFerret/contrib/modules/xml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const (
	xmlEventStartElement = "startElement"
	xmlEventEndElement   = "endElement"
	xmlEventText         = "text"
)

type frameKind uint8

const (
	frameRootURLSet frameKind = iota + 1
	frameRootSitemapIndex
	frameURLEntry
	frameSitemapEntry
	frameFieldLoc
	frameFieldLastMod
	frameFieldChangeFreq
	frameFieldPriority
	frameIgnored
)

type frame struct {
	url     URLEntry
	sitemap SitemapRef
	name    string
	text    strings.Builder
	kind    frameKind
}

// Parse decodes a sitemap XML document by interpreting xml/core events.
func Parse(reader io.Reader, source string) (doc Document, retErr error) {
	iter, err := xmlcore.NewDecodeIteratorFromReader(reader)
	if err != nil {
		return Document{}, wrapError(source, StageParse, err, "failed to initialize XML decoder")
	}

	return parseIterator(context.Background(), iter, source)
}

func parseIterator(ctx context.Context, iter runtime.Iterator, source string) (doc Document, retErr error) {
	if closer, ok := iter.(io.Closer); ok {
		defer func() {
			if err := closer.Close(); err != nil && retErr == nil {
				retErr = wrapError(source, StageParse, err, "failed to close XML iterator")
			}
		}()
	}

	stack := make([]frame, 0, 8)

	for {
		event, _, err := iter.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return Document{}, wrapError(source, StageParse, err, "failed to decode sitemap XML")
		}

		if err := applyEvent(ctx, &doc, &stack, event, source); err != nil {
			return Document{}, err
		}
	}

	if doc.Type == "" {
		return Document{}, newError(source, StageParse, "document has no root element")
	}

	if len(stack) != 0 {
		return Document{}, newError(source, StageParse, "unexpected EOF while parsing sitemap")
	}

	return doc, nil
}

func applyEvent(ctx context.Context, doc *Document, stack *[]frame, event runtime.Value, source string) error {
	obj, err := runtime.CastObject(event)
	if err != nil {
		return wrapError(source, StageParse, err, "expected XML event object")
	}

	eventType, err := stringField(ctx, obj, "type")
	if err != nil {
		return wrapError(source, StageParse, err, "invalid XML event type")
	}

	switch eventType {
	case xmlEventStartElement:
		name, err := stringField(ctx, obj, "name")
		if err != nil {
			return wrapError(source, StageParse, err, "start element is missing name")
		}

		return handleStart(doc, stack, name, source)
	case xmlEventEndElement:
		name, err := stringField(ctx, obj, "name")
		if err != nil {
			return wrapError(source, StageParse, err, "end element is missing name")
		}

		return handleEnd(doc, stack, name, source)
	case xmlEventText:
		text, err := stringField(ctx, obj, "value")
		if err != nil {
			return wrapError(source, StageParse, err, "text event is missing value")
		}

		return handleText(stack, text, source)
	default:
		return newErrorf(source, StageParse, "unsupported XML event type %q", eventType)
	}
}

func handleStart(doc *Document, stack *[]frame, name, source string) error {
	local := localName(name)

	if len(*stack) == 0 {
		switch local {
		case TypeURLSet:
			doc.Type = TypeURLSet
			*stack = append(*stack, frame{name: name, kind: frameRootURLSet})
			return nil
		case TypeSitemapIndex:
			doc.Type = TypeSitemapIndex
			*stack = append(*stack, frame{name: name, kind: frameRootSitemapIndex})
			return nil
		default:
			return newErrorf(source, StageParse, "unsupported root element %q", local)
		}
	}

	parent := (*stack)[len(*stack)-1]
	switch parent.kind {
	case frameFieldLoc, frameFieldLastMod, frameFieldChangeFreq, frameFieldPriority:
		return newErrorf(source, StageParse, "%q must not contain nested elements", localName(parent.name))
	case frameIgnored:
		*stack = append(*stack, frame{name: name, kind: frameIgnored})
	case frameRootURLSet:
		if local == "url" {
			*stack = append(*stack, frame{
				name: name,
				kind: frameURLEntry,
				url: URLEntry{
					Source: source,
				},
			})
		} else {
			*stack = append(*stack, frame{name: name, kind: frameIgnored})
		}
	case frameRootSitemapIndex:
		if local == "sitemap" {
			*stack = append(*stack, frame{name: name, kind: frameSitemapEntry})
		} else {
			*stack = append(*stack, frame{name: name, kind: frameIgnored})
		}
	case frameURLEntry:
		*stack = append(*stack, newEntryChildFrame(name, local))
	case frameSitemapEntry:
		switch local {
		case "loc":
			*stack = append(*stack, frame{name: name, kind: frameFieldLoc})
		case "lastmod":
			*stack = append(*stack, frame{name: name, kind: frameFieldLastMod})
		default:
			*stack = append(*stack, frame{name: name, kind: frameIgnored})
		}
	default:
		return newErrorf(source, StageParse, "unsupported sitemap parse frame %d", parent.kind)
	}

	return nil
}

func newEntryChildFrame(name, local string) frame {
	switch local {
	case "loc":
		return frame{name: name, kind: frameFieldLoc}
	case "lastmod":
		return frame{name: name, kind: frameFieldLastMod}
	case "changefreq":
		return frame{name: name, kind: frameFieldChangeFreq}
	case "priority":
		return frame{name: name, kind: frameFieldPriority}
	default:
		return frame{name: name, kind: frameIgnored}
	}
}

func handleText(stack *[]frame, text, source string) error {
	if len(*stack) == 0 {
		if strings.TrimSpace(text) != "" {
			return newError(source, StageParse, "text outside root element is not supported")
		}

		return nil
	}

	top := &(*stack)[len(*stack)-1]
	switch top.kind {
	case frameFieldLoc, frameFieldLastMod, frameFieldChangeFreq, frameFieldPriority:
		top.text.WriteString(text)
	case frameIgnored:
		return nil
	case frameRootURLSet:
		if strings.TrimSpace(text) != "" {
			return newError(source, StageParse, "text inside urlset outside url entries is not supported")
		}
	case frameRootSitemapIndex:
		if strings.TrimSpace(text) != "" {
			return newError(source, StageParse, "text inside sitemapindex outside sitemap entries is not supported")
		}
	case frameURLEntry:
		if strings.TrimSpace(text) != "" {
			return newError(source, StageParse, "url entries must not contain text outside known fields")
		}
	case frameSitemapEntry:
		if strings.TrimSpace(text) != "" {
			return newError(source, StageParse, "sitemap entries must not contain text outside known fields")
		}
	default:
		return newErrorf(source, StageParse, "unsupported sitemap parse frame %d", top.kind)
	}

	return nil
}

func handleEnd(doc *Document, stack *[]frame, name, source string) error {
	if len(*stack) == 0 {
		return newErrorf(source, StageParse, "unexpected closing tag %q", localName(name))
	}

	index := len(*stack) - 1
	current := (*stack)[index]
	*stack = (*stack)[:index]

	switch current.kind {
	case frameRootURLSet, frameRootSitemapIndex, frameIgnored:
		return nil
	case frameFieldLoc, frameFieldLastMod, frameFieldChangeFreq, frameFieldPriority:
		return applyField(stack, current, source)
	case frameURLEntry:
		if current.url.Loc == "" {
			return newError(source, StageParse, "url entry is missing loc")
		}

		doc.URLs = append(doc.URLs, current.url)
		return nil
	case frameSitemapEntry:
		if current.sitemap.Loc == "" {
			return newError(source, StageParse, "sitemap entry is missing loc")
		}

		doc.Sitemaps = append(doc.Sitemaps, current.sitemap)
		return nil
	default:
		return newErrorf(source, StageParse, "unsupported sitemap parse frame %d", current.kind)
	}
}

func applyField(stack *[]frame, field frame, source string) error {
	if len(*stack) == 0 {
		return newErrorf(source, StageParse, "field %q has no parent entry", localName(field.name))
	}

	value := strings.TrimSpace(field.text.String())
	parent := &(*stack)[len(*stack)-1]

	switch parent.kind {
	case frameURLEntry:
		switch field.kind {
		case frameFieldLoc:
			parent.url.Loc = value
		case frameFieldLastMod:
			parent.url.LastMod = value
		case frameFieldChangeFreq:
			parent.url.ChangeFreq = value
		case frameFieldPriority:
			if value == "" {
				parent.url.Priority = nil
				return nil
			}

			priority, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return wrapError(source, StageParse, err, "invalid priority value")
			}

			parent.url.Priority = &priority
		default:
			return newErrorf(source, StageParse, "unsupported URL field frame %d", field.kind)
		}
	case frameSitemapEntry:
		switch field.kind {
		case frameFieldLoc:
			parent.sitemap.Loc = value
		case frameFieldLastMod:
			parent.sitemap.LastMod = value
		default:
			return newErrorf(source, StageParse, "unsupported sitemap field frame %d", field.kind)
		}
	default:
		return newErrorf(source, StageParse, "field %q has invalid parent frame %d", localName(field.name), parent.kind)
	}

	return nil
}

func stringField(ctx context.Context, obj *runtime.Object, key string) (string, error) {
	value, err := obj.Get(ctx, runtime.NewString(key))
	if err != nil {
		return "", err
	}

	str, ok := value.(runtime.String)
	if !ok {
		return "", runtime.TypeErrorOf(value, runtime.TypeString)
	}

	return str.String(), nil
}

func localName(name string) string {
	if idx := strings.LastIndex(name, ":"); idx >= 0 && idx+1 < len(name) {
		return name[idx+1:]
	}

	return name
}
