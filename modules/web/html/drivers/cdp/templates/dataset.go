package templates

import (
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const getDataset = `(el) => {
	return Object.keys(el.dataset).reduce((out, key) => {
		out[key] = el.dataset[key];
		return out;
	}, {});
}`

func GetDataset(id cdpruntime.RemoteObjectID) *eval.Function {
	return eval.F(getDataset).WithArgRef(id)
}

const setDatasetProperty = `(el, name, value) => {
	el.dataset[name] = value;
}`

func SetDatasetProperty(id cdpruntime.RemoteObjectID, name, value runtime.String) *eval.Function {
	return eval.F(setDatasetProperty).WithArgRef(id).WithArgValue(name).WithArgValue(value)
}

const removeDatasetProperty = `(el, name) => {
	delete el.dataset[name];
}`

func RemoveDatasetProperty(id cdpruntime.RemoteObjectID, name runtime.String) *eval.Function {
	return eval.F(removeDatasetProperty).WithArgRef(id).WithArgValue(name)
}
