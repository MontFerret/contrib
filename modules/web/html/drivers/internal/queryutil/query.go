package queryutil

import "strings"

type Kind string

const (
	CSS   Kind = "css"
	XPath Kind = "xpath"
)

func Parse(value string) Kind {
	switch strings.ToLower(value) {
	case "css":
		return CSS
	case "xpath":
		return XPath
	default:
		return ""
	}
}
