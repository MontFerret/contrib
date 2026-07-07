package core

import (
	"context"
	"fmt"
	"math"
	"strings"

	commonresource "github.com/MontFerret/contrib/pkg/common/resource"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Page is an opaque PDF page handle exposed to Ferret.
type Page struct {
	document *Document
	number   int
	id       uint64
}

// NewPage creates a lazy PDF page handle.
func NewPage(document *Document, number int) *Page {
	return &Page{
		document: document,
		number:   number,
		id:       newResourceID(),
	}
}

func (p *Page) Number() int {
	return p.number
}

func (p *Page) Text(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", OperationError("TEXT", err)
	}

	p.document.mu.RLock()
	defer p.document.mu.RUnlock()

	if err := p.document.ensureOpen(); err != nil {
		return "", OperationError("TEXT", err)
	}

	text, err := p.document.pageTextLocked(p.number)
	if err != nil {
		return "", OperationError("TEXT", err)
	}
	if err := ctx.Err(); err != nil {
		return "", OperationError("TEXT", err)
	}

	return text, nil
}

func (p *Page) Info(ctx context.Context) (PageInfo, error) {
	if err := ctx.Err(); err != nil {
		return PageInfo{}, OperationError("PAGE_INFO", err)
	}

	p.document.mu.RLock()
	defer p.document.mu.RUnlock()

	if err := p.document.ensureOpen(); err != nil {
		return PageInfo{}, OperationError("PAGE_INFO", err)
	}

	page, err := p.document.pdfPageLocked(p.number)
	if err != nil {
		return PageInfo{}, OperationError("PAGE_INFO", err)
	}

	info, err := pageInfo(page, p.number)
	if err != nil {
		return PageInfo{}, OperationError("PAGE_INFO", err)
	}
	if err := ctx.Err(); err != nil {
		return PageInfo{}, OperationError("PAGE_INFO", err)
	}

	return info, nil
}

func (p *Page) Blocks(ctx context.Context) ([]TextBlock, error) {
	if err := ctx.Err(); err != nil {
		return nil, OperationError("BLOCKS", err)
	}

	p.document.mu.RLock()
	defer p.document.mu.RUnlock()

	if err := p.document.ensureOpen(); err != nil {
		return nil, OperationError("BLOCKS", err)
	}

	page, err := p.document.pdfPageLocked(p.number)
	if err != nil {
		return nil, OperationError("BLOCKS", err)
	}

	content, err := pageContent(page)
	if err != nil {
		return nil, OperationError("BLOCKS", err)
	}

	blocks := make([]TextBlock, 0, len(content.Text))
	for idx, text := range content.Text {
		if idx%128 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, OperationError("BLOCKS", err)
			}
		}
		if strings.TrimSpace(text.S) == "" {
			continue
		}

		blocks = append(blocks, TextBlock{
			Text: text.S,
			Bounds: Bounds{
				X:      text.X,
				Y:      text.Y,
				Width:  math.Abs(text.W),
				Height: math.Abs(text.FontSize),
			},
		})
	}

	return blocks, nil
}

func (p *Page) String() string {
	return fmt.Sprintf("PDFPage(%d)", p.number)
}

func (p *Page) Hash() uint64 {
	return commonresource.Hash("document.pdf.page", p.id)
}

func (p *Page) Copy() runtime.Value {
	return p
}

func (p *Page) MarshalJSON() ([]byte, error) {
	return commonresource.MarshalStringJSON(p.String())
}
