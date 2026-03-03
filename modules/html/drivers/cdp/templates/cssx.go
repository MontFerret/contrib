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
	Arity    int    `json:"arity,omitempty"`
	Args     []any  `json:"args,omitempty"`
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
			exp, err := resolveCallExpression(step.Name)

			if err != nil {
				return nil, fmt.Errorf("invalid call operation at %d: %w", idx, err)
			}

			if err := validateCallArgs(exp, step); err != nil {
				return nil, fmt.Errorf("invalid %s call at %d: %w", exp, idx, err)
			}

			args := make([]any, 0, len(step.Args))

			for argIndex, arg := range step.Args {
				switch arg.Kind {
				case cssxc.CallArgString:
					args = append(args, arg.Str)
				case cssxc.CallArgNumber:
					args = append(args, arg.Num)
				default:
					return nil, fmt.Errorf("unsupported literal argument kind %d at call %d arg %d", arg.Kind, idx, argIndex)
				}
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

func resolveCallExpression(name string) (cssx.Expression, error) {
	resolved := strings.TrimSpace(name)

	if resolved == "" {
		return "", fmt.Errorf("call name is empty")
	}

	if !strings.HasPrefix(resolved, ":") {
		resolved = ":" + resolved
	}

	return cssx.ResolveSelector(resolved)
}

func validateCallArgs(exp cssx.Expression, step cssxc.Op) error {
	switch exp {
	case cssx.ExpressionFirst,
		cssx.ExpressionLast,
		cssx.ExpressionParent,
		cssx.ExpressionExists,
		cssx.ExpressionEmpty,
		cssx.ExpressionCount,
		cssx.ExpressionLen,
		cssx.ExpressionText,
		cssx.ExpressionTexts,
		cssx.ExpressionOwnText,
		cssx.ExpressionNormalize,
		cssx.ExpressionTrim,
		cssx.ExpressionHTML,
		cssx.ExpressionOuterHTML,
		cssx.ExpressionValue,
		cssx.ExpressionAbsURL,
		cssx.ExpressionParseURL,
		cssx.ExpressionDedupeByText,
		cssx.ExpressionToNumber:
		if err := validateLiteralCount(step, 0); err != nil {
			return err
		}

		return validateArityRange(step, 0, 1)
	case cssx.ExpressionNth,
		cssx.ExpressionTake,
		cssx.ExpressionSkip:
		if err := validateLiteralCount(step, 1); err != nil {
			return err
		}

		if err := validateLiteralKind(step, 0, cssxc.CallArgNumber); err != nil {
			return err
		}

		return validateArityRange(step, 0, 1)
	case cssx.ExpressionSlice:
		if err := validateLiteralCount(step, 2); err != nil {
			return err
		}

		if err := validateLiteralKind(step, 0, cssxc.CallArgNumber); err != nil {
			return err
		}

		if err := validateLiteralKind(step, 1, cssxc.CallArgNumber); err != nil {
			return err
		}

		return validateArityRange(step, 0, 1)
	case cssx.ExpressionWithin,
		cssx.ExpressionClosest,
		cssx.ExpressionChildren,
		cssx.ExpressionNext,
		cssx.ExpressionPrev,
		cssx.ExpressionHas,
		cssx.ExpressionMatches,
		cssx.ExpressionIndexOf,
		cssx.ExpressionFilter:
		if err := validateLiteralCount(step, 0); err != nil {
			return err
		}

		return validateArityRange(step, 1, 2)
	case cssx.ExpressionJoin:
		if err := validateLiteralCount(step, 1); err != nil {
			return err
		}

		if err := validateLiteralKind(step, 0, cssxc.CallArgString); err != nil {
			return err
		}

		return validateArityRange(step, 0, 1)
	case cssx.ExpressionAttr,
		cssx.ExpressionAttrs,
		cssx.ExpressionProp,
		cssx.ExpressionURL,
		cssx.ExpressionWithAttr,
		cssx.ExpressionWithText,
		cssx.ExpressionDedupeByAttr:
		if err := validateLiteralCount(step, 1); err != nil {
			return err
		}

		if err := validateLiteralKind(step, 0, cssxc.CallArgString); err != nil {
			return err
		}

		return validateArityRange(step, 0, 1)
	case cssx.ExpressionReplace:
		if err := validateLiteralCount(step, 2); err != nil {
			return err
		}

		if err := validateLiteralKind(step, 0, cssxc.CallArgString); err != nil {
			return err
		}

		if err := validateLiteralKind(step, 1, cssxc.CallArgString); err != nil {
			return err
		}

		return validateArityRange(step, 0, 1)
	case cssx.ExpressionRegex:
		if err := validateLiteralCountRange(step, 1, 2); err != nil {
			return err
		}

		if err := validateLiteralKind(step, 0, cssxc.CallArgString); err != nil {
			return err
		}

		if len(step.Args) > 1 {
			if err := validateLiteralKind(step, 1, cssxc.CallArgNumber); err != nil {
				return err
			}
		}

		return validateArityRange(step, 0, 1)
	case cssx.ExpressionToDate:
		if err := validateLiteralCountRange(step, 0, 1); err != nil {
			return err
		}

		if len(step.Args) > 0 {
			if err := validateLiteralKind(step, 0, cssxc.CallArgString); err != nil {
				return err
			}
		}

		return validateArityRange(step, 0, 1)
	default:
		return fmt.Errorf("unsupported expression %q", exp)
	}
}

func validateLiteralCount(step cssxc.Op, expected int) error {
	if len(step.Args) != expected {
		return fmt.Errorf("expected %d literal args, got %d", expected, len(step.Args))
	}

	return nil
}

func validateLiteralCountRange(step cssxc.Op, min, max int) error {
	if len(step.Args) < min || len(step.Args) > max {
		return fmt.Errorf("expected %d..%d literal args, got %d", min, max, len(step.Args))
	}

	return nil
}

func validateLiteralKind(step cssxc.Op, index int, expected cssxc.CallArgKind) error {
	if index < 0 || index >= len(step.Args) {
		return fmt.Errorf("literal arg index %d is out of bounds", index)
	}

	if step.Args[index].Kind != expected {
		return fmt.Errorf("expected literal arg %d to be %s, got %s", index, literalKindName(expected), literalKindName(step.Args[index].Kind))
	}

	return nil
}

func validateArityRange(step cssxc.Op, min, max int) error {
	if step.Arity < min || step.Arity > max {
		return fmt.Errorf("expected arity %d..%d, got %d", min, max, step.Arity)
	}

	return nil
}

func literalKindName(kind cssxc.CallArgKind) string {
	switch kind {
	case cssxc.CallArgString:
		return "string"
	case cssxc.CallArgNumber:
		return "number"
	default:
		return fmt.Sprintf("kind(%d)", kind)
	}
}
