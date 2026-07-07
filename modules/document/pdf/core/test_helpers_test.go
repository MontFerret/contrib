package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	ferretfs "github.com/MontFerret/ferret/v2/pkg/fs"
)

type closeable interface {
	Close() error
}

type pdfTestPage struct {
	CropBox  *[4]float64
	Text     string
	Width    float64
	Height   float64
	X        float64
	Y        float64
	FontSize float64
	Rotation int
}

func testFSContext(t *testing.T, readOnly bool) (context.Context, string) {
	t.Helper()

	root := t.TempDir()
	filesystem, err := ferretfs.New(ferretfs.WithRoot(root), ferretfs.WithReadOnly(readOnly))
	if err != nil {
		t.Fatalf("unexpected filesystem error: %v", err)
	}
	if filesystemCloser, ok := filesystem.(closeable); ok {
		t.Cleanup(func() {
			if err := filesystemCloser.Close(); err != nil {
				t.Fatalf("unexpected filesystem close error: %v", err)
			}
		})
	}

	return ferretfs.WithFileSystem(context.Background(), filesystem), root
}

func writePDFForTest(t *testing.T, root, name string, pages []pdfTestPage) {
	t.Helper()

	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		t.Fatalf("failed to create fixture directory: %v", err)
	}
	if err := os.WriteFile(path, buildPDFForTest(t, pages), 0666); err != nil {
		t.Fatalf("failed to write PDF fixture: %v", err)
	}
}

func buildPDFForTest(t *testing.T, pages []pdfTestPage) []byte {
	t.Helper()

	if len(pages) == 0 {
		t.Fatal("test PDF needs at least one page")
	}

	fontID := 3 + len(pages)*2
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"",
	}
	kids := make([]string, 0, len(pages))

	for idx, page := range pages {
		if page.Width == 0 {
			page.Width = 612
		}
		if page.Height == 0 {
			page.Height = 792
		}
		if page.FontSize == 0 {
			page.FontSize = 24
		}
		if page.X == 0 {
			page.X = 72
		}
		if page.Y == 0 {
			page.Y = 720
		}

		pageID := 3 + idx*2
		contentID := pageID + 1
		kids = append(kids, fmt.Sprintf("%d 0 R", pageID))

		content := ""
		if page.Text != "" {
			content = fmt.Sprintf("BT /F1 %.2f Tf %.2f %.2f Td (%s) Tj ET", page.FontSize, page.X, page.Y, escapePDFText(page.Text))
		}

		pageObject := fmt.Sprintf(
			"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.2f %.2f] ",
			page.Width,
			page.Height,
		)
		if page.CropBox != nil {
			box := page.CropBox
			pageObject += fmt.Sprintf("/CropBox [%.2f %.2f %.2f %.2f] ", box[0], box[1], box[2], box[3])
		}
		if page.Rotation != 0 {
			pageObject += fmt.Sprintf("/Rotate %d ", page.Rotation)
		}
		pageObject += fmt.Sprintf("/Resources << /Font << /F1 %d 0 R >> >> /Contents %d 0 R >>", fontID, contentID)

		objects = append(objects, pageObject)
		objects = append(objects, fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content))
	}

	objects[1] = fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kids, " "), len(pages))
	objects = append(objects, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")

	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")

	offsets := make([]int, len(objects)+1)
	for idx, obj := range objects {
		offsets[idx+1] = out.Len()
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", idx+1, obj)
	}

	xref := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n", len(objects)+1)
	out.WriteString("0000000000 65535 f \n")
	for idx := 1; idx < len(offsets); idx++ {
		fmt.Fprintf(&out, "%010d 00000 n \n", offsets[idx])
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)

	return out.Bytes()
}

func escapePDFText(text string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `(`, `\(`, `)`, `\)`)

	return replacer.Replace(text)
}

func blockText(blocks []TextBlock) string {
	var out strings.Builder
	for _, block := range blocks {
		out.WriteString(block.Text)
	}

	return out.String()
}

type memoryFS struct {
	files    map[string][]byte
	failOpen bool
}

func (m memoryFS) ReadFile(path string) ([]byte, error) {
	data, ok := m.files[path]
	if !ok {
		return nil, fs.ErrNotExist
	}

	return append([]byte(nil), data...), nil
}

func (m memoryFS) Open(path string) (ferretfs.ReadableFile, error) {
	if m.failOpen {
		return nil, errors.New("open blocked")
	}

	data, ok := m.files[path]
	if !ok {
		return nil, fs.ErrNotExist
	}

	return &memoryFile{reader: bytes.NewReader(data), info: memoryFileInfo{name: path, size: int64(len(data))}}, nil
}

func (m memoryFS) OpenFile(string, int, fs.FileMode) (ferretfs.WritableFile, error) {
	return nil, errors.New("not implemented")
}

func (m memoryFS) Stat(path string) (fs.FileInfo, error) {
	data, ok := m.files[path]
	if !ok {
		return nil, fs.ErrNotExist
	}

	return memoryFileInfo{name: path, size: int64(len(data))}, nil
}

func (m memoryFS) Exists(path string) (bool, error) {
	_, ok := m.files[path]

	return ok, nil
}

func (m memoryFS) Mkdir(string, fs.FileMode) error    { return errors.New("not implemented") }
func (m memoryFS) MkdirAll(string, fs.FileMode) error { return errors.New("not implemented") }
func (m memoryFS) WriteFile(string, []byte, fs.FileMode) error {
	return errors.New("not implemented")
}
func (m memoryFS) AppendFile(string, []byte, fs.FileMode) error {
	return errors.New("not implemented")
}
func (m memoryFS) Remove(string) error    { return errors.New("not implemented") }
func (m memoryFS) RemoveAll(string) error { return errors.New("not implemented") }

type memoryFile struct {
	reader *bytes.Reader
	info   memoryFileInfo
}

func (f *memoryFile) Stat() (fs.FileInfo, error) {
	return f.info, nil
}

func (f *memoryFile) Read(data []byte) (int, error) {
	return f.reader.Read(data)
}

func (f *memoryFile) Close() error {
	_, err := io.Copy(io.Discard, f.reader)

	return err
}

type memoryFileInfo struct {
	name string
	size int64
}

func (i memoryFileInfo) Name() string       { return filepath.Base(i.name) }
func (i memoryFileInfo) Size() int64        { return i.size }
func (i memoryFileInfo) Mode() fs.FileMode  { return 0444 }
func (i memoryFileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (i memoryFileInfo) IsDir() bool        { return false }
func (i memoryFileInfo) Sys() any           { return nil }
