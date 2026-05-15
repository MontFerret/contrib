package templates

import (
	"fmt"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const dispatchFormHelpers = `
function nativeValueSetter(el) {
	const proto = el instanceof HTMLTextAreaElement ? HTMLTextAreaElement.prototype : HTMLInputElement.prototype;
	const descriptor = Object.getOwnPropertyDescriptor(proto, "value");

	if (descriptor && descriptor.set) {
		return descriptor.set;
	}

	const fallback = Object.getOwnPropertyDescriptor(HTMLElement.prototype, "value");
	return fallback && fallback.set;
}

function setNativeValue(el, value) {
	const normalized = value == null ? "" : String(value);
	const setter = nativeValueSetter(el);

	if (setter) {
		setter.call(el, normalized);
		return;
	}

	el.value = normalized;
}

function setNativeChecked(el, value) {
	const descriptor = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, "checked");

	if (descriptor && descriptor.set) {
		descriptor.set.call(el, value);
		return;
	}

	el.checked = value;
}

function valuesOf(value) {
	if (Array.isArray(value)) {
		return value.map((item) => String(item));
	}

	return [String(value)];
}

function setControlValue(el, value) {
	const nodeName = el.nodeName.toLowerCase();

	if (nodeName === "select") {
		const values = valuesOf(value);
		const options = Array.from(el.options);

		for (const option of options) {
			option.selected = values.includes(option.value);

			if (option.selected && !el.multiple) {
				break;
			}
		}

		return;
	}

	if ("value" in el) {
		setNativeValue(el, value);
		return;
	}

	throw new Error("element does not support value");
}

function dispatchBubbling(el, eventName, cancelable = false) {
	el.dispatchEvent(new Event(eventName, {
		bubbles: true,
		cancelable
	}));
}

function formFor(el) {
	if (el instanceof HTMLFormElement) {
		return el;
	}

	if (el.form instanceof HTMLFormElement) {
		return el.form;
	}

	throw new Error("element is not a form or form control");
}
`

const dispatchInput = `(el, value) => {
%s
	setControlValue(el, value);
	dispatchBubbling(el, "input");
}`

const dispatchChange = `(el, value, hasValue) => {
%s
	if (hasValue) {
		setControlValue(el, value);
	}

	dispatchBubbling(el, "change");
}`

const dispatchCheck = `(el, action) => {
%s
	if (!(el instanceof HTMLInputElement) || (el.type !== "checkbox" && el.type !== "radio")) {
		throw new Error("element is not a checkbox or radio input");
	}

	const next = action === "toggle" ? !el.checked : action === "check";
	setNativeChecked(el, next);
	dispatchBubbling(el, "input");
	dispatchBubbling(el, "change");
}`

const dispatchSubmit = `(el) => {
%s
	const form = formFor(el);

	if (typeof form.requestSubmit === "function") {
		form.requestSubmit();
		return;
	}

	if (form.dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }))) {
		form.submit();
	}
}`

const dispatchReset = `(el) => {
%s
	const form = formFor(el);
	form.reset();
	dispatchBubbling(form, "reset");
}`

const elementScroll = `(el, mode, opts) => {
	if (mode === "intoView") {
		el.scrollIntoView({
			behavior: opts.behavior,
			block: opts.block,
			inline: opts.inline
		});
		return true;
	}

	const args = {
		left: opts.left,
		top: opts.top,
		behavior: opts.behavior
	};

	if (mode === "by") {
		el.scrollBy(args);
		return true;
	}

	el.scrollTo(args);
	return true;
}`

func DispatchInput(id cdpruntime.RemoteObjectID, value runtime.Value) *eval.Function {
	return eval.F(fmt.Sprintf(dispatchInput, dispatchFormHelpers)).
		WithArgRef(id).
		WithArgValue(value)
}

func DispatchChange(id cdpruntime.RemoteObjectID, value runtime.Value, hasValue bool) *eval.Function {
	return eval.F(fmt.Sprintf(dispatchChange, dispatchFormHelpers)).
		WithArgRef(id).
		WithArgValue(value).
		WithArg(hasValue)
}

func DispatchCheck(id cdpruntime.RemoteObjectID, action string) *eval.Function {
	return eval.F(fmt.Sprintf(dispatchCheck, dispatchFormHelpers)).
		WithArgRef(id).
		WithArg(action)
}

func DispatchSubmit(id cdpruntime.RemoteObjectID) *eval.Function {
	return eval.F(fmt.Sprintf(dispatchSubmit, dispatchFormHelpers)).
		WithArgRef(id)
}

func DispatchReset(id cdpruntime.RemoteObjectID) *eval.Function {
	return eval.F(fmt.Sprintf(dispatchReset, dispatchFormHelpers)).
		WithArgRef(id)
}

func ElementScroll(id cdpruntime.RemoteObjectID, mode string, options drivers.ScrollOptions) *eval.Function {
	return eval.F(elementScroll).
		WithArgRef(id).
		WithArg(mode).
		WithArg(options)
}
