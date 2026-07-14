package templates

import (
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const getValue = `(el) => {
	return el.value
}`

func GetValue(id cdpruntime.RemoteObjectID) *eval.Function {
	return eval.F(getValue).WithArgRef(id)
}

const getTextContent = `(el) => {
	return el.textContent ?? "";
}`

func GetTextContent(id cdpruntime.RemoteObjectID) *eval.Function {
	return eval.F(getTextContent).WithArgRef(id)
}

const getDOMProperty = `(el, name) => {
	const value = el[name];

	if (value == null) {
		return undefined;
	}

	if (typeof value === "function") {
		return undefined;
	}

	if (typeof value !== "object") {
		return value;
	}

	if (typeof Node !== "undefined" && value instanceof Node) {
		return value;
	}

	if (Array.isArray(value)) {
		return value;
	}

	if (typeof NodeList !== "undefined" && value instanceof NodeList) {
		return Array.from(value);
	}

	if (typeof HTMLCollection !== "undefined" && value instanceof HTMLCollection) {
		return Array.from(value);
	}

	if (typeof value.length === "number" && typeof value.item === "function") {
		return Array.from(value);
	}

	return undefined;
}`

func GetDOMProperty(id cdpruntime.RemoteObjectID, name runtime.String) *eval.Function {
	return eval.F(getDOMProperty).WithArgRef(id).WithArgValue(name)
}

const setValue = `(el, value) => {
	el.value = value
}`

func SetValue(id cdpruntime.RemoteObjectID, value runtime.Value) *eval.Function {
	return eval.F(setValue).WithArgRef(id).WithArgValue(value)
}

const setTextContent = `(el, value) => {
	el.textContent = value;
}`

func SetTextContent(id cdpruntime.RemoteObjectID, value runtime.String) *eval.Function {
	return eval.F(setTextContent).WithArgRef(id).WithArgValue(value)
}

const setDOMProperty = `(el, name, value) => {
	el[name] = value;
}`

func SetDOMProperty(id cdpruntime.RemoteObjectID, name runtime.String, value runtime.Value) *eval.Function {
	return eval.F(setDOMProperty).WithArgRef(id).WithArgValue(name).WithArgValue(value)
}
