package types

import (
	"encoding/csv"
	"fmt"
)

type Options struct {
	Delimiter  string   `json:"delimiter"`
	Header     bool     `json:"header"`
	Columns    []string `json:"columns"`
	Trim       bool     `json:"trim"`
	SkipEmpty  bool     `json:"skipEmpty"`
	Strict     bool     `json:"strict"`
	Comment    string   `json:"comment"`
	InferTypes bool     `json:"inferTypes"`
	NullValues []string `json:"nullValues"`
}

func DefaultOptions() Options {
	return Options{
		Delimiter: ",",
		Header:    true,
		SkipEmpty: true,
		Strict:    true,
	}
}

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
