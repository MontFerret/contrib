package core

import (
	"fmt"
	"strings"
)

// Format identifies a supported archive representation.
type Format string

const (
	FormatAuto  Format = "auto"
	FormatZIP   Format = "zip"
	FormatTAR   Format = "tar"
	FormatTARGZ Format = "tar.gz"
)

func normalizeFormat(value Format) (Format, error) {
	switch Format(strings.ToLower(string(value))) {
	case "", FormatAuto:
		return FormatAuto, nil
	case FormatZIP:
		return FormatZIP, nil
	case FormatTAR:
		return FormatTAR, nil
	case FormatTARGZ, "tgz":
		return FormatTARGZ, nil
	default:
		return "", fmt.Errorf("unsupported archive format %q", value)
	}
}

func formatFromName(name string) Format {
	lower := strings.ToLower(name)

	switch {
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return FormatTARGZ
	case strings.HasSuffix(lower, ".zip"):
		return FormatZIP
	case strings.HasSuffix(lower, ".tar"):
		return FormatTAR
	default:
		return FormatAuto
	}
}
