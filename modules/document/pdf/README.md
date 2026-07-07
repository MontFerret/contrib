# DOCUMENT::PDF Module

`github.com/MontFerret/contrib/modules/document/pdf` registers read-only PDF helpers under the `DOCUMENT::PDF` namespace.

The module opens PDFs through Ferret's filesystem abstraction, exposes lazy document and page handles, extracts best-effort text, reports basic page dimensions, and returns low-level positioned text fragments. It uses `github.com/ledongthuc/pdf` internally, but that dependency is hidden behind the module-owned core types so the public FQL API can remain stable if the parser is replaced later.

Out of scope for this version: PDF creation or modification, page insertion or deletion, form filling, annotation editing, digital signatures, password management, OCR, table recognition, paragraph or heading detection, metadata extraction, link extraction, image extraction, rendering, PDF-to-HTML conversion, remote HTTP URLs, and aliases under `PDF::`.

## Install

```sh
go get github.com/MontFerret/contrib/modules/document/pdf
```

## Register The Module

```go
package main

import (
	"github.com/MontFerret/ferret/v2"

	pdfmodule "github.com/MontFerret/contrib/modules/document/pdf"
)

func main() {
	engine, err := ferret.New(
		ferret.WithFSRoot("./documents"),
		ferret.WithModules(pdfmodule.New()),
	)
	if err != nil {
		panic(err)
	}

	_ = engine
}
```

`OPEN` uses Ferret's filesystem from the execution context. Configure `ferret.WithFSRoot` in the host application and use paths relative to that root. The module does not open unrestricted host paths directly.

When the filesystem source cannot provide random access, the module buffers the PDF in memory and passes a `bytes.Reader` to the parser. The default buffer limit is 64 MiB and can be configured at registration time:

```go
ferret.WithModules(pdfmodule.New(pdfmodule.WithMaxBufferSize(128 * 1024 * 1024)))
```

## Function Reference

| Function | Signature | Returns | Notes |
| --- | --- | --- | --- |
| `DOCUMENT::PDF::OPEN` | `OPEN(path)` | Document handle | Opens an existing PDF through Ferret's filesystem. |
| `DOCUMENT::PDF::PAGE_COUNT` | `PAGE_COUNT(document)` | `Int` | Returns the document page count. |
| `DOCUMENT::PDF::PAGES` | `PAGES(document)` | `Array<Page>` | Returns lazy one-based page handles. |
| `DOCUMENT::PDF::PAGE` | `PAGE(document, number)` | Page handle | Returns one lazy page handle; `number` is one-based. |
| `DOCUMENT::PDF::TEXT` | `TEXT(document)` | `String` | Extracts best-effort text from all pages in order, separated by a blank line. |
| `DOCUMENT::PDF::TEXT` | `TEXT(page)` | `String` | Extracts best-effort text from one page. |
| `DOCUMENT::PDF::PAGE_INFO` | `PAGE_INFO(page)` | `Object` | Returns page number, width, height, and rotation. |
| `DOCUMENT::PDF::BLOCKS` | `BLOCKS(page)` | `Array<Object>` | Returns low-level positioned text fragments. |
| `DOCUMENT::PDF::CLOSE` | `CLOSE(document)` | `Boolean` | Releases document resources. Repeated close is idempotent. |

## Examples

```fql
LET document = DOCUMENT::PDF::OPEN("./report.pdf")
RETURN {
  pages: DOCUMENT::PDF::PAGE_COUNT(document),
  text: DOCUMENT::PDF::TEXT(document)
}
```

```fql
LET document = DOCUMENT::PDF::OPEN("./report.pdf")
RETURN (
  FOR page IN DOCUMENT::PDF::PAGES(document)
    RETURN {
      info: DOCUMENT::PDF::PAGE_INFO(page),
      text: DOCUMENT::PDF::TEXT(page)
    }
)
```

```fql
LET document = DOCUMENT::PDF::OPEN("./report.pdf")
LET page = DOCUMENT::PDF::PAGE(document, 2)
RETURN DOCUMENT::PDF::BLOCKS(page)
```

## Page And Coordinate Conventions

Public page numbers are one-based. `PAGE(document, 1)` returns the first page.

Page dimensions and text bounds are in PDF points. For page dimensions, the module prefers a valid crop box and falls back to the media box. Positioned text coordinates use the coordinate system returned by `github.com/ledongthuc/pdf`: X increases left to right, Y increases bottom to top, and the origin is bottom-left. The module does not convert coordinates.

`BLOCKS` returns positioned text entries from the parser, not semantic paragraphs, headings, table cells, or layout regions. Empty and whitespace-only entries are omitted.

## Text Extraction Limitations

Text extraction is best-effort. Reading order can be wrong for multi-column or heavily positioned documents, custom font encodings can produce missing or incorrect characters, and malformed or uncommon PDFs can fail to parse. Scanned pages without embedded text usually return an empty string. OCR is not performed.

Password-protected and encrypted PDFs may not be supported reliably by the underlying parser. Metadata and link APIs are planned follow-up functionality and are not registered as stubs in this version.

## Lifecycle And Concurrency

Document handles retain the source reader or backing buffer while open. Page handles retain a reference to their parent document and become invalid after the document is closed. Document cleanup is safe and idempotent.

Document and page operations guard shared open/closed state and are safe for ordinary concurrent reads. Cancellation is cooperative between major operations, including open, page extraction, full-document page steps, and positioned-text iteration.
