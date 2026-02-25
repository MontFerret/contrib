package templates

import (
	"fmt"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/contrib/modules/html/drivers/cdp/eval"
)

const blur = `(el) => {
	el.blur()
}`

func Blur(id cdpruntime.RemoteObjectID) *eval.Function {
	return eval.F(blur).WithArgRef(id)
}

var (
	blurByCSSSelector = fmt.Sprintf(`
		(el, selector) => {
			const found = el.querySelector(selector);

			%s

			found.blur();
		}
`, notFoundErrorFragment)

	blurByXPathSelector = fmt.Sprintf(`
		(el, selector) => {
			%s

			%s

			found.blur();
		}
`, xpathAsElementFragment, notFoundErrorFragment)
)

func BlurBySelector(id cdpruntime.RemoteObjectID, selector drivers.QuerySelector) *eval.Function {
	var f *eval.Function

	if selector.Kind == drivers.CSSSelector {
		f = eval.F(blurByCSSSelector)
	} else {
		f = eval.F(blurByXPathSelector)
	}

	return f.WithArgRef(id).WithArgSelector(selector)
}
