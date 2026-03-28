package types

import (
	"encoding/csv"
	"fmt"
)

// Options configures CSV decoding and encoding behavior.
type Options struct {
	// Delimiter is the single-character field separator. It defaults to ",".
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
	// Comment marks a single-character comment prefix for reader input.
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
	if o.Delimiter != "" {
		delimiter, err := validateSingleRuneOption("delimiter", o.Delimiter)
		if err != nil {
			return err
		}

		r.Comma = delimiter
	}

	if o.Comment != "" {
		comment, err := validateSingleRuneOption("comment", o.Comment)
		if err != nil {
			return err
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

	return runes[0], nil
}
