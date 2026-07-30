package templates

import (
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const getURL = `() => window.location.toString()`

func GetURL() *eval.Function {
	return eval.F(getURL)
}

const getDocumentURL = `() => document.URL`

func GetDocumentURL() *eval.Function {
	return eval.F(getDocumentURL)
}

const getBaseURL = `() => document.baseURI`

func GetBaseURL() *eval.Function {
	return eval.F(getBaseURL)
}

const resolveURL = `(url) => new URL(url, document.baseURI).href`

func ResolveURL(value runtime.String) *eval.Function {
	return eval.F(resolveURL).WithArgValue(value)
}
