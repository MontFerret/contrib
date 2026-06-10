package templates

const cssxStateMachine = `(el) => {
const ops = %s;

const isNode = (value) => value != null && typeof value === "object" && typeof value.nodeType === "number";
const toSelection = (value) => {
	if (Array.isArray(value)) {
		return value.slice();
	}

	return value == null ? [] : [value];
};
const toNodes = (value) => toSelection(value).filter(isNode);
const queryAll = (root, selector) => {
	try {
		return root != null && typeof root.querySelectorAll === "function"
			? Array.from(root.querySelectorAll(selector))
			: [];
	} catch (_) {
		return [];
	}
};
const matches = (node, selector) => {
	try {
		return node != null && selector !== "" && typeof node.matches === "function" && node.matches(selector);
	} catch (_) {
		return false;
	}
};
const validSelector = (selector) => {
	try {
		document.createDocumentFragment().querySelector(selector);
		return selector !== "";
	} catch (_) {
		return false;
	}
};
const has = (node, selector) => {
	try {
		return node != null && selector !== "" && typeof node.querySelector === "function" && node.querySelector(selector) != null;
	} catch (_) {
		return false;
	}
};
const within = (node, selector) => {
	for (let parent = node?.parentElement ?? null; parent != null; parent = parent.parentElement) {
		if (matches(parent, selector)) {
			return true;
		}
	}

	return false;
};
const textOf = (value) => {
	if (value == null) {
		return "";
	}

	if (isNode(value)) {
		return value.textContent ?? "";
	}

	return String(value);
};
const normalizeSpace = (value) => String(value ?? "").replace(/\s+/g, " ").trim();
const asURL = (value) => {
	const input = value == null ? "" : String(value).trim();
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
	const input = value == null ? "" : String(value).trim();
	if (input === "") {
		return null;
	}

	try {
		const url = new URL(input, document.baseURI);
		return {
			href: url.href,
			protocol: url.protocol,
			username: url.username,
			password: url.password,
			host: url.host,
			hostname: url.hostname,
			port: url.port,
			pathname: url.pathname,
			search: url.search,
			hash: url.hash,
			origin: url.origin
		};
	} catch (_) {
		return null;
	}
};
const toNumber = (value) => {
	const source = textOf(value).trim();
	if (source === "") {
		return null;
	}

	const normalized = source.
		replace(/[^\d+\-.,eE]/g, "").
		replace(/,(?=\d{3}\b)/g, "").
		replace(",", ".");
	const result = Number(normalized);

	return Number.isFinite(result) ? result : null;
};
const toDate = (value) => {
	const source = textOf(value).trim();
	if (source === "") {
		return null;
	}

	const result = new Date(source);
	return Number.isNaN(result.getTime()) ? null : result.toISOString();
};
const cardinality = (name, args, input) => {
	const items = toSelection(input);

	switch (name) {
		case ":first":
			return items.length > 0 ? items[0] : null;
		case ":last":
			return items.length > 0 ? items[items.length - 1] : null;
		case ":nth": {
			const index = Number(args[0]);
			return Number.isInteger(index) && index >= 0 && index < items.length ? items[index] : null;
		}
		default:
			return null;
	}
};
const select = (name, args, input) => {
	const items = toSelection(input);

	switch (name) {
		case ":take": {
			const count = Number(args[0]);
			return Number.isFinite(count) && count > 0 ? items.slice(0, Math.trunc(count)) : [];
		}
		case ":skip": {
			const count = Number(args[0]);
			return Number.isFinite(count) && count > 0 ? items.slice(Math.trunc(count)) : items;
		}
		case ":slice": {
			const start = Number(args[0]);
			const count = Number(args[1]);
			return Number.isFinite(start) && Number.isFinite(count) && count > 0
				? items.slice(Math.trunc(start), Math.trunc(start + count))
				: [];
		}
		case ":compact":
			return items.filter((item) => item != null);
		case ":distinct":
			return Array.from(new Set(items));
		case ":dedupeByAttr": {
			const attr = String(args[0]);
			const seen = new Set();
			const out = [];

			for (const node of toNodes(items)) {
				const value = typeof node.getAttribute === "function" ? node.getAttribute(attr) : null;
				if (!seen.has(value)) {
					seen.add(value);
					out.push(node);
				}
			}

			return out;
		}
		case ":dedupeByText": {
			const seen = new Set();
			const out = [];

			for (const node of toNodes(items)) {
				const value = normalizeSpace(node.textContent ?? "");
				if (!seen.has(value)) {
					seen.add(value);
					out.push(node);
				}
			}

			return out;
		}
		default:
			return [];
	}
};
const appendMatching = (out, node, criterion) => {
	if (node != null && (criterion === "" || matches(node, criterion))) {
		out.push(node);
	}
};
const traverse = (name, args, input) => {
	const criterion = args.length > 0 ? String(args[0]) : "";
	const out = [];

	for (const node of toNodes(input)) {
		switch (name) {
			case ":parent":
				appendMatching(out, node.parentElement, criterion);
				break;
			case ":closest": {
				let current = node;
				while (current != null) {
					if (matches(current, criterion)) {
						out.push(current);
						break;
					}
					current = current.parentElement;
				}
				break;
			}
			case ":children":
				for (const child of Array.from(node.children ?? [])) {
					appendMatching(out, child, criterion);
				}
				break;
			case ":next":
				appendMatching(out, node.nextElementSibling, criterion);
				break;
			case ":prev":
				appendMatching(out, node.previousElementSibling, criterion);
				break;
			case ":siblings":
				for (const sibling of Array.from(node.parentElement?.children ?? [])) {
					if (sibling !== node) {
						appendMatching(out, sibling, criterion);
					}
				}
				break;
		}
	}

	return out;
};
const filter = (name, args, input) => {
	const criterion = String(args[0]);
	const out = [];

	if ((name === ":within" || name === ":has" || name === ":matches" || name === ":not") && !validSelector(criterion)) {
		return out;
	}

	for (const node of toNodes(input)) {
		let keep = false;

		switch (name) {
			case ":within":
				keep = within(node, criterion);
				break;
			case ":has":
				keep = has(node, criterion);
				break;
			case ":matches":
				keep = matches(node, criterion);
				break;
			case ":not":
				keep = !matches(node, criterion);
				break;
			case ":withAttr":
				keep = typeof node.hasAttribute === "function" && node.hasAttribute(criterion);
				break;
			case ":withText":
				keep = (node.textContent ?? "").includes(criterion);
				break;
		}

		if (keep) {
			out.push(node);
		}
	}

	return out;
};
const mapItem = (name, args, input) => {
	switch (name) {
		case ":text":
			return isNode(input) ? (input.textContent ?? "") : null;
		case ":ownText": {
			if (!isNode(input)) {
				return null;
			}

			let out = "";
			for (const child of Array.from(input.childNodes ?? [])) {
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
		case ":attr":
			return isNode(input) && typeof input.getAttribute === "function" ? input.getAttribute(String(args[0])) : null;
		case ":prop":
			return isNode(input) ? (input[String(args[0])] ?? null) : null;
		case ":html":
			return isNode(input) ? (input.innerHTML ?? null) : null;
		case ":outerHtml":
			return isNode(input) ? (input.outerHTML ?? null) : null;
		case ":value":
			return isNode(input) ? (input.value ?? null) : null;
		case ":absUrl":
			return asURL(input);
		case ":url":
			return isNode(input) && typeof input.getAttribute === "function"
				? asURL(input.getAttribute(String(args[0])))
				: null;
		case ":parseUrl":
			return parseURL(input);
		case ":replace": {
			const pattern = String(args[0]);
			const replacement = String(args[1]);
			const source = textOf(input);

			try {
				return source.replace(new RegExp(pattern, "g"), replacement);
			} catch (_) {
				return source.split(pattern).join(replacement);
			}
		}
		case ":regex": {
			const group = args.length > 1 ? Number(args[1]) : 0;
			let match = null;

			try {
				match = new RegExp(String(args[0])).exec(textOf(input));
			} catch (_) {
				return null;
			}

			return match != null && Number.isInteger(group) ? (match[group] ?? null) : null;
		}
		case ":toNumber":
			return toNumber(input);
		case ":toDate":
			return toDate(input);
		default:
			return null;
	}
};
const map = (name, args, input) => toSelection(input).map((item) => item == null ? null : mapItem(name, args, item));
const reduce = (name, args, values, input) => {
	const items = toSelection(input);

	switch (name) {
		case ":exists":
			return items.length > 0;
		case ":empty":
			return items.length === 0;
		case ":count":
			return items.length;
		case ":one":
			return items.length === 1;
		case ":indexOf": {
			if (values.length < 2) {
				return -1;
			}
			const list = toSelection(values[0]);
			const target = toSelection(values[1]);
			return target.length > 0 ? list.indexOf(target[0]) : -1;
		}
		case ":len":
			return typeof input === "string" ? input.length : items.length;
		case ":join":
			return items.map((item) => textOf(item)).join(String(args[0]));
		default:
			return null;
	}
};
const applyCall = (op, values) => {
	const input = values.length > 0 ? values[values.length - 1] : null;
	const args = op.args ?? [];

	switch (op.family) {
		case "cardinality":
			return cardinality(op.name, args, input);
		case "selection":
			return select(op.name, args, input);
		case "traversal":
			return traverse(op.name, args, input);
		case "filter":
			return filter(op.name, args, input);
		case "map":
			return map(op.name, args, input);
		case "reducer":
			return reduce(op.name, args, values, input);
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
	stack.push(applyCall(op, values));
}

if (stack.length === 0) {
	return [];
}

const result = stack[stack.length - 1];

%s
}`
