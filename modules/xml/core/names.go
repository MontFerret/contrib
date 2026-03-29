package core

import (
	"encoding/xml"
	"strings"
	"unicode"
)

func xmlNameToString(name xml.Name) string {
	if name.Space == "" {
		return name.Local
	}

	if name.Local == "" {
		return name.Space
	}

	return name.Space + ":" + name.Local
}

func xmlNameFromString(name string) (xml.Name, error) {
	if name == "" {
		return xml.Name{}, newXMLError("XML names must not be empty")
	}

	if strings.IndexFunc(name, unicode.IsSpace) >= 0 {
		return xml.Name{}, newXMLErrorf("XML name %q must not contain whitespace", name)
	}

	// encoding/xml preserves raw prefixes when the full qualified name is kept
	// in Local; splitting into Space/Local rewrites xmlns-prefixed output.
	return xml.Name{Local: name}, nil
}
