# MontFerret Contrib

`contrib` is a workspace for independently versioned [Ferret](https://github.com/MontFerret/ferret) modules. This README acts as the top-level index for the modules currently available in this repository.

## Available Modules

| Module         | Description                                                                               | README                                                             |
|----------------|-------------------------------------------------------------------------------------------|--------------------------------------------------------------------|
| `ai/llm`       | Text generation, chat, extraction, and classification under `AI::LLM` for Ferret.        | [modules/ai/llm/README.md](./modules/ai/llm/README.md)             |
| `archive`      | Sandboxed ZIP, TAR, TAR.GZ, and TGZ inspection and extraction under `ARCHIVE`.            | [modules/archive/README.md](./modules/archive/README.md)           |
| `csv`          | CSV module and ``CSV`` namespace helpers for Ferret.                                      | [modules/csv/README.md](./modules/csv/README.md)                   |
| `db/postgres`  | Postgres database handles under `DB::POSTGRES` for Ferret.                                | [modules/db/postgres/README.md](./modules/db/postgres/README.md)   |
| `db/redis`     | Redis connection handles and generic commands under `DB::REDIS` for Ferret.               | [modules/db/redis/README.md](./modules/db/redis/README.md)         |
| `db/sqlite`    | SQLite database handles under `DB::SQLITE` for Ferret.                                    | [modules/db/sqlite/README.md](./modules/db/sqlite/README.md)       |
| `document/pdf` | Read-only PDF document handles under `DOCUMENT::PDF` for Ferret.                          | [modules/document/pdf/README.md](./modules/document/pdf/README.md) |
| `document/xlsx` | Excel-compatible `.xlsx` workbook handles under `DOCUMENT::XLSX` for Ferret.              | [modules/document/xlsx/README.md](./modules/document/xlsx/README.md) |
| `net/rest`     | REST-style HTTP API clients under `NET::REST` for Ferret.                                 | [modules/net/rest/README.md](./modules/net/rest/README.md)         |
| `security/jwt` | JWT token signing, verification, and inspection helpers under `SECURITY::JWT` for Ferret. | [modules/security/jwt/README.md](./modules/security/jwt/README.md) |
| `security/oauth2` | OAuth 2.0 client grants and token helpers under `SECURITY::OAUTH2` for Ferret.           | [modules/security/oauth2/README.md](./modules/security/oauth2/README.md) |
| `toml`         | TOML module and ``TOML`` namespace helpers for Ferret.                                    | [modules/toml/README.md](./modules/toml/README.md)                 |
| `web/article`  | Article extraction helpers under `WEB::ARTICLE` for Ferret.                               | [modules/web/article/README.md](./modules/web/article/README.md)   |
| `web/html`     | HTML module for Ferret.                                                                   | [modules/web/html/README.md](./modules/web/html/README.md)         |
| `web/robots`   | robots.txt parsing and policy helpers under `WEB::ROBOTS` for Ferret.                     | [modules/web/robots/README.md](./modules/web/robots/README.md)     |
| `web/sitemap`  | Sitemap discovery helpers under `WEB::SITEMAP` for Ferret.                                | [modules/web/sitemap/README.md](./modules/web/sitemap/README.md)   |
| `xml`          | XML module and ``XML`` namespace helpers for Ferret.                                      | [modules/xml/README.md](./modules/xml/README.md)                   |
| `yaml`         | YAML module and ``YAML`` namespace helpers for Ferret.                                    | [modules/yaml/README.md](./modules/yaml/README.md)                 |

Module-specific documentation lives in each module README and will be expanded there later.

## Module Manifests

Each module directory contains a `ferret.yaml` document conforming to
[Ferret Module Manifest v1](https://schemas.ferretlang.org/module/v1.json).
Install the pinned manifest validator, then validate the complete manifest
inventory and its repository conventions with:

```sh
make install-manifest-validator
make validate-manifests
```

The module release commands update manifest versions as part of the atomic
release transaction described below.

## Development

The modules require Ferret `v2.0.0-alpha.58` or later within the v2 series.

Use the repo-level `Makefile` to run module commands:

```sh
make test [module ...]
make lint [module ...]
make fmt [module ...]
```

If no module names are provided, the commands run against all available modules.

Run `make test-runtime` to validate the binary runtime harness used by Ferret Lab.

Support packages under `pkg/` can be updated across module `go.mod` and `go.sum` files with:

```sh
make update-package <package> <version> [module ...]
```

If no module names are provided, only modules that already require the package are updated.

The Ferret runtime dependency can be updated across all workspace `go.mod` files that require it with:

```sh
make update-ferret <version>
```

## Release Procedure

Each module is released by creating and pushing a module-specific git tag in the format:

`modules/<module>/v<version>`

1. Validate the target module:

```sh
make test <module>
make lint <module>
```

2. From a clean `main` branch synchronized with `origin/main`, run a release
target:

```sh
# Auto-bump from latest module tag
make release-patch <module>
make release-minor <module>
make release-major <module>

# Explicit semantic version (including pre-release)
make release-pre xml 1.0.0-rc.1

# Auto-bump latest matching pre-release
make release-pre xml rc

# Auto-bump latest matching pre-release for all modules.
# Interactive runs prompt once for a base version when modules have no initial tag.
make release-pre-all rc

# Non-interactive bulk prereleases can set the initial base version explicitly.
RELEASE_PRE_BASE_VERSION=1.0.0 make release-pre-all rc
```

A single-module release updates and validates its `ferret.yaml`, creates one
release commit when the version changed, creates the annotated module tag at
that commit, and atomically pushes `main` and the tag. `release-pre-all` updates
all module manifests in one release commit and atomically pushes that commit
with every module tag, so a failed push cannot leave a partial bulk release.
Releases with more than three tags always create a release commit, even when
the manifests already contain the requested versions, to trigger publication
through the update to `main`.
Set `RELEASE_CHECK_ONLY=1` to validate a planned release without creating a
commit, tag, or remote change.

The Barn publication workflow handles individual `modules/**/v*` tag pushes
and pushes to `main`. For a main push, it discovers module release tags pointing
to the commits introduced by that push. This also handles bulk releases, for
which GitHub suppresses tag push events when more than three tags are pushed
together. Ordinary pushes without release tags do not publish anything.

Each selected tag runs independently using its exact source commit. The
workflow uses Barn's public Go API to validate the tagged module, prepares the
immutable registry source records, and opens one ready-for-review pull request
per module version in [MontFerret/barn](https://github.com/MontFerret/barn).
The release appears in the public registry only after that pull request passes
Barn CI, is reviewed, and is merged. Re-running the workflow updates the existing publication branch
or succeeds without changes when the version is already published.
Duplicate triggers are serialized per tag, and a failed module does not cancel
the remaining publications.

To backfill or retry releases, run the workflow from `main` with either a module
tag or the full release commit SHA. A commit selects every module release tag
pointing to it; invalid references and commits without release tags are rejected.
The workflow uses its current publisher tooling with the historical module
source, without changing tags or versions:

```sh
gh workflow run publish-module.yml --repo MontFerret/contrib --ref main \
  -f release_ref=modules/db/redis/v1.0.0-rc.5

# Backfill the complete release batch from this commit.
gh workflow run publish-module.yml --repo MontFerret/contrib --ref main \
  -f release_ref=13f03eedcf202ffd91459aa975a49347d908006a
```

Check the workflow's per-module results and Barn PRs after dispatch. Discovery
tests run as part of `make test-publisher` and require Git, Bash, and `jq`.

Support-package tags under `pkg/` do not trigger Barn publication.

Support packages under `pkg/` are released by creating and pushing a package-specific git tag in the format:

`pkg/<package>/v<version>`

1. Validate the target package:

```sh
make test-packages <package>
make lint-packages <package>
```

2. Create and push an annotated tag using the package release targets:

```sh
# Auto-bump from latest package tag
make release-package-patch <package>
make release-package-minor <package>
make release-package-major <package>

# Explicit semantic version (including pre-release)
make release-package-pre common 0.1.0-rc.1

# Auto-bump latest matching pre-release
make release-package-pre common rc
```

3. Update module requirements and sums that use the released package:

```sh
make update-package <package> 0.1.0
```
