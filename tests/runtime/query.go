package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/MontFerret/ferret/v2"
	"github.com/MontFerret/ferret/v2/pkg/asm"
	"github.com/MontFerret/ferret/v2/pkg/compiler"
)

func execFiles(ctx context.Context, engine *ferret.Engine, opts []ferret.SessionOption, files []string) error {
	return processFiles(ctx, files, "execute", func(ctx context.Context, src ferret.Source) error {
		return runQuery(ctx, engine, opts, src)
	})
}

func runQuery(ctx context.Context, engine *ferret.Engine, opts []ferret.SessionOption, query ferret.Source) error {
	if !*dryRun {
		return execQuery(ctx, engine, opts, query, os.Stdout)
	}

	return analyzeQuery(ctx, os.Stdout, query, compiler.OptimizationLevel(*optimizationLevel))
}

func execQuery(ctx context.Context, engine *ferret.Engine, opts []ferret.SessionOption, query ferret.Source, dst io.Writer) error {
	output, err := engine.Run(ctx, query, opts...)
	if err != nil {
		return err
	}

	return writeResult(ctx, dst, output)
}

func analyzeQuery(ctx context.Context, dst io.Writer, query ferret.Source, level compiler.OptimizationLevel) error {
	if level < 0 || level > 3 {
		return fmt.Errorf("invalid optimization level: %d", level)
	}

	c, err := compiler.New(compiler.WithOptimizationLevel(level))
	if err != nil {
		return err
	}

	prog, err := c.Compile(ctx, query)
	if err != nil {
		return err
	}

	dis, err := asm.Disassemble(prog)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(dst, dis)

	return err
}
