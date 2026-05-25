package templates

import (
	"encoding/json"
	"fmt"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/contrib/modules/web/html/drivers/internal/cssx"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type cssxOutputMode int

const (
	cssxOutputList cssxOutputMode = iota
	cssxOutputOne
	cssxOutputCount
	cssxOutputExists
)

func CSSX(id cdpruntime.RemoteObjectID, expression runtime.String) (*eval.Function, error) {
	return cssxQuery(id, expression, cssxOutputList)
}

func CSSXOne(id cdpruntime.RemoteObjectID, expression runtime.String) (*eval.Function, error) {
	return cssxQuery(id, expression, cssxOutputOne)
}

func CSSXCount(id cdpruntime.RemoteObjectID, expression runtime.String) (*eval.Function, error) {
	return cssxQuery(id, expression, cssxOutputCount)
}

func CSSXExists(id cdpruntime.RemoteObjectID, expression runtime.String) (*eval.Function, error) {
	return cssxQuery(id, expression, cssxOutputExists)
}

func cssxQuery(id cdpruntime.RemoteObjectID, expression runtime.String, mode cssxOutputMode) (*eval.Function, error) {
	ops, err := cssx.CompileOps(string(expression))
	if err != nil {
		return nil, err
	}

	if fn := cssxSelectorFastPath(id, ops, mode); fn != nil {
		return fn, nil
	}

	opsRaw, err := json.Marshal(ops)

	if err != nil {
		return nil, err
	}

	exp := fmt.Sprintf(cssxStateMachine, string(opsRaw), cssxFinalizer(mode))

	return eval.F(exp).WithArgRef(id), nil
}

func cssxSelectorFastPath(
	id cdpruntime.RemoteObjectID,
	ops []cssx.CompiledOp,
	mode cssxOutputMode,
) *eval.Function {
	if len(ops) != 1 || ops[0].Kind != cssx.OpSelect {
		return nil
	}

	selector := ops[0].Selector

	switch mode {
	case cssxOutputOne:
		return eval.F(cssxSelectorOne).WithArgRef(id).WithArg(selector)
	case cssxOutputCount:
		return eval.F(cssxSelectorCount).WithArgRef(id).WithArg(selector)
	case cssxOutputExists:
		return eval.F(cssxSelectorExists).WithArgRef(id).WithArg(selector)
	default:
		return nil
	}
}

func cssxFinalizer(mode cssxOutputMode) string {
	switch mode {
	case cssxOutputOne:
		return cssxFinalizerOne
	case cssxOutputCount:
		return cssxFinalizerCount
	case cssxOutputExists:
		return cssxFinalizerExists
	default:
		return cssxFinalizerList
	}
}

const cssxSelectorOne = `(el, selector) => {
	try {
		if (el == null || typeof el.querySelector !== "function") {
			return null;
		}

		return el.querySelector(selector);
	} catch (_) {
		return null;
	}
}`

const cssxSelectorCount = `(el, selector) => {
	try {
		if (el == null || typeof el.querySelectorAll !== "function") {
			return 0;
		}

		return el.querySelectorAll(selector).length;
	} catch (_) {
		return 0;
	}
}`

const cssxSelectorExists = `(el, selector) => {
	try {
		if (el == null || typeof el.querySelector !== "function") {
			return false;
		}

		return el.querySelector(selector) != null;
	} catch (_) {
		return false;
	}
}`

const cssxFinalizerList = `if (Array.isArray(result)) {
	return result;
}

if (result == null) {
	return [];
}

return [result];`

const cssxFinalizerOne = `if (Array.isArray(result)) {
	return result.length > 0 ? result[0] : null;
}

if (result == null) {
	return null;
}

return result;`

const cssxFinalizerCount = `if (Array.isArray(result)) {
	return result.length;
}

if (result == null) {
	return 0;
}

return 1;`

const cssxFinalizerExists = `if (Array.isArray(result)) {
	return result.length > 0;
}

return result != null;`
