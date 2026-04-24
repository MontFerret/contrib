package cssx

import (
	"fmt"

	"github.com/MontFerret/cssx"
)

func ValidateCallArgs(exp Expression, step cssx.Op) error {
	switch exp {
	case ExpressionFirst,
		ExpressionLast,
		ExpressionParent,
		ExpressionExists,
		ExpressionEmpty,
		ExpressionCount,
		ExpressionLen,
		ExpressionText,
		ExpressionTexts,
		ExpressionOwnText,
		ExpressionNormalize,
		ExpressionTrim,
		ExpressionHTML,
		ExpressionOuterHTML,
		ExpressionValue,
		ExpressionAbsURL,
		ExpressionParseURL,
		ExpressionDedupeByText,
		ExpressionToNumber:
		if err := validateLiteralCount(step, 0); err != nil {
			return err
		}

		return validateArityRange(step, 0, 1)
	case ExpressionNth,
		ExpressionTake,
		ExpressionSkip:
		if err := validateLiteralCount(step, 1); err != nil {
			return err
		}

		if err := validateLiteralKind(step, 0, cssx.CallArgNumber); err != nil {
			return err
		}

		return validateArityRange(step, 0, 1)
	case ExpressionSlice:
		if err := validateLiteralCount(step, 2); err != nil {
			return err
		}

		if err := validateLiteralKind(step, 0, cssx.CallArgNumber); err != nil {
			return err
		}

		if err := validateLiteralKind(step, 1, cssx.CallArgNumber); err != nil {
			return err
		}

		return validateArityRange(step, 0, 1)
	case ExpressionWithin,
		ExpressionClosest,
		ExpressionChildren,
		ExpressionNext,
		ExpressionPrev,
		ExpressionHas,
		ExpressionMatches,
		ExpressionIndexOf,
		ExpressionFilter:
		if err := validateLiteralCount(step, 0); err != nil {
			return err
		}

		return validateArityRange(step, 1, 2)
	case ExpressionJoin:
		if err := validateLiteralCount(step, 1); err != nil {
			return err
		}

		if err := validateLiteralKind(step, 0, cssx.CallArgString); err != nil {
			return err
		}

		return validateArityRange(step, 0, 1)
	case ExpressionAttr,
		ExpressionAttrs,
		ExpressionProp,
		ExpressionURL,
		ExpressionWithAttr,
		ExpressionWithText,
		ExpressionDedupeByAttr:
		if err := validateLiteralCount(step, 1); err != nil {
			return err
		}

		if err := validateLiteralKind(step, 0, cssx.CallArgString); err != nil {
			return err
		}

		return validateArityRange(step, 0, 1)
	case ExpressionReplace:
		if err := validateLiteralCount(step, 2); err != nil {
			return err
		}

		if err := validateLiteralKind(step, 0, cssx.CallArgString); err != nil {
			return err
		}

		if err := validateLiteralKind(step, 1, cssx.CallArgString); err != nil {
			return err
		}

		return validateArityRange(step, 0, 1)
	case ExpressionRegex:
		if err := validateLiteralCountRange(step, 1, 2); err != nil {
			return err
		}

		if err := validateLiteralKind(step, 0, cssx.CallArgString); err != nil {
			return err
		}

		if len(step.Args) > 1 {
			if err := validateLiteralKind(step, 1, cssx.CallArgNumber); err != nil {
				return err
			}
		}

		return validateArityRange(step, 0, 1)
	case ExpressionToDate:
		if err := validateLiteralCountRange(step, 0, 1); err != nil {
			return err
		}

		if len(step.Args) > 0 {
			if err := validateLiteralKind(step, 0, cssx.CallArgString); err != nil {
				return err
			}
		}

		return validateArityRange(step, 0, 1)
	default:
		return fmt.Errorf("unsupported expression %q", exp)
	}
}

func CollectCallArgs(step cssx.Op) ([]any, error) {
	args := make([]any, 0, len(step.Args))

	for argIndex, arg := range step.Args {
		switch arg.Kind {
		case cssx.CallArgString:
			args = append(args, arg.Str)
		case cssx.CallArgNumber:
			args = append(args, arg.Num)
		default:
			return nil, fmt.Errorf("unsupported literal argument kind %d at %d", arg.Kind, argIndex)
		}
	}

	return args, nil
}

func validateLiteralCount(step cssx.Op, expected int) error {
	if len(step.Args) != expected {
		return fmt.Errorf("expected %d literal args, got %d", expected, len(step.Args))
	}

	return nil
}

func validateLiteralCountRange(step cssx.Op, min, max int) error {
	if len(step.Args) < min || len(step.Args) > max {
		return fmt.Errorf("expected %d..%d literal args, got %d", min, max, len(step.Args))
	}

	return nil
}

func validateLiteralKind(step cssx.Op, index int, expected cssx.CallArgKind) error {
	if index < 0 || index >= len(step.Args) {
		return fmt.Errorf("literal arg index %d is out of bounds", index)
	}

	if step.Args[index].Kind != expected {
		return fmt.Errorf("expected literal arg %d to be %s, got %s", index, literalKindName(expected), literalKindName(step.Args[index].Kind))
	}

	return nil
}

func validateArityRange(step cssx.Op, min, max int) error {
	if step.Arity < min || step.Arity > max {
		return fmt.Errorf("expected arity %d..%d, got %d", min, max, step.Arity)
	}

	return nil
}

func literalKindName(kind cssx.CallArgKind) string {
	switch kind {
	case cssx.CallArgString:
		return "string"
	case cssx.CallArgNumber:
		return "number"
	default:
		return fmt.Sprintf("kind(%d)", kind)
	}
}
