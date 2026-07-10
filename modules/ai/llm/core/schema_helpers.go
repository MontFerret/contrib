package core

import (
	"encoding/json"
	"io"
	"strings"
)

func hasExternalReference(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if key == "$ref" || key == "$dynamicRef" || key == "$recursiveRef" {
				ref, ok := child.(string)
				if !ok || (ref != "" && !strings.HasPrefix(ref, "#")) {
					return true
				}
			}

			if hasExternalReference(child) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if hasExternalReference(child) {
				return true
			}
		}
	}

	return false
}

func decoderHasTrailingValue(decoder *json.Decoder) bool {
	var trailing any
	err := decoder.Decode(&trailing)

	return err != io.EOF
}
