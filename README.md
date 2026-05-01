# MontFerret Contrib

`contrib` is a workspace for independently versioned [MontFerret](https://github.com/MontFerret/ferret) modules. This README acts as the top-level index for the modules currently available in this repository.

## Available Modules

| Module        | Description | README |
|---------------| --- | --- |
| `csv`         | CSV module and ``CSV`` namespace helpers for Ferret. | [modules/csv/README.md](./modules/csv/README.md) |
| `toml`        | TOML module and ``TOML`` namespace helpers for Ferret. | [modules/toml/README.md](./modules/toml/README.md) |
| `web/article` | Article extraction helpers under `WEB::ARTICLE` for Ferret. | [modules/web/article/README.md](./modules/web/article/README.md) |
| `web/html`    | HTML module for Ferret. | [modules/web/html/README.md](./modules/web/html/README.md) |
| `web/robots`  | robots.txt parsing and policy helpers under `WEB::ROBOTS` for Ferret. | [modules/web/robots/README.md](./modules/web/robots/README.md) |
| `web/sitemap` | Sitemap discovery helpers under `WEB::SITEMAP` for Ferret. | [modules/web/sitemap/README.md](./modules/web/sitemap/README.md) |
| `xml`         | XML module and ``XML`` namespace helpers for Ferret. | [modules/xml/README.md](./modules/xml/README.md) |
| `yaml`        | YAML module and ``YAML`` namespace helpers for Ferret. | [modules/yaml/README.md](./modules/yaml/README.md) |

Module-specific documentation lives in each module README and will be expanded there later.

## Development

Use the repo-level `Makefile` to run module commands:

```sh
make test [module ...]
make lint [module ...]
make fmt [module ...]
```

If no module names are provided, the commands run against all available modules.

## Release Procedure

Each module is released by creating and pushing a module-specific git tag in the format:

`modules/<module>/v<version>`

1. Validate the target module:

```sh
make test <module>
make lint <module>
```

2. Create a tag using the `Makefile` release targets:

```sh
# Auto-bump from latest module tag
make release-patch <module>
make release-minor <module>
make release-major <module>

# Explicit semantic version (including pre-release)
make release-pre 1.0.0-rc.1 <module>
```

3. Push the created tag:

```sh
git push origin modules/<module>/v<version>
```
