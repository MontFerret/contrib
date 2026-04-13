# AGENTS.md

This file is the canonical operating guide for coding agents working in this repository.

If you see documentation that conflicts with this file, prefer `Makefile`, `go.work`, the per-module `go.mod` files, and the workflows under `.github/workflows`.

## Repo snapshot

- Repository: `github.com/MontFerret/contrib`
- This repository is a Go workspace for independently versioned Ferret modules.
- Workspace root responsibilities:
    - `modules/` contains the actual module implementations.
    - `scripts/` contains repo automation helpers.
    - `Makefile` is the preferred entrypoint for common development tasks.
    - `go.work` wires the module workspace together for local development.
- This is not the core Ferret v2 repository. Do not assume that core compiler, parser, VM, or runtime code lives here.
- The main architectural role of this repository is extension, not core language/runtime ownership.

## Architectural mental model

`contrib` is a collection of optional Ferret modules.

Primary mental model:

`Ferret core runtime` -> `contrib module registration` -> `module-specific functions/types/contracts` -> `consumer usage from FQL or embedding code`

Subsystem responsibilities:

- `modules/*` own the actual optional functionality exposed to Ferret users.
- Each module is independently versioned and has its own `go.mod`.
- The repository root coordinates workspace-level tooling and release/development flows.
- `scripts/modules.sh` is the source of truth for module enumeration in repo-level commands.
- Module-specific README files are the canonical place for module-level documentation.

Agents should reason about changes by ownership boundary:

- If the task changes optional functionality, start in the owning module under `modules/...`.
- If the task changes workspace-wide development behavior, start at the repo root (`Makefile`, `go.work`, `scripts/`).
- If the task would require changing Ferret core semantics, APIs, compiler behavior, VM behavior, or runtime internals, that change likely belongs in `github.com/MontFerret/ferret`, not here.

## Canonical invariants

- This repository contains optional modules, not the Ferret core runtime.
- Modules should integrate with Ferret as extensions rather than reimplementing core behavior.
- Each module should remain independently understandable and independently releasable.
- Repo-level tooling must continue to work both for all modules and for targeted module subsets.
- Module discovery for repo automation is based on finding `go.mod` files under `modules/`.
- Prefer local module ownership over cross-module coupling.
- Do not introduce hidden dependencies between modules unless explicitly required and well justified.
- Avoid turning the workspace root into a shared dumping ground for module internals.

## Repository map

Agents should begin with the package or directory whose responsibility owns the requested behavior.

### Workspace-level surfaces

- `Makefile`
    - Preferred entrypoint for common repo tasks such as `test`, `lint`, and `fmt`.
    - Delegates module-aware behavior to `scripts/modules.sh`.

- `go.work`
    - Defines the local workspace and wires included modules together for development.
    - Update this when adding or removing modules from the workspace.

- `scripts/modules.sh`
    - Central module-discovery and module-targeting script used by repo-level commands.
    - Owns behaviors such as:
        - listing modules
        - validating requested module names
        - applying `build`, `test`, `lint`, and `fmt` per module

- `revive.toml`
    - Repo-level lint configuration used by the lint workflow.

- `.github/workflows`
    - CI and automation definitions for the repository.

### Module surfaces

- `modules/csv`
    - CSV module and `CSV` namespace helpers.

- `modules/toml`
    - TOML module and `TOML` namespace helpers.

- `modules/xml`
    - XML module and `XML` namespace helpers.

- `modules/yaml`
    - YAML module and `YAML` namespace helpers.

- `modules/web/article`
    - Article extraction helpers under `WEB::ARTICLE`.

- `modules/web/html`
    - HTML-related Ferret module functionality.

- `modules/web/robots`
    - robots.txt parsing and policy helpers under `WEB::ROBOTS`.

- `modules/web/sitemap`
    - Sitemap discovery helpers under `WEB::SITEMAP`.

Each module should be treated as its own ownership boundary first.

## Primary surfaces

