package templates

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/MontFerret/contrib/modules/html/drivers/cdp/eval"
	"github.com/MontFerret/contrib/modules/html/drivers/common/cssx"
	cssxc "github.com/MontFerret/cssx"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"
)

type cssxCompiledOp struct {
	Kind     string `json:"kind"`
	Selector string `json:"selector,omitempty"`
	Name     string `json:"name,omitempty"`
	Args     []any  `json:"args,omitempty"`
	Arity    int    `json:"arity,omitempty"`
	Index    int    `json:"index"`
}

func CSSX(id cdpruntime.RemoteObjectID, expression runtime.String) (*eval.Function, error) {
	pipeline, err := cssx.Compile(string(expression))

	if err != nil {
		return nil, err
	}

	ops, err := compileCSSXOps(pipeline)

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

func compileCSSXOps(pipeline cssxc.Pipeline) ([]cssxCompiledOp, error) {
	ops := make([]cssxCompiledOp, 0, len(pipeline.Ops))

	for idx, step := range pipeline.Ops {
		switch step.Kind {
		case cssxc.OpSelect:
			selector := strings.TrimSpace(step.Selector)

			if selector == "" {
				return nil, fmt.Errorf("invalid select operation at %d: selector is empty", idx)
			}

			ops = append(ops, cssxCompiledOp{
				Kind:     "select",
				Selector: selector,
				Index:    idx,
			})
		case cssxc.OpCall:
			exp, err := cssx.ResolveSelector(step.Name)

			if err != nil {
				return nil, fmt.Errorf("invalid call operation at %d: %w", idx, err)
			}

			if err := cssx.ValidateCallArgs(exp, step); err != nil {
				return nil, fmt.Errorf("invalid %s call at %d: %w", exp, idx, err)
			}

			args, err := cssx.CollectCallArgs(step)

			if err != nil {
				return nil, fmt.Errorf("collect call args at %d: %w", idx, err)
			}

			ops = append(ops, cssxCompiledOp{
				Kind:  "call",
				Name:  string(exp),
				Arity: step.Arity,
				Args:  args,
				Index: idx,
			})
		default:
			return nil, fmt.Errorf("unexpected pipeline operation at %d: %d", idx, step.Kind)
		}
	}

	return ops, nil
}
