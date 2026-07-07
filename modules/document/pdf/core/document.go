package core

import (
	"context"
	"fmt"
	"strings"
	"sync"

	ledongpdf "github.com/ledongthuc/pdf"

	commonresource "github.com/MontFerret/contrib/pkg/common/resource"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Document is an opaque PDF document handle exposed to Ferret.
type Document struct {
	source *pdfSource
	reader *ledongpdf.Reader
	path   string
	mu     sync.RWMutex
	id     uint64
	closed bool
}

// Open opens a PDF document through Ferret's filesystem in ctx.
func Open(ctx context.Context, path string, options ...OpenOptions) (*Document, error) {
	opts := DefaultOpenOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	var err error
	opts, err = opts.normalize()
	if err != nil {
		return nil, err
	}

	source, err := openSource(ctx, path, opts)
	if err != nil {
		return nil, OperationError("OPEN", err)
	}

	if err := ctx.Err(); err != nil {
		_ = source.close()
		return nil, OperationError("OPEN", err)
	}

	reader, err := ledongpdf.NewReader(source.reader, source.size)
	if err != nil {
		_ = source.close()
		return nil, OperationErrorf("OPEN", "invalid PDF document %q: %w", path, err)
	}

	if err := ctx.Err(); err != nil {
		_ = source.close()
		return nil, OperationError("OPEN", err)
	}

	return &Document{
		source: source,
		reader: reader,
		path:   path,
		id:     newResourceID(),
	}, nil
}

func (d *Document) PageCount() (int, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if err := d.ensureOpen(); err != nil {
		return 0, OperationError("PAGE_COUNT", err)
	}

	return d.reader.NumPage(), nil
}

func (d *Document) Pages() ([]*Page, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if err := d.ensureOpen(); err != nil {
		return nil, OperationError("PAGES", err)
	}

	count := d.reader.NumPage()
	pages := make([]*Page, 0, count)
	for number := 1; number <= count; number++ {
		pages = append(pages, NewPage(d, number))
	}

	return pages, nil
}

func (d *Document) Page(number int) (*Page, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if err := d.ensureOpen(); err != nil {
		return nil, OperationError("PAGE", err)
	}
	if err := d.validatePageNumber(number); err != nil {
		return nil, OperationError("PAGE", err)
	}

	return NewPage(d, number), nil
}

func (d *Document) Text(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", OperationError("TEXT", err)
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	if err := d.ensureOpen(); err != nil {
		return "", OperationError("TEXT", err)
	}

	count := d.reader.NumPage()
	parts := make([]string, 0, count)
	for number := 1; number <= count; number++ {
		if err := ctx.Err(); err != nil {
			return "", OperationError("TEXT", err)
		}

		text, err := d.pageTextLocked(number)
		if err != nil {
			return "", OperationErrorf("TEXT", "page %d: %w", number, err)
		}
		parts = append(parts, text)
	}

	return strings.Join(parts, "\n\n"), nil
}

func (d *Document) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil
	}

	d.closed = true
	d.reader = nil

	if err := d.source.close(); err != nil {
		return OperationError("CLOSE", err)
	}
	d.source = nil

	return nil
}

func (d *Document) ResourceID() uint64 {
	return d.id
}

func (d *Document) String() string {
	path := d.path
	if path == "" {
		path = "<memory>"
	}

	return fmt.Sprintf("PDFDocument(%q)", path)
}

func (d *Document) Hash() uint64 {
	return commonresource.Hash("document.pdf.document", d.id)
}

func (d *Document) Copy() runtime.Value {
	return d
}

func (d *Document) MarshalJSON() ([]byte, error) {
	return commonresource.MarshalStringJSON(d.String())
}

func (d *Document) ensureOpen() error {
	if d == nil || d.closed || d.reader == nil {
		return errDocumentClosed
	}

	return nil
}

func (d *Document) validatePageNumber(number int) error {
	count := d.reader.NumPage()
	if number < 1 || number > count {
		return fmt.Errorf("page number %d is out of range; document has %d pages", number, count)
	}

	return nil
}

func (d *Document) pdfPageLocked(number int) (ledongpdf.Page, error) {
	if err := d.validatePageNumber(number); err != nil {
		return ledongpdf.Page{}, err
	}

	page := d.reader.Page(number)
	if page.V.IsNull() {
		return ledongpdf.Page{}, errInvalidPage
	}

	return page, nil
}

func (d *Document) pageTextLocked(number int) (string, error) {
	page, err := d.pdfPageLocked(number)
	if err != nil {
		return "", err
	}

	text, err := page.GetPlainText(nil)
	if err != nil {
		return "", err
	}

	return text, nil
}
