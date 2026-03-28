package types

import (
	"encoding/csv"
	"fmt"
	"unicode/utf8"
)

// Options configures CSV decoding and encoding behavior.
type Options struct {
	// Delimiter is the single-character field separator. It defaults to "," and
	// must be a valid CSV delimiter rune.
	Delimiter string `json:"delimiter"`
	// Header controls whether the first row is treated as column names. It
	// defaults to true.
	Header bool `json:"header"`
	// Columns provides explicit column names for object decoding and object
	// encoding order.
	Columns []string `json:"columns"`
	// Trim trims leading space while decoding fields.
	Trim bool `json:"trim"`
	// SkipEmpty skips rows whose fields are all empty strings. It defaults to
	// true.
	SkipEmpty bool `json:"skipEmpty"`
	// Strict enforces matching field counts and strict header validation. It
	// defaults to true.
	Strict bool `json:"strict"`
	// Comment marks a single-character comment prefix for reader input. It must
	// be a valid CSV comment rune and must differ from the effective delimiter.
	Comment string `json:"comment"`
	// InferTypes converts decoded field values to booleans or numbers when
	// possible.
	InferTypes bool `json:"inferTypes"`
	// NullValues lists string values that should decode as runtime.None.
	NullValues []string `json:"nullValues"`
}

// DefaultOptions returns the default CSV options used by the module.
func DefaultOptions() Options {
	return Options{
		Delimiter: ",",
		Header:    true,
		SkipEmpty: true,
		Strict:    true,
	}
}

// ApplyToReader applies the reader-supported options to r.
func (o *Options) ApplyToReader(r *csv.Reader) error {
	delimiter := r.Comma

	if o.Delimiter != "" {
		validDelimiter, err := validateSingleRuneOption("delimiter", o.Delimiter)
		if err != nil {
			return err
		}

		delimiter = validDelimiter
		r.Comma = delimiter
	}

	if o.Comment != "" {
		comment, err := validateSingleRuneOption("comment", o.Comment)
		if err != nil {
			return err
		}

		if comment == delimiter {
			return fmt.Errorf("csv: comment must differ from delimiter")
		}

		r.Comment = comment
	}

	r.LazyQuotes = !o.Strict
	r.TrimLeadingSpace = o.Trim
	r.FieldsPerRecord = -1 // we handle field count validation ourselves

	return nil
}

// ApplyToWriter applies the writer-supported options to w.
func (o *Options) ApplyToWriter(w *csv.Writer) error {
	if o.Delimiter == "" {
		return nil
	}

	delimiter, err := validateSingleRuneOption("delimiter", o.Delimiter)
	if err != nil {
		return err
	}

	w.Comma = delimiter

	return nil
}

func validateSingleRuneOption(name, value string) (rune, error) {
	runes := []rune(value)
	if len(runes) != 1 {
		return 0, fmt.Errorf("csv: %s must be exactly one Unicode character", name)
	}

	r := runes[0]
	if !isValidCSVOptionRune(r) {
		return 0, fmt.Errorf("csv: invalid %s character", name)
	}

	return r, nil
}

func isValidCSVOptionRune(r rune) bool {
	return r != 0 && r != '"' && r != '\r' && r != '\n' && utf8.ValidRune(r) && r != utf8.RuneError
}
