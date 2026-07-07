package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/document/pdf/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// PageInfo returns normalized page dimensions and numbering.
func PageInfo(ctx context.Context, pageValue runtime.Value) (runtime.Value, error) {
	page, err := requirePage(pageValue, "PAGE_INFO")
	if err != nil {
		return runtime.None, err
	}

	info, err := page.Info(ctx)
	if err != nil {
		return runtime.None, err
	}

	return pageInfoValue(info), nil
}

// Blocks returns low-level positioned text fragments from a PDF page.
func Blocks(ctx context.Context, pageValue runtime.Value) (runtime.Value, error) {
	page, err := requirePage(pageValue, "BLOCKS")
	if err != nil {
		return runtime.None, err
	}

	blocks, err := page.Blocks(ctx)
	if err != nil {
		return runtime.None, err
	}

	out := runtime.NewArray(len(blocks))
	for _, block := range blocks {
		if err := out.Append(ctx, blockValue(block)); err != nil {
			return runtime.None, err
		}
	}

	return out, nil
}

func pageInfoValue(info core.PageInfo) runtime.Value {
	return runtime.NewObjectWith(map[string]runtime.Value{
		"number":   runtime.NewInt(info.Number),
		"width":    runtime.NewFloat(info.Width),
		"height":   runtime.NewFloat(info.Height),
		"rotation": runtime.NewInt(info.Rotation),
	})
}

func blockValue(block core.TextBlock) runtime.Value {
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
