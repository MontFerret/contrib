package common

import "strings"

type QueryKind string

const (
	CSSQuery   QueryKind = "css"
	XPathQuery QueryKind = "xpath"
)

func ToQueryKind(s string) QueryKind {
	switch strings.ToLower(s) {
	case "css":
		return CSSQuery
	case "xpath":
		return XPathQuery
	default:
		return ""
	}
}
