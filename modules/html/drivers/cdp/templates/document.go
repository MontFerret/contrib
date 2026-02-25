package templates

import "github.com/MontFerret/contrib/modules/html/drivers/cdp/eval"

const domReady = `() => {
if (document.readyState === 'complete') {
	return true;
}

return null;
}`

func DOMReady() *eval.Function {
	return eval.F(domReady)
}

const getTitle = `() => document.title`

func GetTitle() *eval.Function {
	return eval.F(getTitle)
}

const getDocument = `() => document`

func GetDocument() *eval.Function {
	return eval.F(getDocument)
}