- Repo root
    - workspace orchestration
    - lint/format/test/build entrypoints
    - module discovery
    - CI alignment

- Per-module directory
    - module code
    - module tests
    - module-level public API
    - module README
    - module-specific dependencies in `go.mod`

## Where to start by task

- Add or change functionality in a module:
    - inspect the owning module under `modules/...`
    - inspect its `go.mod`
    - inspect its README
    - update tests in that module
    - validate only that module first
    - broaden validation only if the change affects shared tooling or cross-module behavior

- Add a new module:
    - create the new directory under `modules/...`
    - add a `go.mod` for the module
    - add module documentation
    - update `go.work`
    - ensure repo-level scripts can discover and operate on it
    - validate through `make test <module>`, `make lint <module>`, and `make fmt <module>`

- Change repo-wide build/test/lint/fmt behavior:
    - inspect `Makefile`
    - inspect `scripts/modules.sh`
    - inspect `revive.toml` if lint behavior changes
    - verify both all-modules and targeted-module flows

- Change CI or automation:
    - inspect `.github/workflows`
    - verify consistency with `Makefile` and `scripts/modules.sh`
    - avoid duplicating module selection logic in multiple places when the script already owns it

- Change documentation:
    - repo-wide documentation belongs in the root `README.md`
    - module-specific documentation belongs in that module’s `README.md`
    - keep module README examples aligned with actual exported behavior

- Change Ferret integration behavior:
    - inspect the owning module first
    - verify whether the change truly belongs in `contrib`
    - if the change depends on new core runtime/compiler/VM behavior, note that it likely requires changes in the main Ferret repository instead

## Stability guide

Treat these as relatively stable unless the task explicitly targets them:

- the repository’s role as an extension workspace
- module ownership under `modules/`
- repo-level command entry through `Makefile`
- module discovery through `scripts/modules.sh`
- independent versioning and per-module `go.mod` ownership

Treat these as implementation-sensitive and verify current code before proposing changes:

- module-internal exported surfaces
- any shared conventions across modules
- lint/format details
- workspace wiring in `go.work`
- release automation and CI behavior

Do not treat historical discussions, stale TODOs, or assumptions from the core Ferret repository as authoritative for this repository.

## Module design guidance

These rules are mandatory unless the task explicitly requires otherwise.

- Keep module ownership local.
- Prefer implementing functionality inside the smallest responsible module.
- Do not add repo-wide abstractions for behavior that is only used by one module.
- Share code across modules only when:
    - the duplication is substantial,
    - the abstraction is clearly stable,
    - and the shared location has a clean ownership story.
- Avoid introducing cross-module imports unless there is a strong, explicit reason.
- A module should remain understandable without needing to inspect unrelated modules.
- Public behavior exposed by a module should be documented in that module’s README.

## Go type and file structure rules

These rules are mandatory unless the task explicitly requires otherwise.

- Do not define multiple method-bearing structs in the same `.go` file.
- Prefer declaring a method-bearing struct as a standalone `type Name struct { ... }`.
- A method-bearing struct should usually live in its own file, named after the primary type or responsibility whenever practical.
- Grouped `type ( ... )` declarations are allowed for interfaces, passive data-only structs, and other small related helper/value types that belong to the same narrow concern.
- A grouped `type ( ... )` block may also contain exactly one method-bearing struct when:
    - it is the only behavioral type in the file, and
    - the other grouped types are passive helper/value types from the same narrow concern.
- Do not use grouped `type ( ... )` declarations to hide multiple substantial behavioral types.
- If a helper struct later gains methods and would create more than one method-bearing struct in the file, extract it into its own file immediately.
- Methods for a struct should live in the same file as the struct unless there is a strong, explicit reason to split by concern.
- Do not place a new method-bearing struct into an existing file just because the code compiles.

### Function and method ownership rules

These rules are mandatory unless the task explicitly requires otherwise.

