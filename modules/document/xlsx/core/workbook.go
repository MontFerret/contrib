package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/xuri/excelize/v2"

	commonresource "github.com/MontFerret/contrib/pkg/common/resource"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Workbook is an opaque XLSX workbook handle exposed to Ferret.
type Workbook struct {
	file           *excelize.File
	generations    map[string]uint64
	path           string
	mu             sync.Mutex
	id             uint64
	nextGeneration uint64
	closed         bool
}

// Create creates a new in-memory XLSX workbook handle.
func Create() *Workbook {
	return NewWorkbook(excelize.NewFile(), "")
}

// Open opens an existing XLSX workbook through the Ferret filesystem in ctx.
func Open(ctx context.Context, path string) (*Workbook, error) {
	if path == "" {
		return nil, OperationErrorf("OPEN", "path must not be empty")
	}

	file, err := openWorkbookFile(ctx, path)
	if err != nil {
		return nil, OperationError("OPEN", err)
	}

	return NewWorkbook(file, path), nil
}

// NewWorkbook creates a tracked XLSX workbook handle around an Excel file.
func NewWorkbook(file *excelize.File, path string) *Workbook {
	w := &Workbook{
		file:        file,
		generations: make(map[string]uint64),
		path:        path,
		id:          newResourceID(),
	}

	for _, name := range file.GetSheetList() {
		w.nextGeneration++
		w.generations[name] = w.nextGeneration
	}

	return w
}

func (w *Workbook) Sheets() ([]string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.ensureOpen(); err != nil {
		return nil, OperationError("SHEETS", err)
	}

	return append([]string(nil), w.file.GetSheetList()...), nil
}

func (w *Workbook) Sheet(name string) (*Worksheet, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.ensureOpen(); err != nil {
		return nil, OperationError("SHEET", err)
	}

	generation, err := w.activeSheetGeneration(name)
	if err != nil {
		return nil, OperationError("SHEET", err)
	}

	return NewWorksheet(w, name, generation), nil
}

func (w *Workbook) AddSheet(name string) (*Worksheet, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.ensureOpen(); err != nil {
		return nil, OperationError("ADD_SHEET", err)
	}

	if err := validateSheetName(name); err != nil {
		return nil, OperationError("ADD_SHEET", err)
	}

	if _, exists := w.generations[name]; exists {
		return nil, OperationErrorf("ADD_SHEET", "worksheet %q already exists", name)
	}

	if _, err := w.file.NewSheet(name); err != nil {
		return nil, OperationError("ADD_SHEET", err)
	}

	w.nextGeneration++
	w.generations[name] = w.nextGeneration

	return NewWorksheet(w, name, w.generations[name]), nil
}

func (w *Workbook) DeleteSheet(name string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.ensureOpen(); err != nil {
		return OperationError("DELETE_SHEET", err)
	}

	if _, err := w.activeSheetGeneration(name); err != nil {
		return OperationError("DELETE_SHEET", err)
	}

	if err := w.file.DeleteSheet(name); err != nil {
		return OperationError("DELETE_SHEET", err)
	}

	delete(w.generations, name)

	return nil
}

func (w *Workbook) Save(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.ensureOpen(); err != nil {
		return OperationError("SAVE", err)
	}

	if w.path == "" {
		return OperationErrorf("SAVE", "cannot save an in-memory workbook without a path; use DOCUMENT::XLSX::SAVE_AS")
	}

	if err := writeWorkbookFile(ctx, w.file, w.path); err != nil {
		return OperationError("SAVE", err)
	}

	return nil
}

func (w *Workbook) SaveAs(ctx context.Context, path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.ensureOpen(); err != nil {
		return OperationError("SAVE_AS", err)
	}

	if path == "" {
		return OperationErrorf("SAVE_AS", "path must not be empty")
	}

	if err := ensureWorkbookParentDirectory(ctx, path); err != nil {
		return OperationError("SAVE_AS", err)
	}

	if err := writeWorkbookFile(ctx, w.file, path); err != nil {
		return OperationError("SAVE_AS", err)
	}

	w.path = path

	return nil
}

func (w *Workbook) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}

	w.closed = true
	w.generations = make(map[string]uint64)

	if err := w.file.Close(); err != nil {
		return OperationError("CLOSE", err)
	}

	return nil
}

func (w *Workbook) ResourceID() uint64 {
	return w.id
}

func (w *Workbook) String() string {
	path := w.path
	if path == "" {
		path = "<memory>"
	}

	return fmt.Sprintf("XLSXWorkbook(%q)", path)
}

func (w *Workbook) Hash() uint64 {
	return commonresource.Hash("document.xlsx.workbook", w.id)
}

func (w *Workbook) Copy() runtime.Value {
	return w
}

func (w *Workbook) MarshalJSON() ([]byte, error) {
	return commonresource.MarshalStringJSON(w.String())
}

func (w *Workbook) ensureOpen() error {
	if w.closed {
		return errWorkbookClosed
	}

	return nil
}

func (w *Workbook) activeSheetGeneration(name string) (uint64, error) {
	if _, err := w.file.GetSheetIndex(name); err != nil {
		return 0, fmt.Errorf("worksheet %q does not exist", name)
	}

	generation, ok := w.generations[name]
	if !ok {
		return 0, fmt.Errorf("worksheet %q does not exist", name)
	}

	return generation, nil
}

func (w *Workbook) validateWorksheet(name string, generation uint64) error {
	if err := w.ensureOpen(); err != nil {
		return err
	}

	current, ok := w.generations[name]
	if !ok {
		return errSheetDeleted
	}
	if current != generation {
		return errSheetStale
	}
	if _, err := w.file.GetSheetIndex(name); err != nil {
		return fmt.Errorf("worksheet %q does not exist", name)
	}

	return nil
}
