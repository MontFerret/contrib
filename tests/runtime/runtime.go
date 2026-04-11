package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/MontFerret/contrib/modules/csv"
	"github.com/MontFerret/contrib/modules/toml"
	"github.com/MontFerret/contrib/modules/web/article"
	"github.com/MontFerret/contrib/modules/web/html"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp"
	"github.com/MontFerret/contrib/modules/web/html/drivers/memory"
	"github.com/MontFerret/contrib/modules/web/robots"
	"github.com/MontFerret/contrib/modules/web/sitemap"
	"github.com/MontFerret/contrib/modules/xml"
	"github.com/MontFerret/contrib/modules/yaml"
	"github.com/MontFerret/ferret/v2"
	"github.com/MontFerret/ferret/v2/pkg/asm"
	"github.com/MontFerret/ferret/v2/pkg/compiler"
	"github.com/MontFerret/ferret/v2/pkg/diagnostics"
	"github.com/MontFerret/ferret/v2/pkg/logging"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/source"

	"github.com/rs/zerolog"
)

type Params []string

func (p *Params) String() string {
	return "[" + strings.Join(*p, ",") + "]"
}

func (p *Params) Set(value string) error {
	*p = append(*p, value)
	return nil
}

func (p *Params) ToMap() (runtime.Params, error) {
	res := runtime.NewParams()

	for _, entry := range *p {
		pair := strings.SplitN(entry, ":", 2)

		if len(pair) < 2 {
			return nil, runtime.Error(runtime.ErrInvalidArgument, entry)
		}

		var value interface{}
		key := pair[0]

		err := json.Unmarshal([]byte(pair[1]), &value)

		if err != nil {
			fmt.Println(pair[1])
			return nil, err
		}

		if err := res.Set(key, value); err != nil {
			fmt.Println(pair[1])
			return nil, err
		}
	}

	return res, nil
}

var (
	conn = flag.String(
		"browser-address",
		"http://127.0.0.1:9222",
		"set CDP address",
	)

	dryRun = flag.Bool(
		"dry-run",
		false,
		"compiles a given query, but does not execute",
	)

	optimizationLevel = flag.Int(
		"ol",
		int(compiler.O1),
		"set optimization level (0-3)",
	)

	logLevel = flag.String(
		"log-level",
		logging.ErrorLevel.String(),
		"log level",
	)
)

var logger zerolog.Logger

func main() {
	var params Params

	flag.Var(
		&params,
		"param",
		`query parameter (--param=foo:\"bar\", --param=id:1)`,
	)

	flag.Parse()

	console := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "15:04:05.999",
	}
	logger = zerolog.New(console).
		Level(zerolog.Level(logging.MustParseLogLevel(*logLevel))).
		With().
		Timestamp().
		Logger()

	stat, _ := os.Stdin.Stat()

	var query string
	var files []string

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// check whether the app is getting a query via standard input
		std := bufio.NewReader(os.Stdin)

		b, err := io.ReadAll(std)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		query = string(b)
	} else if flag.NArg() > 0 {
		files = flag.Args()
	} else {
		fmt.Println(flag.NArg())
		fmt.Println("File or input stream are required")
		os.Exit(1)
	}

	var err error

	p, err := params.ToMap()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			<-c
			cancel()
		}
	}()

	drivers, err := html.New(
		html.WithDefaultDriver(memory.New()),
		html.WithDrivers(cdp.New(cdp.WithAddress(*conn))),
	)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	engine, e := ferret.New(
		ferret.WithLog(console),
		ferret.WithLogLevel(ferret.MustParseLogLevel(*logLevel)),
		ferret.WithModules(
			csv.New(),
			toml.New(),
			xml.New(),
			yaml.New(),
			article.New(),
			robots.New(),
			sitemap.New(),
			drivers,
		),
	)

	if e != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sessionOptions := []ferret.SessionOption{
		ferret.WithSessionRuntimeParams(p),
	}

	if query != "" {
		err = runQuery(ctx, engine, sessionOptions, source.New("stdin", query))
	} else {
		err = execFiles(ctx, engine, sessionOptions, files)
	}

	if err != nil {
		fmt.Println(ferret.FormatError(err))
		os.Exit(1)
	}
}

func execFiles(ctx context.Context, engine *ferret.Engine, opts []ferret.SessionOption, files []string) error {
	return processFiles(ctx, files, "execute", func(ctx context.Context, src *source.Source) error {
		return runQuery(ctx, engine, opts, src)
	})
}

func runQuery(ctx context.Context, engine *ferret.Engine, opts []ferret.SessionOption, query *source.Source) error {
	if !(*dryRun) {
		return execQuery(ctx, engine, opts, query)
	}

	return analyzeQuery(query)
}

