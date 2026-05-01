package templates

import (
	"encoding/json"
	"fmt"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/contrib/modules/web/html/drivers/internal/cssx"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func CSSX(id cdpruntime.RemoteObjectID, expression runtime.String) (*eval.Function, error) {
	ops, err := cssx.CompileOps(string(expression))
	if err != nil {
		return nil, err
	}

	opsRaw, err := json.Marshal(ops)

	if err != nil {
		return nil, err
	}

	exp := fmt.Sprintf(cssxStateMachine, string(opsRaw))

	return eval.F(exp).WithArgRef(id), nil
}
