package pdf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type pdfTestPage struct {
	Text     string
	Width    float64
	Height   float64
	X        float64
	Y        float64
	FontSize float64
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

		objects = append(
			objects,
			fmt.Sprintf(
				"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.2f %.2f] /Resources << /Font << /F1 %d 0 R >> >> /Contents %d 0 R >>",
				page.Width,
				page.Height,
				fontID,
				contentID,
			),
			fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content),
		)
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