func execQuery(ctx context.Context, engine *ferret.Engine, opts []ferret.SessionOption, query *source.Source) error {
	plan, err := engine.Compile(ctx, query)

	if err != nil {
		return err
	}

	sess, err := plan.NewSession(ctx, opts...)

	if err != nil {
		return err
	}

	defer sess.Close()

	res, err := sess.Run(ctx)

	if err == nil {
		printResult(ctx, res)
	}

	if err != nil {
		frmt, ok := err.(diagnostics.Formattable)

		if ok {
			fmt.Println(frmt.Format())
		} else {
			fmt.Println(err)
		}

		os.Exit(1)
	}

	return nil
}

func processFiles(ctx context.Context, files []string, op string, predicate func(ctx context.Context, src *source.Source) error) error {
	errList := make([]diagnostics.FormattableError, 0, len(files))

	for _, path := range files {
		log := logger.With().Str("path", path).Logger()
		log.Debug().Msg("checking path...")

		info, err := os.Stat(path)

		if err != nil {
			log.Debug().Err(err).Msg("failed to get path info")

			errList = append(errList, &diagnostics.Diagnostic{
				Kind:    diagnostics.UnexpectedError,
				Message: "failed to get path info",
				Source:  source.New("stdin", path),
				Cause:   err,
			})

			continue
		}

		if info.IsDir() {
			log.Debug().Msg("path points to a directory. retrieving list of files...")

			fileInfos, err := os.ReadDir(path)

			if err != nil {
				log.Debug().Err(err).Msg("failed to retrieve list of files")

				errList = append(errList, &diagnostics.Diagnostic{
					Kind:    diagnostics.UnexpectedError,
					Message: "failed to retrieve list of files",
					Source:  source.New("stdin", path),
					Cause:   err,
				})

				continue
			}

			log.Debug().Int("size", len(fileInfos)).Msg("retrieved list of files. starting to iterate...")

			dirFiles := make([]string, 0, len(fileInfos))

			for _, info := range fileInfos {
				if filepath.Ext(info.Name()) == ".fql" {
					dirFiles = append(dirFiles, filepath.Join(path, info.Name()))
				}
			}

			if len(dirFiles) > 0 {
				if err := processFiles(ctx, dirFiles, op, predicate); err != nil {
					log.Debug().Err(err).Msg(fmt.Sprintf("failed to %s files", op))

					errList = append(errList, &diagnostics.Diagnostic{
						Kind:    diagnostics.UnexpectedError,
						Message: fmt.Sprintf("failed to %s files", op),
						Source:  source.New("stdin", path),
						Cause:   err,
					})
				} else {
					log.Debug().Int("size", len(fileInfos)).Err(err).Msg(fmt.Sprintf("successfully %sed files", op))
				}
			} else {
				log.Debug().Int("size", len(fileInfos)).Err(err).Msg("no FQL files found")
			}

			continue
		}

		log.Debug().Msg("path points to a file. starting to read content")

		src, err := source.Read(path)

		if err != nil {
			log.Debug().Err(err).Msg("failed to read content")

			errList = append(errList, &diagnostics.Diagnostic{
				Kind:    diagnostics.UnexpectedError,
				Message: "failed to read content",
				Source:  source.New("stdin", path),
				Cause:   err,
			})

			continue
		}

		log.Debug().Msg("successfully read file")
		log.Debug().Msg(fmt.Sprintf("starting to %s file...", op))
		err = predicate(ctx, src)

		if err != nil {
			log.Debug().Err(err).Msg("failed to execute file")

			derr, ok := err.(diagnostics.FormattableError)

			if ok {
				errList = append(errList, derr)
			} else {
				errList = append(errList, &diagnostics.Diagnostic{
					Kind:    diagnostics.UnexpectedError,
					Message: "failed to execute file",
					Source:  src,
					Cause:   err,
				})
			}

			log.Debug().Err(derr).Msg("failed to execute file with diagnostics")

			continue
		}

		log.Debug().Msg("successfully executed file")
	}

	if len(errList) > 0 {
		if len(errList) == len(files) {
			logger.Debug().Interface("errors", errList).Msg("failed to execute file(s)")
		} else {
			logger.Debug().Interface("errors", errList).Msg("executed with errors")
		}

		return diagnostics.NewDiagnosticsOf(errList)
	}

	return nil
}

func printResult(_ context.Context, res *ferret.Output) {
	_, _ = os.Stdout.Write(res.Content)
	_, _ = os.Stdout.WriteString("\n")
}

func analyzeQuery(query *source.Source) error {
	optLevel := compiler.OptimizationLevel(*optimizationLevel)

	if optLevel < 0 || optLevel > 3 {
		fmt.Printf("Invalid optimization level: %d.", optLevel)
		os.Exit(1)
	}

	c := compiler.New(compiler.WithOptimizationLevel(optLevel))
	prog, err := c.Compile(query)

	if err != nil {
		fmt.Println(diagnostics.Format(err))
		os.Exit(1)
	}

	dis, err := asm.Disassemble(prog)

	if err != nil {
		fmt.Println("Failed to disassemble program:", err)
		os.Exit(1)
	}

	fmt.Println(dis)

	return nil
}
