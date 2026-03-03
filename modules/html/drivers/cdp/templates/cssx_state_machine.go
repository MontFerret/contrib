package templates

const cssxStateMachine = `(el) => {
const ops = %s;

const isNode = (v) => v != null && typeof v === "object" && typeof v.nodeType === "number";
const toArray = (v) => {
	if (Array.isArray(v)) {
		return v.filter((i) => i != null);
	}

	if (v == null) {
		return [];
	}

	return [v];
};
const toNodes = (v) => toArray(v).filter(isNode);
const firstNode = (v) => {
	if (isNode(v)) {
		return v;
	}

	const nodes = toNodes(v);
	return nodes.length > 0 ? nodes[0] : null;
};
const queryAll = (root, selector) => {
	try {
		if (root == null || typeof root.querySelectorAll !== "function") {
			return [];
		}

		return Array.from(root.querySelectorAll(selector));
	} catch (_) {
		return [];
	}
};
const textOf = (v) => {
	if (v == null) {
		return "";
	}

	if (Array.isArray(v)) {
		return v.map((i) => textOf(i)).join("");
	}

	if (isNode(v)) {
		return v.textContent ?? "";
	}

	return String(v);
};
const normalizeSpace = (v) => String(v ?? "").replace(/\s+/g, " ").trim();
const containsNode = (ancestor, node) => {
	try {
		return ancestor === node || (ancestor != null && typeof ancestor.contains === "function" && ancestor.contains(node));
	} catch (_) {
		return false;
	}
};
const dedupeNodes = (nodes) => {
	const out = [];
	const seen = new Set();

	for (const node of toNodes(nodes)) {
		if (seen.has(node)) {
			continue;
		}

		seen.add(node);
		out.push(node);
	}

	return out;
};
const asURL = (value) => {
	const input = value == null ? "" : String(value);

	if (input === "") {
		return null;
	}

	try {
		return new URL(input, document.baseURI).href;
	} catch (_) {
		return null;
	}
};
const parseURL = (value) => {
	const input = value == null ? "" : String(value);

	if (input === "") {
		return null;
	}

	try {
		const u = new URL(input, document.baseURI);
		return {
			href: u.href,
			protocol: u.protocol,
			username: u.username,
			password: u.password,
			host: u.host,
			hostname: u.hostname,
			port: u.port,
			pathname: u.pathname,
			search: u.search,
			hash: u.hash,
			origin: u.origin
		};
	} catch (_) {
		return null;
	}
};
const toNumber = (value) => {
	const str = textOf(value).trim();

	if (str === "") {
		return null;
	}

	const normalized = str.
		replace(/[^\d+\-.,eE]/g, "").
		replace(/,(?=\d{3}\b)/g, "").
		replace(",", ".");
	const out = Number(normalized);

	return Number.isFinite(out) ? out : null;
};
const toDate = (value) => {
	const str = textOf(value).trim();

	if (str === "") {
		return null;
	}

	const date = new Date(str);

	return Number.isNaN(date.getTime()) ? null : date.toISOString();
};
const applyCall = (name, args, values) => {
	const input = values.length > 0 ? values[values.length - 1] : null;

	switch (name) {
		case ":first": {
			const arr = toArray(input);
			return arr.length > 0 ? arr[0] : null;
		}
		case ":last": {
			const arr = toArray(input);
			return arr.length > 0 ? arr[arr.length - 1] : null;
		}
		case ":nth": {
			const arr = toArray(input);
			const idx = Number(args[0]);

			if (!Number.isInteger(idx) || idx < 0 || idx >= arr.length) {
				return null;
			}

			return arr[idx];
		}
		case ":take": {
			const arr = toArray(input);
			const n = Number(args[0]);

			if (!Number.isFinite(n) || n <= 0) {
				return [];
			}

			return arr.slice(0, Math.trunc(n));
		}
		case ":skip": {
			const arr = toArray(input);
			const n = Number(args[0]);

			if (!Number.isFinite(n) || n <= 0) {
				return arr;
			}

			return arr.slice(Math.trunc(n));
		}
		case ":slice": {
			const arr = toArray(input);
			const start = Number(args[0]);
			const count = Number(args[1]);

			if (!Number.isFinite(start) || !Number.isFinite(count) || count <= 0) {
				return [];
			}

			return arr.slice(Math.trunc(start), Math.trunc(start + count));
		}
		case ":within": {
			const scope = values.length > 1 ? values[0] : null;
			const scopedNodes = toNodes(scope);

			if (scopedNodes.length === 0) {
				return [];
			}

			if (typeof input === "string") {
				const out = [];

				for (const node of scopedNodes) {
					out.push(...queryAll(node, input));
				}

				return dedupeNodes(out);
			}

			const nodeValues = toNodes(input);

			if (nodeValues.length > 0) {
				return nodeValues.filter((node) => scopedNodes.some((scopeNode) => containsNode(scopeNode, node)));
			}

			return input;
		}
		case ":parent": {
			const node = firstNode(input);
			return node != null ? node.parentElement : null;
		}
		case ":closest": {
			const candidates = toNodes(values.length > 1 ? values[0] : []);
			const node = firstNode(input);

			if (node == null) {
				return null;
			}

			if (candidates.length === 0) {
				return node.parentElement;
			}

			const cset = new Set(candidates);
			let cursor = node;

			for (; cursor != null; cursor = cursor.parentElement) {
				if (cset.has(cursor)) {
					return cursor;
				}
			}

			return null;
		}
		case ":children": {
			const node = firstNode(input);

			if (node == null) {
				return [];
			}

			const children = Array.from(node.children ?? []);
			const candidates = toNodes(values.length > 1 ? values[0] : []);

			if (candidates.length === 0) {
				return children;
			}

			const cset = new Set(candidates);
			return children.filter((child) => cset.has(child));
		}
		case ":next": {
			const node = firstNode(input);

			if (node == null) {
				return null;
			}

			const next = node.nextElementSibling;

			if (next == null) {
				return null;
			}

			const candidates = toNodes(values.length > 1 ? values[0] : []);

			if (candidates.length === 0) {
				return next;
			}

			const cset = new Set(candidates);
			return cset.has(next) ? next : null;
		}
		case ":prev": {
			const node = firstNode(input);

			if (node == null) {
				return null;
			}

			const prev = node.previousElementSibling;

			if (prev == null) {
				return null;
			}

			const candidates = toNodes(values.length > 1 ? values[0] : []);

			if (candidates.length === 0) {
				return prev;
			}

			const cset = new Set(candidates);
			return cset.has(prev) ? prev : null;
		}
		case ":exists":
			return toArray(input).length > 0;
		case ":empty":
			return toArray(input).length === 0;
		case ":has": {
			const node = firstNode(input);

			if (node == null) {
				return false;
			}

			const candidates = toNodes(values.length > 1 ? values[0] : []);
			return candidates.some((candidate) => containsNode(node, candidate));
		}
		case ":matches": {
			const node = firstNode(input);

			if (node == null) {
				return false;
			}

			const candidates = toNodes(values.length > 1 ? values[0] : []);
			return candidates.some((candidate) => candidate === node);
		}
		case ":count":
			return toArray(input).length;
		case ":indexOf": {
			const list = toArray(values.length > 1 ? values[0] : []);
			const item = firstNode(input) ?? input;
			return list.indexOf(item);
		}
		case ":len": {
			if (typeof input === "string" || Array.isArray(input)) {
				return input.length;
			}

			if (input == null) {
				return 0;
			}

			return toArray(input).length;
		}
		case ":text": {
			const node = firstNode(input);

			if (node == null) {
				return null;
			}

			return node.textContent ?? "";
		}
		case ":texts":
			return toNodes(input).map((node) => node.textContent ?? "");
		case ":ownText": {
			const node = firstNode(input);

			if (node == null) {
				return null;
			}

			let out = "";

			for (const child of Array.from(node.childNodes ?? [])) {
				if (child.nodeType === 3) {
					out += child.textContent ?? "";
				}
			}

			return out;
		}
		case ":normalize":
			return normalizeSpace(textOf(input));
		case ":trim":
			return textOf(input).trim();
		case ":join": {
			const arr = toArray(input).map((i) => textOf(i));
			const sep = String(args[0]);
			return arr.join(sep);
		}
		case ":attr": {
			const node = firstNode(input);
			return node != null && typeof node.getAttribute === "function" ? node.getAttribute(String(args[0])) : null;
		}
		case ":attrs": {
			const name = String(args[0]);
			return toNodes(input).map((node) => (typeof node.getAttribute === "function" ? node.getAttribute(name) : null));
		}
		case ":prop": {
			const node = firstNode(input);
			return node != null ? node[String(args[0])] : null;
		}
		case ":html": {
			const node = firstNode(input);
			return node != null ? (node.innerHTML ?? null) : null;
		}
		case ":outerHtml": {
			const node = firstNode(input);
			return node != null ? (node.outerHTML ?? null) : null;
		}
		case ":value": {
			const node = firstNode(input);
			return node != null ? (node.value ?? null) : null;
		}
		case ":absUrl":
			return asURL(Array.isArray(input) ? input[0] : input);
		case ":url": {
			const node = firstNode(input);

			if (node == null || typeof node.getAttribute !== "function") {
				return null;
			}

			return asURL(node.getAttribute(String(args[0])));
		}
		case ":parseUrl":
			return parseURL(Array.isArray(input) ? input[0] : input);
		case ":filter": {
			const predicate = values.length > 1 ? values[0] : true;
			const source = values.length > 1 ? values[1] : input;
			const arr = toArray(source);

			if (Array.isArray(predicate)) {
				const pset = new Set(predicate);
				return arr.filter((item) => pset.has(item));
			}

			if (typeof predicate === "boolean") {
				return predicate ? arr : [];
			}

			if (typeof predicate === "string") {
				return arr.filter((item) => textOf(item).includes(predicate));
			}

			if (predicate == null) {
				return [];
			}

			return arr.filter((item) => Boolean(item));
		}
		case ":withAttr": {
			const attr = String(args[0]);
			return toNodes(input).filter((node) => typeof node.hasAttribute === "function" && node.hasAttribute(attr));
		}
		case ":withText": {
			const needle = String(args[0]);
			return toNodes(input).filter((node) => (node.textContent ?? "").includes(needle));
		}
		case ":dedupeByAttr": {
			const attr = String(args[0]);
			const seen = new Set();
			const out = [];

			for (const node of toNodes(input)) {
				if (typeof node.getAttribute !== "function") {
					continue;
				}

				const value = node.getAttribute(attr);

				if (seen.has(value)) {
					continue;
				}

				seen.add(value);
				out.push(node);
			}

			return out;
		}
		case ":dedupeByText": {
			const seen = new Set();
			const out = [];

			for (const node of toNodes(input)) {
				const key = normalizeSpace(node.textContent ?? "");

				if (seen.has(key)) {
					continue;
				}

				seen.add(key);
				out.push(node);
			}

			return out;
		}
		case ":replace": {
			const pattern = String(args[0]);
			const replacement = String(args[1]);
			const source = textOf(input);
			let rx = null;

			try {
				rx = new RegExp(pattern, "g");
			} catch (_) {}

			if (rx != null) {
				return source.replace(rx, replacement);
			}

			return source.split(pattern).join(replacement);
		}
		case ":regex": {
			const pattern = String(args[0]);
			const group = args.length > 1 ? Number(args[1]) : 0;
			const source = textOf(input);
			let rx = null;

			try {
				rx = new RegExp(pattern);
			} catch (_) {
				return null;
			}

			const match = rx.exec(source);

			if (match == null) {
				return null;
			}

			const idx = Number.isInteger(group) ? group : 0;
			return match[idx] ?? null;
		}
		case ":toNumber":
			return toNumber(input);
		case ":toDate":
			return toDate(input);
		default:
			return [];
	}
};

const stack = [];

for (const op of ops) {
	if (op.kind === "select") {
		stack.push(queryAll(el, op.selector));
		continue;
	}

	if (op.kind !== "call") {
		continue;
	}

	let consume = op.arity;

	if (consume === 0 && stack.length > 0) {
		consume = 1;
	}

	if (consume > stack.length) {
		stack.push([]);
		continue;
	}

	const values = consume > 0 ? stack.splice(stack.length - consume, consume) : [];
	stack.push(applyCall(op.name, op.args ?? [], values));
}

if (stack.length === 0) {
	return [];
}

const result = stack[stack.length - 1];

if (Array.isArray(result)) {
	return result;
}

if (result == null) {
	return [];
}

return [result];
}`
