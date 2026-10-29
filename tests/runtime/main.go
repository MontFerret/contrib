package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"

	llmmodule "github.com/MontFerret/contrib/modules/ai/llm"
	archivemodule "github.com/MontFerret/contrib/modules/archive"
	"github.com/MontFerret/contrib/modules/csv"
	dbpostgres "github.com/MontFerret/contrib/modules/db/postgres"
	dbredis "github.com/MontFerret/contrib/modules/db/redis"
	dbsqlite "github.com/MontFerret/contrib/modules/db/sqlite"
	documentpdf "github.com/MontFerret/contrib/modules/document/pdf"
	documentxlsx "github.com/MontFerret/contrib/modules/document/xlsx"
	restmodule "github.com/MontFerret/contrib/modules/net/rest"
	jwtmodule "github.com/MontFerret/contrib/modules/security/jwt"
	oauth2module "github.com/MontFerret/contrib/modules/security/oauth2"
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
	"github.com/MontFerret/ferret/v2/pkg/diagnostics"
	"github.com/MontFerret/ferret/v2/pkg/logging"
	ferretnet "github.com/MontFerret/ferret/v2/pkg/net"
	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
	"github.com/MontFerret/ferret/v2/pkg/source"

	"github.com/rs/zerolog"
)

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
		int(ferret.OptimizationFull),
		"set optimization level (none, basic, full)",
	)

	logLevel = flag.String(
		"log-level",
		ferret.LogError.String(),
		"log level",
	)
)

var logger zerolog.Logger

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, ferret.FormatError(err))
		os.Exit(1)
	}
}

func run() (resultErr error) {
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

	stat, err := os.Stdin.Stat()
	if err != nil {
		return err
	}

	var query string
	var files []string

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// check whether the app is getting a query via standard input
		std := bufio.NewReader(os.Stdin)

		b, err := io.ReadAll(std)
		if err != nil {
			return err
		}

		query = string(b)
	} else if flag.NArg() > 0 {
		files = flag.Args()
	} else {
		return fmt.Errorf("file or input stream is required")
	}

	p, err := params.ToMap()
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	network, err := ferretnet.New(
		ferretnet.WithHTTPPolicies(ferrethttp.WithAllowLocalhost(true)),
	)
	if err != nil {
		return err
	}

	engine, err := ferret.New(
		ferret.WithLog(console),
		ferret.WithLogLevel(ferret.MustParseLogLevel(*logLevel)),
		ferret.WithNetwork(network),
		ferret.WithModules(
			llmmodule.New(),
			archivemodule.New(),
			csv.New(),
			dbpostgres.New(),
			dbredis.New(),
			dbsqlite.New(),
			documentpdf.New(),
			documentxlsx.New(),
			restmodule.New(),
			jwtmodule.New(),
			oauth2module.New(),
			toml.New(),
			xml.New(),
			yaml.New(),
			article.New(),
			robots.New(),
			sitemap.New(),
			html.New(
				html.WithDefaultDriver(memory.New()),
				html.WithDrivers(cdp.New(cdp.WithAddress(*conn))),
			),
		),
	)

	if err != nil {
		return err
	}

	defer func() { resultErr = joinErrors(resultErr, engine.Close()) }()

	sessionOptions := []ferret.SessionOption{
		ferret.WithSessionRuntimeParams(p),
	}

	if query != "" {
		err = runQuery(ctx, engine, sessionOptions, ferret.NewSource("stdin", query))
	} else {
		err = execFiles(ctx, engine, sessionOptions, files)
	}

	return err
}

func processFiles(ctx context.Context, files []string, op string, predicate func(ctx context.Context, src ferret.Source) error) error {
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
