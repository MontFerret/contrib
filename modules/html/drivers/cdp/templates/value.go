package templates

import (
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/html/drivers/cdp/eval"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const getValue = `(el) => {
	return el.value
}`

func GetValue(id cdpruntime.RemoteObjectID) *eval.Function {
	return eval.F(getValue).WithArgRef(id)
}

const setValue = `(el, value) => {
	el.value = value
}`

func SetValue(id cdpruntime.RemoteObjectID, value runtime.Value) *eval.Function {
	return eval.F(setValue).WithArgRef(id).WithArgValue(value)
}
