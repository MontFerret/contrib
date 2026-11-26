package main

import (
	"fmt"
	"io"
	"runtime/debug"
	"strings"

	"github.com/rs/zerolog"

	"github.com/MontFerret/api"
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
	ferretnet "github.com/MontFerret/ferret/v2/pkg/net"
	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
	"github.com/MontFerret/ferret/v2/uapi"
)

func ferretVersion() (api.Version, error) {
	const modulePath = "github.com/MontFerret/ferret/v2"

	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, dep := range info.Deps {
			if dep.Path != modulePath {
				continue
			}

			if dep.Replace != nil {
				dep = dep.Replace
			}

			if dep.Version != "" && dep.Version != "(devel)" {
				return api.Version(strings.TrimPrefix(dep.Version, "v")), nil
			}
		}
	}

	return "", fmt.Errorf("build metadata has no version for %s", modulePath)
}

func newRuntime(browserAddress string, logLevel ferret.LogLevel, stderr io.Writer) (*uapi.Runtime, error) {
	version, err := ferretVersion()
	if err != nil {
		return nil, err
	}

	return uapi.New(version,
		ferret.WithLog(zerolog.ConsoleWriter{Out: stderr, TimeFormat: "15:04:05.999"}),
		ferret.WithLogLevel(logLevel),
		ferret.WithNetworkOptions(ferretnet.WithHTTPPolicies(ferrethttp.WithAllowLocalhost(true))),
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
				html.WithDrivers(cdp.New(cdp.WithAddress(browserAddress))),
			),
		),
	)
}
