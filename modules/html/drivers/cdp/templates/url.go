package templates

import "github.com/MontFerret/contrib/modules/html/drivers/cdp/eval"

const getURL = `() => window.location.toString()`

func GetURL() *eval.Function {
	return eval.F(getURL)
}
