package templates

import (
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/html/drivers/cdp/eval"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const getChildren = "(el) => Array.from(el.children)"

func GetChildren(id cdpruntime.RemoteObjectID) *eval.Function {
	return eval.F(getChildren).WithArgRef(id)
}

const getChildrenCount = "(el) => el.children.length"

func GetChildrenCount(id cdpruntime.RemoteObjectID) *eval.Function {
	return eval.F(getChildrenCount).WithArgRef(id)
}

const getChildByIndex = "(el, idx) => el.children[idx]"

func GetChildByIndex(id cdpruntime.RemoteObjectID, index runtime.Int) *eval.Function {
	return eval.F(getChildByIndex).WithArgRef(id).WithArgValue(index)
}
