package templates

import (
	"fmt"

	"github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
)

const (
	isElementInViewportFragment = `function isInViewport(i) {
	var bounding = i.getBoundingClientRect();
	
	return (
		bounding.top >= 0 &&
		bounding.left >= 0 &&
		bounding.bottom <= (window.innerHeight || document.documentElement.clientHeight) &&
		bounding.right <= (window.innerWidth || document.documentElement.clientWidth)
	);
}`

	scrollToPositionFragment = `function scrollToPosition(left, top, opts) {
	const scrollingElement = document.scrollingElement || document.documentElement || document.body;
	if (scrollingElement == null) {
		return false;
	}

	const maxLeft = Math.max(0, scrollingElement.scrollWidth - scrollingElement.clientWidth);
	const maxTop = Math.max(0, scrollingElement.scrollHeight - scrollingElement.clientHeight);
	const targetLeft = Math.min(Math.max(left, 0), maxLeft);
	const targetTop = Math.min(Math.max(top, 0), maxTop);

	if (scrollingElement.scrollLeft === targetLeft && scrollingElement.scrollTop === targetTop) {
		return false;
	}

	window.scrollTo({
		left: targetLeft,
		top: targetTop,
		behavior: opts.behavior,
		block: opts.block, 
		inline: opts.inline
	});

	return true;
}`

	scrollBy = `(opts) => {
	window.scrollBy({
		left: opts.left,
		top: opts.top,
		behavior: opts.behavior
	});
}`
)

var (
	scroll = fmt.Sprintf(`(opts) => {
	%s

	return scrollToPosition(opts.left, opts.top, opts);
}`, scrollToPositionFragment)

	scrollTop = fmt.Sprintf(`(opts) => {
	%s

	return scrollToPosition(0, 0, opts);
}`, scrollToPositionFragment)

	scrollBottom = fmt.Sprintf(`(opts) => {
	%s

	return scrollToPosition(0, Number.MAX_SAFE_INTEGER, opts);
}`, scrollToPositionFragment)

	scrollIntoView = fmt.Sprintf(`(el, opts) => {
	%s

	if (!isInViewport(el)) {
		el.scrollIntoView({
			behavior: opts.behavior,
			block: opts.block, 
			inline: opts.inline
		});
	}

	return true;
}`, isElementInViewportFragment)

	scrollIntoViewByCSSSelector = fmt.Sprintf(`(el, selector, opts) => {
		const found = el.querySelector(selector);

		%s

		%s

		if (isInViewport(found)) {
			return false;
		}

		found.scrollIntoView({
			behavior: opts.behavior,
			block: opts.block,
			inline: opts.inline
		});

		return true;
}`, notFoundErrorFragment, isElementInViewportFragment)

	scrollIntoViewByXPathSelector = fmt.Sprintf(`(el, selector, opts) => {
		%s

		%s

		%s

		if (isInViewport(found)) {
			return false;
		}

		found.scrollIntoView({
			behavior: opts.behavior,
			block: opts.block,
			inline: opts.inline
		});

		return true;
}`, xpathAsElementFragment, notFoundErrorFragment, isElementInViewportFragment)
)

func Scroll(options drivers.ScrollOptions) *eval.Function {
	return eval.F(scroll).WithArg(options)
}

func ScrollBy(options drivers.ScrollOptions) *eval.Function {
	return eval.F(scrollBy).WithArg(options)
}

func ScrollTop(options drivers.ScrollOptions) *eval.Function {
	return eval.F(scrollTop).WithArg(options)
}

func ScrollBottom(options drivers.ScrollOptions) *eval.Function {
	return eval.F(scrollBottom).WithArg(options)
}

func ScrollIntoView(id runtime.RemoteObjectID, options drivers.ScrollOptions) *eval.Function {
	return eval.F(scrollIntoView).WithArgRef(id).WithArg(options)
}

func ScrollIntoViewBySelector(id runtime.RemoteObjectID, selector drivers.QuerySelector, options drivers.ScrollOptions) *eval.Function {
	return toFunction(selector, scrollIntoViewByCSSSelector, scrollIntoViewByXPathSelector).
		WithArgRef(id).
		WithArgSelector(selector).
		WithArg(options)
}
