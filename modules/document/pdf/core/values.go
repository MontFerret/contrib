package core

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func pageInfoProperty(info PageInfo, name string) runtime.Value {
	switch name {
	case "width":
		return runtime.NewFloat(info.Width)
	case "height":
		return runtime.NewFloat(info.Height)
	case "rotation":
		return runtime.NewInt(info.Rotation)
	default:
		return runtime.None
	}
}

func textBlocksValue(ctx context.Context, blocks []TextBlock) (runtime.Value, error) {
	out := runtime.NewArray(len(blocks))
	for _, block := range blocks {
		if err := out.Append(ctx, textBlockValue(block)); err != nil {
			return runtime.None, err
		}
	}

	return out, nil
}

func textBlockValue(block TextBlock) runtime.Value {
	return runtime.NewObjectWith(map[string]runtime.Value{
		"text": runtime.NewString(block.Text),
		"bounds": runtime.NewObjectWith(map[string]runtime.Value{
			"x":      runtime.NewFloat(block.Bounds.X),
			"y":      runtime.NewFloat(block.Bounds.Y),
			"width":  runtime.NewFloat(block.Bounds.Width),
			"height": runtime.NewFloat(block.Bounds.Height),
		}),
	})
}
