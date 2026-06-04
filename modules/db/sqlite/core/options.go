package core

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

const (
	openModeReadOnly      = "ro"
	openModeReadWrite     = "rw"
	openModeReadWriteMake = "rwc"
)

// OpenOptions configures SQLite database creation.
type OpenOptions struct {
	Path     *string `json:"path"`
	Memory   *bool   `json:"memory"`
	URI      *string `json:"uri"`
	ReadOnly *bool   `json:"readOnly"`
	Create   *bool   `json:"create"`
}

// DecodeOpenOptions decodes a Ferret options object into OpenOptions.
func DecodeOpenOptions(value runtime.Value) (OpenOptions, error) {
	optsMap, err := runtime.Cast[runtime.Map](value)
	if err != nil {
		return OpenOptions{}, OperationError("OPEN", err)
	}

	var opts OpenOptions
	if err := sdk.Decode(optsMap, &opts); err != nil {
		return OpenOptions{}, OperationError("OPEN", err)
	}

	return opts, nil
}

func (o OpenOptions) dsn() (string, error) {
	readOnly := valueOrDefault(o.ReadOnly, false)
	create := valueOrDefault(o.Create, true)
	if readOnly && create {
		return "", fmt.Errorf("readOnly and create cannot both be true")
	}

	sourceCount := 0
	if o.pathProvided() {
		sourceCount++
	}
	if o.memoryProvided() {
		sourceCount++
	}
	if o.uriProvided() {
		sourceCount++
	}
	if sourceCount != 1 {
		return "", fmt.Errorf("exactly one of path, memory, or uri must be provided")
	}

	switch {
	case o.memoryProvided():
		return ":memory:", nil
	case o.uriProvided():
		return *o.URI, nil
	default:
		mode := openModeReadWriteMake
		if readOnly {
			mode = openModeReadOnly
		} else if !create {
			mode = openModeReadWrite
		}

		return pathDSN(*o.Path, mode), nil
	}
}

func (o OpenOptions) pathProvided() bool {
	return o.Path != nil && strings.TrimSpace(*o.Path) != ""
}

func (o OpenOptions) memoryProvided() bool {
	return o.Memory != nil && *o.Memory
}

func (o OpenOptions) uriProvided() bool {
	return o.URI != nil && strings.TrimSpace(*o.URI) != ""
}

func valueOrDefault(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}

	return *value
}

func pathDSN(path, mode string) string {
	query := url.Values{}
	query.Set("mode", mode)

	escapedPath := strings.ReplaceAll(path, "?", "%3f")

	return "file:" + escapedPath + "?" + query.Encode()
}