- A file centered on a method-bearing type should contain the type, its methods, and its constructors only.
- Do not mix package-level helper functions into a file that already contains methods for a primary type.
- In type-centered files, constructor functions are the only normally allowed package-level functions.
- If logic conceptually belongs to the primary type, implement it as a method.
- If logic does not belong to the type and must remain a package-level function, place it in a separate helper-focused file.
- Package-level functions are preferred only when there is no natural owning type or when the behavior is genuinely package-level.
- If a file contains both methods and non-constructor package-level functions, that is usually a structure violation and should be refactored.

## Comment rules for functions and methods

- Do not add comments to every function or method by default.
- Exported functions and methods should usually have doc comments, especially for module public APIs.
- Unexported functions and methods should be commented only when they carry non-obvious behavior, invariants, side effects, ownership rules, cleanup expectations, or protocol/lifecycle constraints.
- Comments must explain intent, contract, invariants, side effects, or lifecycle behavior.
- Prefer comments that explain why the code exists, what must remain true, or how the method is meant to be used.
- Do not write comments that merely restate the method name or signature.
- Avoid comment wallpaper. Dense, meaningful comments are preferred over mechanically documenting obvious code.

## Development practice expectations

Agents must follow repository-specific engineering discipline rather than generic style preferences.

### Core principles

- Preserve correctness first.
- Preserve module ownership boundaries.
- Prefer the smallest local change that fully solves the task.
- Avoid introducing abstractions, indirection, or refactors unless they are necessary for correctness, maintainability, or an explicitly requested design change.
- Keep behavioral ownership obvious in code structure, naming, and file layout.
- Prefer changing one module over changing many modules unless the task truly spans them.

### Mandatory expectations

- Identify the owning module or workspace surface before making a non-trivial change.
- Preserve existing behavior unless the task explicitly requires changing it.
- Add or update tests for any behavior change.
- Run the narrowest relevant validation first, then broaden as appropriate.
- Do not claim tests or validation were completed unless they were actually run.
- Do not perform opportunistic refactors unrelated to the requested task unless they are required for correctness.
- Do not assume that a pattern from one module automatically belongs in another; verify first.

## Required workflow for non-trivial changes

Before making a non-trivial change, agents must:

1. Identify the owning module or repo-level surface.
2. Identify the contract, invariant, or behavior being preserved or changed.
3. Choose the smallest reasonable implementation that fits the existing design.
4. Determine whether the change affects only one module or shared workspace behavior.
5. Add or update correctness tests.
6. Run relevant validation and summarize the results accurately.

## Validation and evidence

When finishing a non-trivial change, agents should report:

- owning module or owning repo-level surface
- files changed
- tests added or updated
- validation commands run
- notable invariants preserved or intentionally changed

### Change discipline

- Prefer adapting an existing local pattern over introducing a new architectural pattern.
- Do not add new helper layers, wrappers, or abstractions only for aesthetic reasons.
- Do not move code across modules unless the ownership boundary is genuinely wrong.
- Keep diffs focused on the requested task.
- If cleanup is necessary to make the requested change safe, keep it tightly scoped and explain why it was needed.

### Decision bias when uncertain

When uncertain:

- preserve existing behavior
- prefer the smaller local change
- add a focused test
- prefer module-local ownership
- verify whether the change belongs in `contrib` or in the core Ferret repository before expanding scope

## Tooling prerequisites

- Go must be installed.
- `make` is the preferred entrypoint for repo-defined workflows.
- The repo-level workflows depend on:
    - `staticcheck`
    - `fieldalignment`
    - `goimports`
    - `revive`
- Use the repo-level `Makefile` when possible instead of invoking ad-hoc commands directly.

## Validation commands

Prefer these forms first:

```sh
make test [module ...]
make lint [module ...]
make fmt [module ...]
```

Where `[module ...]` is an optional list of module names to target. If no modules are specified, the command runs for all modules.

Useful module discovery:

```sh
make modules
```

If no module names are provided, repo-level commands operate on all discovered modules.