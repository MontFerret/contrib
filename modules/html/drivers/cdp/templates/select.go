package templates

import (
	"fmt"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/contrib/modules/html/drivers/cdp/eval"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const selectFragment = `
	if (found.nodeName.toLowerCase() !== 'select') {
		throw new Error('element is not a <select> element.');
	}

	const options = Array.from(found.options);

	found.value = undefined;

	for (var option of options) {
		option.selected = runtime.includes(option.value);
	
		if (option.selected && !found.multiple) {
			break;
		}
	}

	found.dispatchEvent(new Event('input', { 'bubbles': true }));
	found.dispatchEvent(new Event('change', { 'bubbles': true }));
	
	return options.filter(option => option.selected).map(option => option.value);
`

var selekt = fmt.Sprintf(`(el, values) => {
const found = el;

%s
}`, selectFragment)

func Select(id cdpruntime.RemoteObjectID, inputs runtime.List) *eval.Function {
	return eval.F(selekt).WithArgRef(id).WithArgValue(inputs)
}

var (
	selectByCSSSelector = fmt.Sprintf(`(el, selector, values) => {
	%s
	
	%s

	%s
}`, queryCSSSelectorFragment, notFoundErrorFragment, selectFragment)

	selectByXPathSelector = fmt.Sprintf(`(el, selector, values) => {
	%s
	
	%s

	%s
}`, xpathAsElementFragment, notFoundErrorFragment, selectFragment)
)

func SelectBySelector(id cdpruntime.RemoteObjectID, selector drivers.QuerySelector, inputs runtime.List) *eval.Function {
	return toFunction(selector, selectByCSSSelector, selectByXPathSelector).
		WithArgRef(id).
		WithArgSelector(selector).
		WithArgValue(inputs)
}
