package templates

import (
	"fmt"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/contrib/modules/html/drivers/cdp/eval"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const (
	waitExistenceFragment = `(el, op, ...args) => {
	const actual = %s; // check

	// presence 
	if (op === 0) {
		if (actual != null) {
			return true;
		}
	} else {
		if (actual == null) {
			return true;
		}
	}
	
	// null means we need to repeat
	return null;
}`

	waitEqualityFragment = `(el, expected, op, ...args) => {
	const actual = %s; // check

	// presence 
	if (op === 0) {
		if (actual === expected) {
			return true;
		}
	} else {
		if (actual !== expected) {
			return true;
		}
	}
	
	// null means we need to repeat
	return null;
}`

	waitExistenceBySelectorFragment = `(el, selector, op, ...args) => {
	// selector
	%s

	if (found == null) {
		return false;
	}

	const actual = %s; // check

	// presence 
	if (op === 0) {
		if (actual != null) {
			return true;
		}
	} else {
		if (actual == null) {
			return true;
		}
	}
	
	// null means we need to repeat
	return null;
}`

	waitEqualityBySelectorFragment = `(el, selector, expected, op, ...args) => {
	// selector
	%s

	if (found == null) {
		return false;
	}

	const actual = %s; // check

	// presence 
	if (op === 0) {
		if (actual === expected) {
			return true;
		}
	} else {
		if (actual !== expected) {
			return true;
		}
	}
	
	// null means we need to repeat
	return null;
}`

	waitExistenceBySelectorAllFragment = `(el, selector, op, ...args) => {
	// selector
	%s
	
	if (found == null || !found || found.length === 0) {
		return false;
	}
	
	let resultCount = 0;
	
	found.forEach((el) => {
		let actual = %s; // check
	
		// when
		// presence 
		if (op === 0) {
			if (actual != null) {
				resultCount++;
			}
		} else {
			if (actual == null) {
				resultCount++;
			}
		}
	});
	
	if (resultCount === found.length) {
		return true;
	}
	
	// null means we need to repeat
	return null;
}`

	waitEqualityBySelectorAllFragment = `(el, selector, expected, op, ...args) => {
	// selector
	%s
	
	if (found == null || !found || found.length === 0) {
		return false;
	}
	
	let resultCount = 0;

	found.forEach((el) => {
		let actual = %s; // check
	
		// when
		// presence 
		if (op === 0) {
			if (actual === expected) {
				resultCount++;
			}
		} else {
			if (actual !== expected) {
				resultCount++;
			}
		}
	});
	
	if (resultCount === found.length) {
		return true;
	}
	
	// null means we need to repeat
	return null;
}`
)

func partialWaitExistence(id cdpruntime.RemoteObjectID, when drivers.WaitEvent, fragment string) *eval.Function {
	return eval.F(fmt.Sprintf(waitExistenceFragment, fragment)).
		WithArgRef(id).
		WithArg(int(when))
}

func partialWaitEquality(id cdpruntime.RemoteObjectID, expected runtime.Value, when drivers.WaitEvent, fragment string) *eval.Function {
	return eval.F(fmt.Sprintf(waitEqualityFragment, fragment)).
		WithArgRef(id).
		WithArgValue(expected).
		WithArg(int(when))
}

func partialWaitExistenceBySelector(id cdpruntime.RemoteObjectID, selector drivers.QuerySelector, when drivers.WaitEvent, fragment string) *eval.Function {
	var tmpl string

	if selector.Kind == drivers.CSSSelector {
		tmpl = fmt.Sprintf(waitExistenceBySelectorFragment, queryCSSSelectorFragment, fragment)
	} else {
		tmpl = fmt.Sprintf(waitExistenceBySelectorFragment, xpathAsElementFragment, fragment)
	}

	return eval.F(tmpl).
		WithArgRef(id).
		WithArgSelector(selector).
		WithArg(int(when))
}

func partialWaitEqualityBySelector(id cdpruntime.RemoteObjectID, selector drivers.QuerySelector, expected runtime.Value, when drivers.WaitEvent, fragment string) *eval.Function {
	var tmpl string

	if selector.Kind == drivers.CSSSelector {
		tmpl = fmt.Sprintf(waitEqualityBySelectorFragment, queryCSSSelectorFragment, fragment)
	} else {
		tmpl = fmt.Sprintf(waitEqualityBySelectorFragment, xpathAsElementFragment, fragment)
	}

	return eval.F(tmpl).
		WithArgRef(id).
		WithArgSelector(selector).
		WithArgValue(expected).
		WithArg(int(when))
}

func partialWaitExistenceBySelectorAll(id cdpruntime.RemoteObjectID, selector drivers.QuerySelector, when drivers.WaitEvent, fragment string) *eval.Function {
	var tmpl string

	if selector.Kind == drivers.CSSSelector {
		tmpl = fmt.Sprintf(waitExistenceBySelectorAllFragment, queryCSSSelectorAllFragment, fragment)
	} else {
		tmpl = fmt.Sprintf(waitExistenceBySelectorAllFragment, xpathAsElementArrayFragment, fragment)
	}

	return eval.F(tmpl).
		WithArgRef(id).
		WithArgSelector(selector).
		WithArg(int(when))
}

