package core

import (
	"encoding/csv"
	"fmt"
	"unicode/utf8"
)

// Options configures CSV decoding and encoding behavior.
type Options struct {
	Delimiter  string   `json:"delimiter"`
	Comment    string   `json:"comment"`
	Columns    []string `json:"columns"`
	NullValues []string `json:"nullValues"`
	Header     bool     `json:"header"`
	Trim       bool     `json:"trim"`
	SkipEmpty  bool     `json:"skipEmpty"`
	Strict     bool     `json:"strict"`
	InferTypes bool     `json:"inferTypes"`
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
