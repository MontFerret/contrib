package templates

import (
	"github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/html/drivers/cdp/eval"
)

const getParent = "(el) => el.parentElement"

func GetParent(id runtime.RemoteObjectID) *eval.Function {
	return eval.F(getParent).WithArgRef(id)
}
