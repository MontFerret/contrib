package types

import "encoding/csv"

type Options struct {
	Delimiter  string   `json:"delimiter"`
	Header     bool     `json:"header"`
	Columns    []string `json:"columns"`
	Trim       bool     `json:"trim"`
	SkipEmpty  bool     `json:"skip_empty"`
	Strict     bool     `json:"strict"`
	Comment    string   `json:"comment"`
	InferTypes bool     `json:"infer_types"`
	NullValues []string `json:"null_values"`
}

func DefaultOptions() Options {
	return Options{
		Delimiter: ",",
		Header:    true,
		SkipEmpty: true,
		Strict:    true,
	}
}

func (o *Options) ApplyToReader(r *csv.Reader) {
	if o.Delimiter != "" {
		runes := []rune(o.Delimiter)
		r.Comma = runes[0]
	}

	if o.Comment != "" {
		runes := []rune(o.Comment)
		r.Comment = runes[0]
	}

	r.LazyQuotes = !o.Strict
	r.TrimLeadingSpace = o.Trim
	r.FieldsPerRecord = -1 // we handle field count validation ourselves
}
