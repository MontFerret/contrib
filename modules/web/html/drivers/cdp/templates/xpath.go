package templates

import (
	"fmt"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const xpath = `(el, expression, resType) => {
	const unwrap = (item) => {
		return item.nodeType != 2 ? item : item.nodeValue;
	};
	const out = document.evaluate(
		expression,
		el,
		null,
		resType == null ? XPathResult.ANY_TYPE : resType
	);
	let result;

	switch (out.resultType) {
		case XPathResult.UNORDERED_NODE_ITERATOR_TYPE:
		case XPathResult.ORDERED_NODE_ITERATOR_TYPE: {
			result = [];
			let item;

			while ((item = out.iterateNext())) {
				result.push(unwrap(item));
			}

			break;
		}
		case XPathResult.UNORDERED_NODE_SNAPSHOT_TYPE:
		case XPathResult.ORDERED_NODE_SNAPSHOT_TYPE: {
			result = [];

			for (let i = 0; i < out.snapshotLength; i++) {
				const item = out.snapshotItem(i);

				if (item != null) {
					result.push(unwrap(item));
				}
			}
			break;
		}
		case XPathResult.NUMBER_TYPE: {
			result = out.numberValue;
			break;
		}
		case XPathResult.STRING_TYPE: {
			result = out.stringValue;
			break;
		}
		case XPathResult.BOOLEAN_TYPE: {
			result = out.booleanValue;
			break;
		}
		case XPathResult.ANY_UNORDERED_NODE_TYPE:
		case XPathResult.FIRST_ORDERED_NODE_TYPE: {
			const node = out.singleNodeValue;
			
			if (node != null) {
				result = unwrap(node);
			}
			
			break;
		}
		default: {
			break;
		}
	}

	return result;
}
`

var (
	xpathAsElementFragment = fmt.Sprintf(`
const xpath = %s;
const found = xpath(el, selector, XPathResult.FIRST_ORDERED_NODE_TYPE);
`, xpath)

	xpathAsElementArrayFragment = fmt.Sprintf(`
const xpath = %s;
const found = xpath(el, selector, XPathResult.ORDERED_NODE_ITERATOR_TYPE);
`, xpath)
)

func XPath(id cdpruntime.RemoteObjectID, expression runtime.String) *eval.Function {
	return eval.F(xpath).WithArgRef(id).WithArgValue(expression)
}

const xpathOne = `(el, expression) => {
	const unwrap = (item) => {
		return item.nodeType != 2 ? item : item.nodeValue;
	};
	const out = document.evaluate(
		expression,
		el,
		null,
		XPathResult.ANY_TYPE
	);

	switch (out.resultType) {
		case XPathResult.UNORDERED_NODE_ITERATOR_TYPE:
		case XPathResult.ORDERED_NODE_ITERATOR_TYPE: {
			const item = out.iterateNext();
			return item != null ? unwrap(item) : null;
		}
		case XPathResult.UNORDERED_NODE_SNAPSHOT_TYPE:
		case XPathResult.ORDERED_NODE_SNAPSHOT_TYPE: {
			const item = out.snapshotLength > 0 ? out.snapshotItem(0) : null;
			return item != null ? unwrap(item) : null;
		}
		case XPathResult.NUMBER_TYPE:
			return out.numberValue;
		case XPathResult.STRING_TYPE:
			return out.stringValue;
		case XPathResult.BOOLEAN_TYPE:
			return out.booleanValue;
		case XPathResult.ANY_UNORDERED_NODE_TYPE:
		case XPathResult.FIRST_ORDERED_NODE_TYPE: {
			const node = out.singleNodeValue;
			return node != null ? unwrap(node) : null;
		}
		default:
			return null;
	}
}`

func XPathOne(id cdpruntime.RemoteObjectID, expression runtime.String) *eval.Function {
	return eval.F(xpathOne).WithArgRef(id).WithArgValue(expression)
}

const xpathCount = `(el, expression) => {
	const out = document.evaluate(
		expression,
		el,
		null,
		XPathResult.ANY_TYPE
	);

	switch (out.resultType) {
		case XPathResult.UNORDERED_NODE_ITERATOR_TYPE:
		case XPathResult.ORDERED_NODE_ITERATOR_TYPE: {
			let count = 0;

			while (out.iterateNext() != null) {
				count++;
			}

			return count;
		}
		case XPathResult.UNORDERED_NODE_SNAPSHOT_TYPE:
		case XPathResult.ORDERED_NODE_SNAPSHOT_TYPE:
			return out.snapshotLength;
		case XPathResult.NUMBER_TYPE:
		case XPathResult.STRING_TYPE:
		case XPathResult.BOOLEAN_TYPE:
			return 1;
		case XPathResult.ANY_UNORDERED_NODE_TYPE:
		case XPathResult.FIRST_ORDERED_NODE_TYPE:
			return out.singleNodeValue != null ? 1 : 0;
		default:
			return 0;
	}
}`

func XPathCount(id cdpruntime.RemoteObjectID, expression runtime.String) *eval.Function {
	return eval.F(xpathCount).WithArgRef(id).WithArgValue(expression)
}

const xpathExists = `(el, expression) => {
	const out = document.evaluate(
		expression,
		el,
		null,
		XPathResult.ANY_TYPE
	);

	switch (out.resultType) {
		case XPathResult.UNORDERED_NODE_ITERATOR_TYPE:
		case XPathResult.ORDERED_NODE_ITERATOR_TYPE:
			return out.iterateNext() != null;
		case XPathResult.UNORDERED_NODE_SNAPSHOT_TYPE:
		case XPathResult.ORDERED_NODE_SNAPSHOT_TYPE:
			return out.snapshotLength > 0;
		case XPathResult.NUMBER_TYPE:
		case XPathResult.STRING_TYPE:
		case XPathResult.BOOLEAN_TYPE:
			return true;
		case XPathResult.ANY_UNORDERED_NODE_TYPE:
		case XPathResult.FIRST_ORDERED_NODE_TYPE:
			return out.singleNodeValue != null;
		default:
			return false;
	}
}`

func XPathExists(id cdpruntime.RemoteObjectID, expression runtime.String) *eval.Function {
	return eval.F(xpathExists).WithArgRef(id).WithArgValue(expression)
}
