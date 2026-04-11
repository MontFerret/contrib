package templates

import (
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const addEventListener = `(rootTarget, eventName, bindingName, config) => {
	const registryKey = '__ferretDomEventHandlers';
	const registry = globalThis[registryKey] || (globalThis[registryKey] = new WeakMap());
	const settings = config == null ? {} : config;
	const listenerOptions = settings.listener == null ? undefined : settings.listener;
	const delegate = settings.delegate == null ? null : settings.delegate;
	const targetSelector = settings.targetSelector == null ? null : settings.targetSelector;
	const props = Array.isArray(settings.props) ? settings.props : null;
	const maxDepth = Number.isInteger(settings.maxDepth) ? settings.maxDepth : 4;

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

	const serializeSelectedEvent = (event) => {
		const output = { type: event.type };

		for (const key of props) {
			if (key === 'type') {
				continue;
			}

			try {
				output[key] = serialize(event[key], 1, new Set([event]));
			} catch (err) {
				output[key] = null;
			}
		}

		return output;
	};

	const serializeEvent = (event) => {
		if (props != null) {
			return serializeSelectedEvent(event);
		}

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

	const matchesDelegate = (event) => {
		if (delegate == null) {
			return true;
		}

		const eventTarget = event.target;
		let candidate = null;

		if (typeof Element !== 'undefined' && eventTarget instanceof Element) {
			candidate = eventTarget;
		} else if (eventTarget != null && eventTarget.parentElement != null) {
			candidate = eventTarget.parentElement;
		}

		if (candidate == null || typeof candidate.closest !== 'function') {
			return false;
		}

		const match = candidate.closest(delegate);

		if (match == null) {
			return false;
		}

		if (typeof Document !== 'undefined' && rootTarget instanceof Document) {
			return true;
		}

		if (match === rootTarget) {
			return true;
		}

		return typeof rootTarget.contains === 'function' && rootTarget.contains(match);
	};

	if (delegate != null && typeof rootTarget.querySelector === 'function') {
		rootTarget.querySelector(delegate);
	}

	let attachTarget = rootTarget;

	if (targetSelector != null) {
		if (typeof rootTarget.querySelector !== 'function') {
			throw new Error('event target does not support querySelector');
		}

		attachTarget = rootTarget.querySelector(targetSelector);

		if (attachTarget == null) {
			throw new Error('failed to resolve event target by selector: ' + targetSelector);
		}
	}

	const handler = (event) => {
		if (!matchesDelegate(event)) {
			return;
		}

		globalThis[bindingName](JSON.stringify(serializeEvent(event)));
	};

	let handlers = registry.get(rootTarget);

	if (handlers == null) {
		handlers = {};
		registry.set(rootTarget, handlers);
	}

	handlers[bindingName] = {
		handler,
		eventName,
		listenerOptions,
		target: attachTarget
	};

	attachTarget.addEventListener(eventName, handler, listenerOptions);
}`

const removeEventListener = `(rootTarget, bindingName) => {
	const registryKey = '__ferretDomEventHandlers';
	const registry = globalThis[registryKey];

	if (registry == null) {
		return;
	}

	const handlers = registry.get(rootTarget);

	if (handlers == null) {
		return;
	}

	const entry = handlers[bindingName];

	if (entry == null) {
		return;
	}

	entry.target.removeEventListener(entry.eventName, entry.handler, entry.listenerOptions);

	delete handlers[bindingName];

	if (Object.keys(handlers).length === 0) {
		registry.delete(rootTarget);
	}
}`

func AddEventListener(id cdpruntime.RemoteObjectID, eventName runtime.String, bindingName string, config runtime.Map) *eval.Function {
	return eval.F(addEventListener).
		WithArgRef(id).
		WithArgValue(eventName).
		WithArg(bindingName).
		WithArgValue(orNone(config))
}

func RemoveEventListener(id cdpruntime.RemoteObjectID, bindingName string) *eval.Function {
	return eval.F(removeEventListener).
		WithArgRef(id).
		WithArg(bindingName)
}

func orNone(options runtime.Map) runtime.Value {
	if options == nil {
		return runtime.None
	}

	return options
}