func partialWaitEqualityBySelectorAll(id cdpruntime.RemoteObjectID, selector drivers.QuerySelector, expected runtime.Value, when drivers.WaitEvent, fragment string) *eval.Function {
	var tmpl string

	if selector.Kind == drivers.CSSSelector {
		tmpl = fmt.Sprintf(waitEqualityBySelectorAllFragment, queryCSSSelectorAllFragment, fragment)
	} else {
		tmpl = fmt.Sprintf(waitEqualityBySelectorAllFragment, xpathAsElementArrayFragment, fragment)
	}

	return eval.F(tmpl).
		WithArgRef(id).
		WithArgSelector(selector).
		WithArgValue(expected).
		WithArg(int(when))
}

const waitForElementByCSSFragment = `el.querySelector(args[0])`

var waitForElementByXPathFragment = fmt.Sprintf(`(() => {
const selector = args[0];

%s

return found;
})()`, xpathAsElementFragment)

func WaitForElement(id cdpruntime.RemoteObjectID, selector drivers.QuerySelector, when drivers.WaitEvent) *eval.Function {
	var tmpl string

	if selector.Kind == drivers.CSSSelector {
		tmpl = waitForElementByCSSFragment
	} else {
		tmpl = waitForElementByXPathFragment
	}

	return partialWaitExistence(id, when, tmpl).WithArgSelector(selector)
}

const waitForElementAllByCSSFragment = `(function() {
const elements = el.querySelector(args[0]);

return elements.length;
})()`

var waitForElementAllByXPathFragment = fmt.Sprintf(`(function() {
const selector = args[0];

%s

return found;
})()`, xpathAsElementArrayFragment)

func WaitForElementAll(id cdpruntime.RemoteObjectID, selector drivers.QuerySelector, when drivers.WaitEvent) *eval.Function {
	var tmpl string

	if selector.Kind == drivers.CSSSelector {
		tmpl = waitForElementAllByCSSFragment
	} else {
		tmpl = waitForElementAllByXPathFragment
	}

	return partialWaitEquality(id, runtime.ZeroInt, when, tmpl).WithArgSelector(selector)
}

const waitForClassFragment = `el.className.split(' ').find(i => i === args[0]);`

func WaitForClass(id cdpruntime.RemoteObjectID, class runtime.String, when drivers.WaitEvent) *eval.Function {
	return partialWaitExistence(id, when, waitForClassFragment).WithArgValue(class)
}

const waitForClassBySelectorFragment = `found.className.split(' ').find(i => i === args[0]);`

func WaitForClassBySelector(id cdpruntime.RemoteObjectID, selector drivers.QuerySelector, class runtime.String, when drivers.WaitEvent) *eval.Function {
	return partialWaitExistenceBySelector(id, selector, when, waitForClassBySelectorFragment).WithArgValue(class)
}

func WaitForClassBySelectorAll(id cdpruntime.RemoteObjectID, selector drivers.QuerySelector, class runtime.String, when drivers.WaitEvent) *eval.Function {
	return partialWaitExistenceBySelectorAll(id, selector, when, waitForClassFragment).WithArgValue(class)
}

const waitForAttributeFragment = `el.getAttribute(args[0])`

func WaitForAttribute(id cdpruntime.RemoteObjectID, name runtime.String, expected runtime.Value, when drivers.WaitEvent) *eval.Function {
	return partialWaitEquality(id, expected, when, waitForAttributeFragment).WithArgValue(name)
}

const waitForAttributeBySelectorFragment = `found.getAttribute(args[0])`

func WaitForAttributeBySelector(id cdpruntime.RemoteObjectID, selector drivers.QuerySelector, name runtime.Value, expected runtime.Value, when drivers.WaitEvent) *eval.Function {
	return partialWaitEqualityBySelector(id, selector, expected, when, waitForAttributeBySelectorFragment).WithArgValue(name)
}

func WaitForAttributeBySelectorAll(id cdpruntime.RemoteObjectID, selector drivers.QuerySelector, name runtime.String, expected runtime.Value, when drivers.WaitEvent) *eval.Function {
	return partialWaitEqualityBySelectorAll(id, selector, expected, when, waitForAttributeFragment).WithArgValue(name)
}

const waitForStyleFragment = `(function getStyles() {
	const styles = window.getComputedStyle(el);
	return styles[args[0]];
})()`

func WaitForStyle(id cdpruntime.RemoteObjectID, name runtime.String, expected runtime.Value, when drivers.WaitEvent) *eval.Function {
	return partialWaitEquality(id, expected, when, waitForStyleFragment).WithArgValue(name)
}

const waitForStyleBySelectorFragment = `(function getStyles() {
	const styles = window.getComputedStyle(found);
	return styles[args[0]];
})()`

func WaitForStyleBySelector(id cdpruntime.RemoteObjectID, selector drivers.QuerySelector, name runtime.String, expected runtime.Value, when drivers.WaitEvent) *eval.Function {
	return partialWaitEqualityBySelector(id, selector, expected, when, waitForStyleBySelectorFragment).WithArgValue(name)
}

func WaitForStyleBySelectorAll(id cdpruntime.RemoteObjectID, selector drivers.QuerySelector, name runtime.String, expected runtime.Value, when drivers.WaitEvent) *eval.Function {
	return partialWaitEqualityBySelectorAll(id, selector, expected, when, waitForStyleFragment).WithArgValue(name)
}
