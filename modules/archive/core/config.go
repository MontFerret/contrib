package core

import (
	"fmt"
	"strings"
)

const defaultLimit int64 = 64 << 20

type (
	// Config controls archive resource limits.
	Config struct {
		MaxEntrySize     int64
		MaxZIPBufferSize int64
	}

	// ListConfig controls archive listing.
	ListConfig struct {
		Format Format `json:"format"`
		Config
	}

	// ReadConfig controls archive entry reads.
	ReadConfig struct {
		Format  Format `json:"format"`
		As      string `json:"as"`
		Missing string `json:"missing"`
		Config
	}

	// ExtractConfig controls archive extraction.
	ExtractConfig struct {
		Format  Format   `json:"format"`
		Links   string   `json:"links"`
		Include []string `json:"include"`
		Exclude []string `json:"exclude"`
		Config
		Overwrite  bool `json:"overwrite"`
		CreateDirs bool `json:"createDirs"`
	}
)

// DefaultConfig returns the archive module's default resource limits.
func DefaultConfig() Config {
	return Config{
		MaxEntrySize:     defaultLimit,
		MaxZIPBufferSize: defaultLimit,
	}
}

func DefaultConfigOr(cfg []Config) Config {
	if len(cfg) > 0 {
		return cfg[0]
	}

	return DefaultConfig()
}

// DefaultListOptions returns LIST defaults.
func DefaultListOptions(cfg ...Config) ListConfig {
	return ListConfig{
		Config: DefaultConfigOr(cfg),
		Format: FormatAuto,
	}
}

// DefaultReadOptions returns READ defaults.
func DefaultReadOptions(cfg ...Config) ReadConfig {
	return ReadConfig{
		Config:  DefaultConfigOr(cfg),
		Format:  FormatAuto,
		As:      "binary",
		Missing: "error",
	}
}

// DefaultExtractOptions returns EXTRACT defaults.
func DefaultExtractOptions(cfg ...Config) ExtractConfig {
	return ExtractConfig{
		Config:     DefaultConfigOr(cfg),
		Format:     FormatAuto,
		CreateDirs: true,
		Links:      "skip",
	}
}

func (c ListConfig) normalize() (ListConfig, error) {
	format, err := normalizeFormat(c.Format)
	if err != nil {
		return c, err
	}

	c.Format = format

	return c, nil
}

func (o ReadConfig) normalize() (ReadConfig, error) {
	format, err := normalizeFormat(o.Format)
	if err != nil {
		return o, err
	}

	o.Format = format
	o.As = strings.ToLower(o.As)
	o.Missing = strings.ToLower(o.Missing)

	if o.As != "binary" && o.As != "string" {
		return o, fmt.Errorf("unsupported READ as value %q", o.As)
	}
	if o.Missing != "error" && o.Missing != "none" {
		return o, fmt.Errorf("unsupported READ missing value %q", o.Missing)
	}

	return o, nil
}

func (o ExtractConfig) normalize() (ExtractConfig, error) {
	format, err := normalizeFormat(o.Format)
	if err != nil {
		return o, err
	}
	o.Format = format
	o.Links = strings.ToLower(o.Links)

	if o.Links != "skip" && o.Links != "error" {
		return o, fmt.Errorf("unsupported EXTRACT links value %q", o.Links)
	}

	return o, nil
}
