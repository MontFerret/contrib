package templates

import (
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const getClassList = `(el) => {
	return Array.from(el.classList).reduce((out, name) => {
		out[name] = true;
		return out;
	}, {});
}`

func GetClassList(id cdpruntime.RemoteObjectID) *eval.Function {
	return eval.F(getClassList).WithArgRef(id)
}

const setClass = `(el, name, enabled) => {
	if (enabled) {
		el.classList.add(name);
	} else {
		el.classList.remove(name);
	}
}`

func SetClass(id cdpruntime.RemoteObjectID, name runtime.String, enabled runtime.Boolean) *eval.Function {
	return eval.F(setClass).WithArgRef(id).WithArgValue(name).WithArgValue(enabled)
}
