package templates

import (
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const addEventListener = `(target, eventName, bindingName, options) => {
	const registryKey = '__ferretDomEventHandlers';
	const maxDepth = 4;
	const listenerOptions = options == null ? undefined : options;
	const registry = globalThis[registryKey] || (globalThis[registryKey] = new WeakMap());

	const isHostObject = (value) => {
		if (typeof Window !== 'undefined' && value instanceof Window) {
			return true;
		}

		if (typeof Document !== 'undefined' && value instanceof Document) {
			return true;
		}

		if (typeof Node !== 'undefined' && value instanceof Node) {
			return true;
		}

		if (typeof EventTarget !== 'undefined' && value instanceof EventTarget) {
			return true;
		}

		return false;
	};

	const isPlainObject = (value) => {
		if (Object.prototype.toString.call(value) !== '[object Object]') {
			return false;
		}

		const prototype = Object.getPrototypeOf(value);

		return prototype === Object.prototype || prototype === null;
	};

	const serialize = (value, depth, stack) => {
		if (value == null) {
			return null;
		}

		if (depth > maxDepth) {
			return null;
		}

		switch (typeof value) {
			case 'string':
			case 'boolean':
				return value;
			case 'number':
				return Number.isFinite(value) ? value : null;
			case 'undefined':
			case 'function':
			case 'symbol':
			case 'bigint':
				return null;
		}

		if (isHostObject(value)) {
			return null;
		}

		if (stack.has(value)) {
			return null;
		}

		if (Array.isArray(value)) {
			stack.add(value);

			const output = value.map((item) => serialize(item, depth + 1, stack));

			stack.delete(value);

			return output;
		}

		if (!isPlainObject(value)) {
			return null;
		}

		stack.add(value);

		const output = {};

		for (const key of Object.keys(value)) {
			output[key] = serialize(value[key], depth + 1, stack);
		}

		stack.delete(value);

		return output;
	};

	const serializeEvent = (event) => {
		const output = { type: event.type };
		const keys = new Set(['type']);
		let current = event;

		while (current != null) {
			for (const key of Object.getOwnPropertyNames(current)) {
				if (keys.has(key)) {
					continue;
				}

				keys.add(key);

				try {
					output[key] = serialize(event[key], 1, new Set([event]));
				} catch (err) {
					output[key] = null;
				}
			}

			if (typeof Event !== 'undefined' && current === Event.prototype) {
				break;
			}

			current = Object.getPrototypeOf(current);
		}

		return output;
	};

	const handler = (event) => {
		globalThis[bindingName](JSON.stringify(serializeEvent(event)));
	};

	let handlers = registry.get(target);

	if (handlers == null) {
		handlers = {};
		registry.set(target, handlers);
	}

	handlers[bindingName] = handler;

	target.addEventListener(eventName, handler, listenerOptions);
}`

const removeEventListener = `(target, eventName, bindingName, options) => {
	const registryKey = '__ferretDomEventHandlers';
	const listenerOptions = options == null ? undefined : options;
	const registry = globalThis[registryKey];

	if (registry == null) {
		return;
	}

	const handlers = registry.get(target);

	if (handlers == null) {
		return;
	}

	const handler = handlers[bindingName];

	if (handler == null) {
		return;
	}

	target.removeEventListener(eventName, handler, listenerOptions);

	delete handlers[bindingName];

	if (Object.keys(handlers).length === 0) {
		registry.delete(target);
	}
}`

func AddEventListener(id cdpruntime.RemoteObjectID, eventName runtime.String, bindingName string, options runtime.Map) *eval.Function {
	return eval.F(addEventListener).
		WithArgRef(id).
		WithArgValue(eventName).
		WithArg(bindingName).
		WithArgValue(orNone(options))
}

func RemoveEventListener(id cdpruntime.RemoteObjectID, eventName runtime.String, bindingName string, options runtime.Map) *eval.Function {
	return eval.F(removeEventListener).
		WithArgRef(id).
		WithArgValue(eventName).
		WithArg(bindingName).
		WithArgValue(orNone(options))
}

func orNone(options runtime.Map) runtime.Value {
	if options == nil {
		return runtime.None
	}

	return options
}
