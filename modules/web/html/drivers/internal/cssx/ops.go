package cssx

import (
	"fmt"
	"strings"

	cssxpipeline "github.com/MontFerret/cssx"
)

type OpKind string

const (
	OpSelect OpKind = "select"
	OpCall   OpKind = "call"
)

type CompiledOp struct {
	Kind     OpKind `json:"kind"`
	Selector string `json:"selector,omitempty"`
	Name     string `json:"name,omitempty"`
	Args     []any  `json:"args,omitempty"`
	Arity    int    `json:"arity,omitempty"`
	Index    int    `json:"index"`
}

func CompileOps(input string) ([]CompiledOp, error) {
	pipeline, err := Compile(input)
	if err != nil {
		return nil, err
	}

	return CompilePipeline(pipeline)
}

func CompilePipeline(pipeline cssxpipeline.Pipeline) ([]CompiledOp, error) {
	ops := make([]CompiledOp, 0, len(pipeline.Ops))

	for idx, step := range pipeline.Ops {
		switch step.Kind {
		case cssxpipeline.OpSelect:
			selector := strings.TrimSpace(step.Selector)
			if selector == "" {
				return nil, fmt.Errorf("invalid select operation at %d: selector is empty", idx)
			}

			ops = append(ops, CompiledOp{
				Kind:     OpSelect,
				Selector: selector,
				Index:    idx,
			})
		case cssxpipeline.OpCall:
			exp, err := ResolveSelector(step.Name)
			if err != nil {
				return nil, fmt.Errorf("invalid call operation at %d: %w", idx, err)
			}

			if err := ValidateCallArgs(exp, step); err != nil {
				return nil, fmt.Errorf("invalid %s call at %d: %w", exp, idx, err)
			}

			args, err := CollectCallArgs(step)
			if err != nil {
				return nil, fmt.Errorf("collect call args at %d: %w", idx, err)
			}

			ops = append(ops, CompiledOp{
				Kind:  OpCall,
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
