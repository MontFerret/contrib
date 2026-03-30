package core

import (
	"context"
	"io"
)

type (
	pathNode struct {
		Parent *pathNode
		URL    string
	}

	workItem struct {
		Path  *pathNode
		URL   string
		Depth int
	}

	walker struct {
		visitedSitemaps map[string]struct{}
		yieldedURLs     map[string]struct{}
		pending         []workItem
		current         []URLEntry
		opts            Options
		currentIndex    int
		done            bool
	}
)

// CollectURLs eagerly expands a sitemap tree into URL entries.
func CollectURLs(ctx context.Context, target string, opts Options) ([]URLEntry, error) {
	w := newWalker(target, opts)
	entries := make([]URLEntry, 0)

	for {
		entry, err := w.nextEntry(ctx)
		if err != nil {
			if err == io.EOF {
				return entries, nil
			}

			return nil, err
		}

		entries = append(entries, entry)
	}
}

func newWalker(target string, opts Options) *walker {
	w := &walker{
		opts: opts,
		pending: []workItem{
			{
				URL:   target,
				Depth: 0,
				Path: &pathNode{
					URL: target,
				},
			},
		},
	}

	if opts.Dedupe {
		w.visitedSitemaps = make(map[string]struct{})
		w.yieldedURLs = make(map[string]struct{})
	}

	return w
}

func (w *walker) close() {
	w.done = true
	w.pending = nil
	w.current = nil
	w.currentIndex = 0
}

func (w *walker) nextEntry(ctx context.Context) (URLEntry, error) {
	if w.done {
		return URLEntry{}, io.EOF
	}

	for {
		if w.currentIndex < len(w.current) {
			entry := w.current[w.currentIndex]
			w.currentIndex++

			if w.opts.Dedupe {
				if _, exists := w.yieldedURLs[entry.Loc]; exists {
					continue
				}

				w.yieldedURLs[entry.Loc] = struct{}{}
			}

			return entry, nil
		}

		if len(w.pending) == 0 {
			w.close()

			return URLEntry{}, io.EOF
		}

		item := w.pending[len(w.pending)-1]
		w.pending = w.pending[:len(w.pending)-1]

		if w.opts.Dedupe {
			if _, exists := w.visitedSitemaps[item.URL]; exists {
				continue
			}
		}

		document, err := Fetch(ctx, item.URL, w.opts)
		if err != nil {
			w.close()

			return URLEntry{}, err
		}

		if w.opts.Dedupe {
			w.visitedSitemaps[item.URL] = struct{}{}
		}

		switch document.Type {
		case TypeURLSet:
			w.current = document.URLs
			w.currentIndex = 0
		case TypeSitemapIndex:
			w.current = nil
			w.currentIndex = 0

			if !w.opts.Recursive {
				continue
			}

			if item.Depth >= w.opts.MaxDepth && len(document.Sitemaps) > 0 {
				w.close()

				return URLEntry{}, newErrorf(item.URL, StageExpand, "maximum recursion depth %d exceeded", w.opts.MaxDepth)
			}

			for idx := len(document.Sitemaps) - 1; idx >= 0; idx-- {
				ref := document.Sitemaps[idx]

				if inPath(item.Path, ref.Loc) {
					continue
				}

				if w.opts.Dedupe {
					if _, exists := w.visitedSitemaps[ref.Loc]; exists {
						continue
					}
				}

				w.pending = append(w.pending, workItem{
					URL:   ref.Loc,
					Depth: item.Depth + 1,
					Path: &pathNode{
						URL:    ref.Loc,
						Parent: item.Path,
					},
				})
			}
		default:
			w.close()

			return URLEntry{}, newErrorf(item.URL, StageParse, "unsupported sitemap document type %q", document.Type)
		}
	}
}

func inPath(node *pathNode, target string) bool {
	for current := node; current != nil; current = current.Parent {
		if current.URL == target {
			return true
		}
	}

	return false
}
