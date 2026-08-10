# AGENTS.md

This file is the canonical operating guide for coding agents working in this repository.

If repository documentation conflicts with this file, prefer `Makefile`, `go.work`, the per-module and per-package `go.mod` files, and the workflows under `.github/workflows` for commands, toolchains, and CI behavior.

## Repo snapshot

- Repository: `github.com/MontFerret/contrib`
- This repository is a Go workspace for independently versioned Ferret modules and support packages.
- Workspace root responsibilities:
    - `modules/` contains the optional Ferret module implementations.
    - `pkg/` contains independently versioned support packages shared only when there is a stable ownership case.
    - `scripts/` contains module/package orchestration, dependency update, publication, and release tooling.
    - `tests/` contains the runtime harness, manifest checks, shared fixtures, and FQL integration suites.
    - `Makefile` is the preferred entrypoint for common development and validation tasks.
    - `go.work` wires modules, support packages, and the runtime harness together for local development.
- Go versions are owned by `go.work` and each independently versioned `go.mod`; do not assume every workspace member uses the same version.
- This is not the Ferret core repository. Do not assume that compiler, parser, bytecode, VM, or core runtime implementation belongs here.
- The main architectural role of this repository is extension, not core language/runtime ownership.

## Architectural mental model

`contrib` is a collection of optional Ferret modules plus narrowly owned support packages and repository tooling.

Primary flow:

```text
Ferret host/runtime
        ↓
top-level contrib module registration
        ↓
thin Ferret-facing lib bindings
        ↓
module-owned core behavior
        ↓
external service, format, protocol, or resource
```

Agents should reason about changes by ownership boundary:

- Optional user-facing behavior belongs in the owning module under `modules/`.
- Module construction, options, and registration belong in the module's top-level package.
- Ferret argument decoding, runtime value conversion, and function registration belong in `lib/`.
- Reusable module implementation, resource handling, protocols, validation, parsing, and client logic belong in `core/`.
- Cross-module helpers belong in `pkg/` only when the duplication is substantial, the abstraction is stable, and the ownership story is clear.
- Workspace-wide build, test, lint, format, dependency, publication, and release behavior belongs at the repository root or under `scripts/`.
- FQL integration behavior is exercised through the runtime harness and fixtures under `tests/`.
- Changes to Ferret language semantics or core runtime/compiler behavior usually belong in `github.com/MontFerret/ferret`, not here.

## Canonical invariants

- This repository contains optional modules, not the Ferret core runtime or standard library.
- Modules integrate with Ferret through the current module and SDK contracts rather than reimplementing core behavior.
- Each module and support package remains independently understandable and independently releasable.
- Public FQL namespaces, function names, arguments, results, errors, and resource behavior are user-visible contracts.
- Module registration should use Ferret's extension-facing SDK helpers when they fit the existing contract.
- Runtime access to user-controlled filesystem paths must go through Ferret's filesystem API resolved from the execution context.
- Resources and lazy values must preserve deterministic cleanup on success, failure, and cancellation paths.
- Repo-level module and package tooling must continue to work for both the full inventory and explicitly targeted subsets.
- Module discovery is based on `go.mod` files under `modules/`; support-package discovery is based on `go.mod` files under `pkg/`.
- Prefer module-local ownership over cross-module coupling.
- Do not introduce hidden dependencies between modules unless explicitly required and justified.
- Avoid turning the workspace root or `pkg/` into a dumping ground for module internals.
- Preserve existing behavior unless the task explicitly requires changing it.

## Repository map

Agents should begin with the package or directory whose responsibility owns the requested behavior.

### Workspace-level surfaces

- `Makefile`
    - Preferred entrypoint for common repository tasks.
    - Delegates module-aware behavior to `scripts/modules.sh`.
    - Delegates support-package behavior to `scripts/packages.sh`.
- `go.work`
    - Defines the local workspace and its development-time replacements.
    - Update it when adding or removing workspace members.
- `scripts/modules.sh`
    - Source of truth for module discovery and targeting.
    - Owns module build, unit test, FQL integration, lint, format, version, and dependency traversal behavior.
    - Knows module-specific integration prerequisites such as Chromium for `web/html`.
- `scripts/packages.sh`
    - Source of truth for support-package discovery and targeting.
    - Owns package build, unit test, lint, and format behavior.
- `scripts/publish/`
    - Isolated Go module for Barn publication preparation used by the publication workflow.
- Release and update scripts under `scripts/`
    - Own dependency updates, manifest version updates, release transactions, and package/module tag workflows.
    - Do not duplicate their discovery, versioning, or transaction logic elsewhere.
- `revive.toml`
    - Repository lint configuration consumed by module and package lint flows.
- `.github/workflows/`
    - Owns the current CI, service-backed integration, publication, and dependency-update behavior.

### Module surfaces

Modules live under `modules/`. A module identifier may be one segment, such as `archive`, or nested, such as `db/sqlite`.

Each module owns:

- its `go.mod` and `go.sum`;
- its `ferret.yaml` publication manifest;
- its public constructor and options;
- its FQL namespace and function registrations;
- its implementation, tests, and user-facing README;
- its external protocol, format, or service behavior.

Do not rely on this file for a complete module list. Use `make modules` or `scripts/modules.sh` because discovery is based on `go.mod` files under `modules/`.

### Support-package surfaces

Support packages live under `pkg/` and are independently versioned.

- `pkg/common` owns narrowly shared helpers used across modules.
- Support-package APIs are cross-module contracts and must not become a shortcut for module-specific internals.
- Use `make packages` for the current package inventory.
- Update module requirements through the repository's package update flow rather than by inconsistently editing consumers.

### Test and tooling surfaces

- `tests/runtime/`
    - Builds the Ferret runtime harness used by FQL integration tests.
    - Explicitly imports and registers modules that must be available to integration fixtures.
- `tests/modules/`
    - Contains FQL-level integration suites organized by module identifier.
- `tests/data/`
    - Contains shared static, dynamic, API, and other integration fixtures.
- `tests/manifests/`
    - Contains repository-level manifest convention tests and runs outside the main workspace.
- `scripts/publish/`
    - Contains publication preparation tests and is validated as an isolated Go module.
- `scripts/release-test.sh`
    - Exercises release transaction behavior.

## Where to start by task

- Add or change module behavior:
    - inspect the owning module, its `go.mod`, `ferret.yaml`, README, and existing tests;
    - locate the top-level, `lib`, and `core` ownership seams;
    - test the smallest responsible layer first;
    - add FQL integration coverage when the behavior crosses the Ferret boundary.
- Add a new module:
    - create the module under `modules/<module>` with its own `go.mod` and `ferret.yaml`;
    - add top-level construction, `core`, `lib`, package documentation, README, and tests as applicable;
    - add the module to `go.work`;
    - add it to the root README index;
    - import and register it in `tests/runtime/main.go`;
    - add fixtures under `tests/modules/<module>` when FQL integration coverage is appropriate;
    - rely on `scripts/modules.sh` discovery rather than adding a second module registry;
    - validate its manifest and targeted build, test, lint, format, and integration flows.
- Add or change a support package:
    - inspect the owning package under `pkg/` and all current consumers;
    - keep the public surface narrow and stable;
    - validate the package first, then affected modules;
    - use the package update and release tooling for versioned consumer changes.
- Change repo-wide build, test, lint, or format behavior:
    - inspect `Makefile`, the owning script, and relevant CI workflows;
    - verify both full-inventory and targeted flows;
    - keep module and package discovery centralized.
- Change dependency update behavior:
    - inspect the relevant update script and workspace/module manifests;
    - preserve independent module versions and workspace consistency;
    - verify affected `go.mod`, `go.sum`, and workspace vendor state through the owning flow.
- Change publication or release behavior:
    - inspect the owning script, its tests, module manifests, and publication workflow;
    - preserve atomicity, tag conventions, and dry-run/check-only behavior;
    - do not perform a real release or publication unless explicitly requested.
- Change CI or integration behavior:
    - inspect `.github/workflows`, `Makefile`, and the owning discovery script;
    - keep local commands aligned with CI;
    - account for Chromium, Lab, Postgres, Redis, or other declared service prerequisites.
- Change documentation:
    - put repository-wide guidance in the root README or this file as appropriate;
    - put module-specific public documentation in the module README;
    - keep examples aligned with actual exported and FQL-visible behavior.
- Change Ferret integration behavior:
    - inspect the owning module first;
    - prefer current `pkg/sdk`, `pkg/module`, `pkg/runtime`, and `pkg/fs` contracts;
    - if the change needs new compiler, VM, language, or core runtime behavior, identify the upstream Ferret change instead of implementing it here.

## Stability guide

Treat these as relatively stable unless the task explicitly targets them:

- the repository's role as an extension workspace;
- module ownership under `modules/`;
- support-package ownership under `pkg/`;
- top-level to `lib` to `core` dependency direction;
- repo-level command entry through `Makefile`;
- discovery through `scripts/modules.sh` and `scripts/packages.sh`;
- independent versioning and per-member `go.mod` ownership;
- Ferret filesystem mediation for user-controlled runtime paths.

Treat these as implementation-sensitive and verify current code before proposing changes:

- module-internal exported surfaces;
- FQL namespace and function contracts;
- Ferret SDK and runtime integration details;
- resource ownership, streaming, cancellation, and cleanup behavior;
- shared-package APIs and their consumers;
- lint/format details;
- workspace wiring and dependency replacements;
- manifests, publication, release automation, and CI behavior.

Do not treat historical discussions, stale TODOs, old branches, or assumptions from the Ferret core repository as authoritative over current code and repository guidance.

## Module code organization

Ferret contrib modules should follow a consistent internal layout so that implementation code, Ferret runtime bindings, and public module wiring remain easy to understand and maintain.

Use `modules/db/sqlite` as the general structural reference for new modules, while adapting the layout to the module's actual needs.

### Recommended structure

```text
modules/<module>/
  core/
    ...
  lib/
    ...
  README.md
  doc.go
  options.go
  module.go
  module_test.go
  ferret.yaml
  go.mod
  go.sum
```

File names may follow an established module-local convention, such as `<module>.go` instead of `module.go`. Do not rename existing files only to match the example.

### `core/`

The `core` package contains the module's main implementation.

Put business logic, resource management, protocol/client code, validation, parsing, query handling, and reusable internal types here. Code in `core` should not depend on Ferret function binding details unless a Ferret runtime contract is intrinsic to the value or resource being implemented.

This package should be testable as normal Go code.

Examples of behavior that belongs in `core`:

- database connections and transactions;
- archive readers and writers;
- JWT signing and verification logic;
- OAuth2 clients and token handling;
- queryable and resource implementations;
- option validation and normalization;
- parsers, encoders, and decoders;
- provider and protocol clients;
- error normalization tied to the module domain.

### `lib/`

The `lib` package contains the Ferret-facing function bridge.

Keep this layer thin. Its job is to validate and decode Ferret arguments, call `core`, convert results back into runtime values, and register functions.

Avoid putting main implementation logic in `lib`. If a Ferret function contains meaningful domain behavior, move that behavior into `core` and let `lib` adapt it.

Examples of behavior that belongs in `lib`:

- Ferret function declarations;
- arity and argument validation;
- SDK-based argument decoding;
- runtime value conversion;
- function registration helpers;
- small wrappers around `core` operations;
- Ferret-specific error context at the function boundary.

### Top-level module package

The top-level package wires the module together.

It owns:

- the public module constructor;
- public configuration options;
- module identity passed to `sdk.NewModule` or the current equivalent;
- namespace selection and registration orchestration;
- package documentation;
- module-level integration tests where appropriate.

Keep registration atomic where the SDK supports it. Prefer `sdk.RegisterFunctions` over ad hoc sequential registration when defining a Ferret function set.

### Dependency direction

Keep dependencies flowing in one direction:

```text
top-level module package
        ↓
       lib
        ↓
      core
```

- `core` must not import `lib`.
- `core` should generally not know Ferret function names, namespaces, or registration details.
- `lib` translates between Ferret runtime contracts and module implementation behavior.
- The top-level package configures and registers the module without absorbing `core` or `lib` responsibilities.

### Namespace naming

Use a namespace that matches the module category and purpose.

Examples:

```text
DB::SQLITE
SECURITY::JWT
SECURITY::OAUTH2
DOCUMENT::PDF
WEB::HTML
ARCHIVE
```

Prefer namespaces that leave room for related modules under the same umbrella. Treat an existing namespace and its function names as compatibility-sensitive public behavior.

### Rule of thumb

- If the code would still be useful without Ferret function registration, it probably belongs in `core`.
- If the code exists only to expose behavior as a Ferret function, it belongs in `lib`.
- If the code constructs, configures, registers, documents, or packages the module, it belongs in the top-level package.
- If code is useful to multiple modules, keep it local until there is a substantial and stable reason to publish it through `pkg/`.

## Public API and module boundary rules

- Treat module constructors, public options, exported resource/value types, FQL namespaces and functions, manifests, and support-package exports as API-sensitive.
- Do not export new symbols unless the task explicitly requires an external contract.
- Prefer unexported helpers inside the owning package before adding exported APIs.
- If a new exported symbol is necessary, add a doc comment that explains its external contract and ownership or lifecycle expectations.
- Do not move internals into `pkg/common` only to make tests or cross-module access easier.
- Do not introduce cross-module imports when a module-local implementation is sufficient.
- Preserve public FQL behavior during refactors, including argument validation, option defaults, return value types, errors, and resource semantics.
- Any intentional backward-incompatible Go or FQL behavior change must be called out explicitly, documented, and covered by tests showing the new expected behavior.
- Keep module README examples and `ferret.yaml` metadata aligned with actual public behavior.
- Do not change manifest versions as a side effect of ordinary feature work unless the task explicitly includes release preparation.

## Ferret binding rules

- Keep Ferret-facing argument validation close to the `lib` function boundary.
- Prefer the current SDK helpers such as `sdk.NewModule`, `sdk.RegisterFunctions`, `sdk.Func`, `sdk.DecodeArg`, `sdk.DecodeArgOr`, and strict structured decoding where they match the contract.
- Do not replace a precise existing validation contract with permissive generic conversion.
- Keep host functions small and delegate domain behavior to `core`.
- Do not duplicate Ferret runtime value semantics inside a module.
- Preserve argument position, arity, and function context in user-facing errors where practical.
- Test public functions at the Ferret-language level whenever practical, especially when value conversion or registration is part of the behavior.

## Resource, filesystem, and lifecycle rules

- Module runtime behavior that accesses user-controlled paths must use `github.com/MontFerret/ferret/v2/pkg/fs` resolved from the execution context.
- Do not use direct host filesystem calls such as `os.Open`, `os.ReadFile`, `os.WriteFile`, `os.Stat`, `os.Mkdir`, `os.Create`, or third-party path-based open/save APIs for user-controlled runtime paths unless access is first mediated by Ferret's filesystem policy.
- Direct host filesystem access is acceptable in tests and repository tooling when it is not module runtime behavior and the target is appropriately scoped.
- Values or handles that own resources must make ownership and cleanup behavior explicit.
- Cleanup must be deterministic when the contract exposes `Close` or an equivalent lifecycle method.
- Preserve cleanup on normal return, error, partial initialization, and cancellation paths.
- Propagate `context.Context` through blocking network, database, browser, parsing, and provider operations when the underlying API supports it.
- Preserve cancellation and deadline errors so callers can recognize them.
- Do not materialize lazy or streaming values eagerly unless the task explicitly requires it.
- When adapting third-party libraries, verify whether constructors, iterators, readers, responses, transactions, or clients require explicit cleanup.
- Do not create process-global resource registries when execution-scoped Ferret lifecycle hooks can own the resource.

## Error quality rules

- User-facing errors should identify the module operation or Ferret function that failed when context is available.
- Preserve actionable argument, option, path, provider, protocol, or resource context without exposing credentials, tokens, or other secrets.
- Keep domain and external-service error handling in `core`; let `lib` add Ferret-boundary context without discarding the underlying cause.
- Do not replace specific errors with generic failures.
- Preserve `context.Canceled` and `context.DeadlineExceeded` through wrapping and translation.
- Distinguish invalid user input, unsupported behavior, external failures, and internal invariant violations when the distinction affects callers.
- Tests for changed error behavior should verify the useful contract rather than brittle full strings unless the exact text is public behavior.

## Module design guidance

These rules are mandatory unless the task explicitly requires otherwise.

- Keep module ownership local.
- Prefer implementing functionality inside the smallest responsible module and layer.
- Do not add repository-wide abstractions for behavior used by one module.
- Share code across modules only when:
    - the duplication is substantial;
    - the abstraction is clearly stable;
    - the shared package has a clean ownership and versioning story.
- Avoid introducing cross-module imports unless there is a strong, explicit reason.
- A module should remain understandable without inspecting unrelated modules.
- Public behavior exposed by a module should be documented in that module's README.
- Adapt an established local pattern before introducing a new module architecture.
- Do not assume that a pattern from one module automatically belongs in another; verify the requirements and current contracts first.

## Go type and file structure rules

These rules are mandatory unless the task explicitly requires otherwise.

- Do not define multiple method-bearing structs in the same `.go` file.
- Prefer declaring a method-bearing struct as a standalone `type Name struct { ... }`.
- A method-bearing struct should usually live in its own file, named after the primary type or responsibility whenever practical, for example:
    - `client.go` for `Client`;
    - `document.go` for `Document`;
    - `iterator.go` for `Iterator`.
- Grouped `type ( ... )` declarations are allowed for interfaces, passive data-only structs, and other small related helper/value types that belong to the same narrow concern.
- A grouped `type ( ... )` block may also contain exactly one method-bearing struct when:
    - it is the only behavioral type in the file, and
    - the other grouped types are passive helper/value types from the same narrow concern.
- Do not use grouped `type ( ... )` declarations to hide multiple substantial behavioral types.
- If a helper struct later gains methods and would create more than one method-bearing struct in the file, extract it into its own file immediately.
- Methods for a struct should live in the same file as the struct unless there is a strong, explicit reason to split by concern.
- Do not place a new method-bearing struct into an existing file just because the code compiles.

Allowed:

```go
type (
	ParseResult struct {
		Metadata map[string]any
		Modified bool
	}
	ParseOptions struct {
		Strict bool
	}
	Parser interface {
		Parse(context.Context, []byte, ParseOptions) (ParseResult, error)
	}
)
```

Avoid:

```go
type (
	Client struct {
		// ...
	}
	requestState struct {
		// ...
	}
)
```

Rationale:

- one method-bearing type per file keeps ownership of behavior obvious
- standalone method-bearing types make diffs and reviews clearer
- grouped type blocks are fine for passive, closely related types, but should not hide substantial behavioral types

Do not rewrite unrelated existing files solely to apply this rule. Apply it to new code and to touched structure when the requested change makes the ownership issue relevant.

## Function and method ownership rules

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
- Exported functions and methods should usually have doc comments, especially for module, package, and extension-facing APIs.
- Unexported functions and methods should be commented only when they carry non-obvious behavior, invariants, side effects, ownership rules, cleanup expectations, or protocol/lifecycle constraints.
- Comments must explain intent, contract, invariants, side effects, or lifecycle behavior.
- Prefer comments that explain why the code exists, what must remain true, or how it is meant to be used.
- Do not write comments that merely restate the method name or signature.
- For resources, streaming values, protocols, and external clients, prefer comments about semantics and ownership over implementation narration.
- Avoid comment wallpaper. Dense, meaningful comments are preferred over mechanically documenting obvious code.

Preferred:

```go
// Close releases resources associated with the document.
// It is safe to call multiple times. Once closed, the document must not be reused.
func (d *Document) Close() error
```

Preferred for internal code:

```go
// releaseOwned closes resources acquired during partial initialization
// so constructor failures preserve the all-or-nothing ownership contract.
func (s *openState) releaseOwned(...)
```

Avoid:

```go
// Close closes the document.
func (d *Document) Close() error
```

## Go control-flow spacing rules

These rules are mandatory for handwritten Go code.

Blank lines should separate logical units and make control-flow and termination boundaries visually obvious.

### Immediate producer + check

A declaration, assignment, function call, type assertion, lookup, parse operation, or similar statement may remain directly adjacent to a following `if` when the `if` immediately checks or consumes the value produced by that statement.

This includes error checks, boolean/result checks, type assertions, nil checks, bounds checks, and other immediate validation.

Preferred:

```go
result, err := client.Fetch(ctx)
if err != nil {
	return err
}
```

Preferred:

```go
resource, ok := value.(runtime.Resource)
if !ok || resource == nil {
	return ErrInvalidResource
}
```

Preferred:

```go
value := lookup(name)
if value == nil {
	return ErrNotFound
}
```

Preferred:

```go
count := len(items)
if count == 0 {
	return nil
}
```

The producer and its immediate check form one logical unit and should not be separated by a blank line.

### Separation from preceding logic

If an immediate producer + check unit follows another statement or logical unit, separate it from the preceding code with a blank line.

Preferred:

```go
prepareRequest()

response, err := client.Do(req)
if err != nil {
	return err
}
```

Avoid:

```go
prepareRequest()
response, err := client.Do(req)
if err != nil {
	return err
}
```

No leading blank line is required when the producer begins the enclosing block:

```go
func fetch(ctx context.Context, client *Client) error {
	response, err := client.Fetch(ctx)
	if err != nil {
		return err
	}

	return consume(response)
}
```

### Consecutive control-flow blocks

Separate independent `if` statements with a blank line.

Avoid:

```go
if source != nil {
	useSource(source)
}
if destination != nil {
	useDestination(destination)
}
```

Prefer:

```go
if source != nil {
	useSource(source)
}

if destination != nil {
	useDestination(destination)
}
```

This applies even when both conditions are short. Independent control-flow decisions should remain visually distinct.

### Statements after control flow

Add a blank line after a completed `if` block before continuing with a separate statement or logical unit.

Avoid:

```go
if cached {
	useCached()
}
loadFresh()
```

Prefer:

```go
if cached {
	useCached()
}

loadFresh()
```

### Return and break separation

`return` and `break` are termination or control-transfer statements and should be visually separated from preceding statements.

A `return` or `break` must begin a new logical group: when another statement precedes it in the same block, place a blank line immediately before it.

This rule applies inside nested control-flow blocks as well as at the function-body level.

Avoid:

```go
if fallback {
	result = defaultResult
}
return result
```

Prefer:

```go
if fallback {
	result = defaultResult
}

return result
```

Avoid:

```go
if match {
	found = true
	break
}
```

Prefer:

```go
if match {
	found = true

	break
}
```

The same rule applies when ordinary computation precedes a return:

Avoid:

```go
result := buildResult()
return result
```

Prefer:

```go
result := buildResult()

return result
```

Likewise for `break`:

Avoid:

```go
found = true
break
```

Prefer:

```go
found = true

break
```

No blank line is required before a `return` when it is already the first statement in its block:

```go
if err != nil {
	return err
}
```

No artificial leading blank line should be introduced:

```go
func value() int {
	return 42
}
```

The intent is not to surround every `return` or `break` with whitespace. The rule specifically requires separation from a preceding statement in the same block.

## Local type declarations

Local types declared inside functions are allowed, but should be used deliberately.

Prefer a local type when all of the following are true:

- it is small;
- it is passive and method-free;
- it is used only within that function;
- it exists purely to support the local algorithm;
- keeping it local makes the function easier to understand rather than harder to scan.

Prefer a package-level unexported type when one or more of the following are true:

- the type represents a meaningful domain or algorithmic concept;
- the type is used across a substantial portion of a long or complex function;
- moving the type declaration out of the control flow improves readability;
- the type may reasonably gain methods or behavior;
- the type is likely to be reused by nearby helpers;
- the type name helps explain the algorithm or responsibility at package scope.

Do not promote a tiny throwaway struct to package scope merely for consistency.

Do not keep a meaningful concept local merely to avoid adding a package-level type.

Example of an appropriate local type:

```go
func collect(...) {
	type entry struct {
		name  string
		index int
	}

	// Small, passive, function-local algorithm state.
}
```

Example where a package-level type is preferable:

```go
type archiveEntryCandidate struct {
	name     string
	format   Format
	size     int64
	priority int
}
```

when that value represents a meaningful candidate selected and ranked throughout a substantial archive inspection or extraction algorithm.

The decision should be based on readability, conceptual ownership, and expected evolution, not on a blanket preference for either local or package-level types.

## Response and code style

When assisting with this repository, avoid large unstructured blocks of prose or code.

Prefer responses that are easy to scan:

- Use short sections with clear headings.
- Use bullet points for decisions, trade-offs, and follow-up work.
- Use code blocks only for actual code, commands, or configuration.
- Prefer focused snippets or diffs over full-file dumps.
- Explain why a change is needed before showing how to implement it.
- Keep comments in code useful and minimal.
- Avoid repeating the same context in multiple places.
- When the change touches multiple files, summarize the role of each file first.

The expected tone is practical, concise, and engineering-focused.

## Development practice expectations

Agents must follow repository-specific engineering discipline rather than generic style preferences.

### Core principles

- Preserve correctness first.
- Preserve module, package, and layer ownership boundaries.
- Prefer the smallest local change that fully solves the task.
- Avoid introducing abstractions, indirection, or refactors unless necessary for correctness, maintainability, or an explicitly requested design change.
- Do not optimize by intuition alone; use measurements for performance-sensitive work.
- Keep behavioral ownership obvious in code structure, naming, and file layout.
- Prefer changing one module over changing many modules unless the task truly spans them.
- Do not treat the first working implementation as final.
- A task is complete only after implementation, validation, self-review, necessary corrections, final validation, and complete diff inspection.

### Mandatory expectations

- Identify the owning module, support package, or repository surface before making a non-trivial change.
- Identify the public contract, invariant, or internal behavior being preserved or changed.
- Preserve existing behavior unless the task explicitly requires changing it.
- Add or update tests for any behavior change.
- Add or update benchmarks for any significant performance-sensitive change.
- Run the narrowest relevant validation first, then broaden as appropriate.
- Perform the mandatory final self-review for every non-trivial task.
- Inspect the complete final diff before declaring the task complete.
- Re-run affected validation after any review-driven changes.
- Do not claim tests, benchmarks, review, or validation were completed unless they were actually performed.
- Do not perform opportunistic refactors unrelated to the requested task unless required for correctness.
- Distinguish code failures from missing network, browser, service, toolchain, or sandbox prerequisites.

### Required workflow for non-trivial changes

Before and while making a non-trivial change, agents must:

1. Identify the owning module, support package, or repository surface.
2. Identify the contract, invariant, or behavior being preserved or changed.
3. Choose the smallest reasonable implementation that fits the existing design.
4. Determine whether the change affects one module, shared packages, workspace tooling, or upstream Ferret behavior.
5. Determine whether the change is performance-significant.
6. Capture a benchmark baseline when required.
7. Add or update correctness tests.
8. Add or update benchmarks when required.
9. Run relevant validation.
10. Perform the mandatory final self-review.
11. Address issues discovered during self-review.
12. Re-run validation and benchmarks affected by review-driven changes.
13. Inspect the complete final diff and summarize the results accurately.

Do not consider a task complete merely because the implementation compiles and its first tests pass.

## Mandatory final self-review

After implementation and initial validation for any non-trivial task, review the complete resulting change before considering the task finished.

The review must evaluate the implementation itself, not merely confirm that tests pass. It must not justify unrelated refactoring or redesign.

### Correctness

- Verify that the implementation completely satisfies the task requirements.
- Look for missing cases, incorrect assumptions, regressions, boundary conditions, and failure paths.
- Check validation, error handling, cancellation, cleanup, state transitions, ownership, and lifecycle behavior where applicable.
- Verify concurrency behavior when relevant.
- Verify that public Go and FQL semantics match the intended contract.
- Verify that tests exercise intended behavior rather than merely mirroring the implementation.
- For bug fixes, add a regression test that would fail without the fix whenever practical.

### Code clarity and cleanliness

- Look for unnecessary complexity, duplication, excessive nesting, awkward control flow, misleading naming, and code that is hard to reason about.
- Prefer straightforward, idiomatic Go over clever implementations.
- Remove temporary artifacts, dead branches, obsolete helpers, debugging code, and comments describing abandoned approaches.
- Avoid unnecessary abstraction layers and indirection.
- Ensure the main execution path remains easy to follow.

### Repository and Go best practices

- Verify compliance with this file and the owning module or package conventions.
- Check error handling, API shape, resource ownership, concurrency, synchronization, context propagation, and lifecycle management where relevant.
- Check whether errors are wrapped or propagated appropriately.
- Check whether resources can leak on failure, cancellation, partial initialization, or early return.
- Check whether ownership expectations are explicit where needed.
- Do not introduce a pattern merely because it is fashionable elsewhere; it must improve this repository specifically.

### Architecture

- Verify that responsibilities remain in the correct module, package, layer, type, and file.
- Check the top-level to `lib` to `core` dependency direction.
- Look for unwanted coupling, leaked implementation details, duplicated semantics, or abstractions at the wrong level.
- Verify that Ferret runtime semantics are consumed rather than redefined by modules.
- Verify that public or shared APIs are introduced only when genuinely required.
- Verify that user-controlled file access remains mediated by Ferret's filesystem policy.
- Consider whether the design remains understandable and maintainable as the module evolves.

### Code organization and split

- Verify that files, types, methods, functions, and packages have clear responsibilities.
- Check the Go type/file and function/method ownership rules in this file.
- Look for files, functions, or types doing too much.
- Look for unrelated responsibilities grouped together.
- Avoid unnecessary fragmentation into excessive helpers, files, or abstractions.
- Keep helpers at the narrowest appropriate ownership level.
- Ensure behavioral ownership is obvious from code layout.

### Tests

- Review coverage for meaningful behavioral gaps.
- Look especially for missing negative cases, edge conditions, invalid input, cleanup paths, cancellation paths, partial failures, and boundary values.
- Check for brittle tests unnecessarily coupled to implementation details.
- Check for redundant tests that add maintenance cost without meaningful coverage.
- Check for weak assertions that allow plausible regressions to pass.
- Verify FQL integration coverage when user-visible behavior crosses the Ferret boundary.
- Verify external-service and browser behavior at the appropriate layer when the task affects it.

### Performance

For significant changes:

- inspect for accidental allocations, repeated work, unnecessary materialization, unnecessary synchronization, or added hot-path overhead;
- compare the required benchmark results against the recorded baseline;
- verify that benchmark differences are attributable to the implementation rather than a changed setup;
- do not trade clear correctness or maintainability for speculative micro-optimization;
- investigate meaningful regressions before considering the task complete.

### Review findings and remediation

When self-review finds a problem:

1. Fix correctness issues and regressions.
2. Fix meaningful ownership, architecture, lifecycle, API, or maintainability problems.
3. Simplify unnecessarily complicated code when it clearly improves the implementation.
4. Correct relevant file, type, or method ownership violations.
5. Add or improve tests when review exposes a coverage gap.
6. Re-run validation affected by the correction.
7. Re-run relevant benchmarks when the correction affects benchmarked code.

Do not leave a known correctness, architecture, ownership, lifecycle, or significant test-coverage problem unresolved merely because the initial implementation works.

Minor stylistic preferences do not require changes. Distinguish actual problems from optional preferences, and leave existing clear and correct code alone.

Do not use self-review to justify:

- speculative refactoring;
- unrelated cleanup or API redesign;
- broad module or package reshuffling;
- rewriting existing code merely for stylistic consistency;
- abstractions without a concrete need;
- FQL behavior changes outside the task.

### Final diff inspection

Immediately before finishing a non-trivial task, inspect the complete final diff as a whole.

Verify that:

- every changed line is relevant to the requested task or a necessary supporting change;
- no temporary or debugging code remains;
- no accidental FQL, Go API, manifest, dependency, or release change was introduced;
- no unrelated refactor slipped into the change;
- generated or derived files changed only when their source inputs required it;
- tests describe intended behavior rather than implementation details;
- comments describe current contracts and invariants rather than abandoned approaches;
- module, package, file, type, and function boundaries remain coherent;
- resource ownership and lifecycle behavior remain correct;
- the result is the smallest coherent change that fully solves the task.

If final diff inspection reveals an issue, correct it and repeat the affected validation before finishing.

## Significant changes

A change is significant when it could reasonably affect:

- execution throughput or common-path latency;
- allocation patterns or memory use;
- streaming, iteration, materialization, parsing, encoding, or decoding cost;
- database, HTTP, browser, provider, or protocol client hot paths;
- resource reuse, pooling, caching, cleanup, or synchronization;
- shared support-package behavior used by multiple modules.

This includes, but is not limited to, changes in:

- parsers, encoders, decoders, extractors, and matchers;
- query, row, document, archive, and stream iteration;
- database execution and result conversion;
- browser event loops and high-frequency driver paths;
- buffering, resource tracking, caching, and shared stream helpers;
- frequently called Ferret bindings when conversion or validation cost materially changes.

This usually does not include:

- comment-only, documentation-only, or formatting-only edits;
- pure renames with no behavior change;
- test-only changes;
- narrow refactors that do not affect behavior or hot paths.

When in doubt, treat the change as significant and measure it.

### Benchmark workflow for significant changes

For significant changes, agents must:

- run relevant benchmarks before implementation and save the results as a baseline;
- implement the change;
- run the same benchmarks afterward under the same setup;
- compare results, preferably including `ns/op`, `B/op`, and `allocs/op`;
- report the command and summarize the performance delta.

If no relevant benchmark exists for a deterministic hot path, add one.

External-service or browser timing is not automatically a useful benchmark. Prefer deterministic coverage for module-owned work and use service-backed tests for correctness. If a meaningful benchmark or required environment is unavailable, state that explicitly and do not claim benchmark validation was completed.

## Test placement rules

- `core` behavior should have package-level unit tests that do not require Ferret registration when practical.
- `lib` behavior should test function registration, arity, argument decoding, runtime value conversion, and Ferret-specific errors.
- Top-level module tests should cover constructors, options, namespace wiring, lifecycle hooks, and public integration behavior as appropriate.
- User-visible FQL behavior should have fixtures under `tests/modules/<module>` when the runtime harness can exercise it meaningfully.
- Support packages under `pkg/` should have package-local tests plus affected-module validation when their contract changes.
- Manifest behavior should be covered through `make validate-manifests` and tests under `tests/manifests`.
- Publisher behavior should be tested in the isolated `scripts/publish` module.
- Release transaction behavior should be tested through the release test harness rather than a real tag or push.
- Browser-backed behavior should use the existing Lab and Chromium integration flow.
- Postgres, Redis, or other service-backed behavior should mirror the relevant workflow's service and environment contract.
- Resource behavior should test normal close, repeated close when supported, failure cleanup, cancellation, and partial initialization where applicable.
- Tests should verify public contracts and meaningful outcomes, not private implementation details.

## Validation and evidence

When finishing a non-trivial change, agents must report:

- owning module, support package, or repository surface;
- files changed;
- tests added or updated;
- benchmarks added or updated, when applicable;
- validation commands actually run;
- benchmark commands and before/after results, when applicable;
- confirmation that self-review and complete diff inspection were performed;
- notable issues found and corrected during review, if any;
- invariants preserved or intentionally changed;
- remaining concerns, environment limitations, or skipped validation.

For significant changes, tests alone are not sufficient. Correctness tests and relevant benchmarks are required when the environment permits meaningful measurement.

Do not claim:

- tests passed unless they were actually run;
- benchmarks completed unless they were actually run;
- self-review completed unless the implementation and complete diff were inspected;
- validation succeeded when commands failed or were skipped;
- an environment-dependent failure is a product regression without evidence.

If validation, benchmarking, or review cannot be completed because of tooling, network, sandbox, browser, or service limitations, state that explicitly.

## Change discipline

- Prefer adapting an existing local pattern over introducing a new architecture.
- Do not add helper layers, wrappers, interfaces, or abstractions only for aesthetic reasons.
- Do not move code across modules or packages unless the ownership boundary is genuinely wrong.
- Keep diffs focused on the requested task.
- If cleanup is necessary to make the change safe, keep it tightly scoped and explain why.
- Self-review must not expand task scope unless a discovered problem directly affects correctness, safety, architecture, lifecycle, or maintainability of the requested change.
- Preserve unrelated user changes in a dirty worktree.

## Comment and documentation discipline

- Add comments where semantics, invariants, side effects, ownership, lifecycle, or recovery behavior are non-obvious.
- Do not add comment wallpaper.
- Prefer comments that explain why, contract, or invariants rather than implementation narration.
- Document public module and support-package behavior more carefully than local obvious helpers.
- Keep module README examples, Go exports, FQL functions, and manifests consistent.
- Update documentation in the same change when public behavior changes.

## Decision bias when uncertain

When uncertain:

- preserve existing behavior;
- prefer the smaller local change;
- add a focused test;
- prefer module-local ownership;
- treat a change as significant when performance might be affected;
- verify whether behavior belongs in Contrib or the Ferret core repository;
- verify ownership before adding a shared package or cross-module dependency;
- fix actual review findings rather than performing speculative cleanup;
- leave already-correct code alone.

## Tooling prerequisites

- Go must be installed at the versions required by `go.work` and the owning `go.mod`.
- `make` is the preferred entrypoint for repository-defined workflows.
- Module and package lint/format flows require:
    - `staticcheck`;
    - `fieldalignment`;
    - `goimports`;
    - `revive`.
- Install the pinned manifest validator through `make install-manifest-validator` before manifest validation.
- FQL integration tests require Ferret Lab.
- `web/html` integration requires Chromium's remote debugging endpoint at `http://127.0.0.1:9222`.
- Service-backed module tests may require Postgres, Redis, credentials, network access, or other workflow-defined prerequisites.
- Use `make install-tools` when the task authorizes installing the repository's development tools.

## Command matrix

Use the repository commands below rather than ad hoc workspace loops.

### Discovery

- Modules: `make modules`
- Support packages: `make packages`

### Modules

- Build the runtime harness and selected modules: `make build [module ...]`
- Unit test selected modules: `make test-unit [module ...]`
- Lint selected modules: `make lint [module ...]`
- Format selected modules: `make fmt [module ...]`
- Run selected FQL integration suites: `make test-integration [module ...]`

Build the runtime harness before integration tests:

```sh
make build [module ...]
make test-integration [module ...]
```

For all modules, use the explicit build, unit, and integration sequence:

```sh
make build
make test-unit
make test-integration
```

The `Makefile` declares `make test` as an all-module aggregate, not a targeted command. Its current goal-forwarding logic passes `test` to the module selector, so use the explicit sequence above unless that Makefile behavior is corrected. Never use `make test <module>`.

Targeted validation uses `make test-unit <module>` and, when applicable, `make build <module>` followed by `make test-integration <module>`.

### Support packages

- Build selected packages: `make build-packages [package ...]`
- Unit test selected packages: `make test-packages [package ...]`
- Lint selected packages: `make lint-packages [package ...]`
- Format selected packages: `make fmt-packages [package ...]`

### Repository tooling

- Validate all module manifests and repository manifest conventions: `make validate-manifests`
- Test Barn publisher preparation: `make test-publisher`
- Test release transactions: `make test-release`
- Install lint/format and manifest tools: `make install-tools`

Run the narrowest owning command first. Broaden to affected modules, packages, manifests, integration suites, and repository tooling according to the scope of the change.

## Editing rules

- Treat `Makefile`, `scripts/modules.sh`, and `scripts/packages.sh` as the sources of truth for local command behavior.
- Treat `.github/workflows` as the source of truth for CI service setup and matrix behavior.
- Do not manually maintain module or package inventory lists in automation when discovery already owns them.
- Update `go.work` when workspace members change.
- Update the root README index when public module or package inventory changes.
- Register new integration-tested modules in `tests/runtime/main.go`.
- Treat `ferret.yaml` as a source publication contract, not generated output.
- Use the owning update or release scripts for dependency-wide or version transaction changes.
- Do not hand-edit workspace vendor output; refresh it through the owning dependency workflow or `go work vendor` when required.
- Run `make fmt` or `make fmt-packages` only when formatting changes are intended because these commands rewrite files.
- Do not perform real commits, tags, pushes, releases, or publications unless explicitly requested.
- Do not rewrite unrelated code solely to enforce the style rules in this file.

## Validation expectations

- After code changes, run the narrowest tests that prove the behavior touched.
- Run targeted module or package build, lint, and format flows as appropriate.
- Build the runtime harness before FQL integration testing.
- Add FQL integration coverage when user-visible behavior crosses the module boundary.
- Validate manifests when module identity, metadata, versions, or inventory changes.
- Validate publisher or release tooling through its dedicated test target when those surfaces change.
- Validate affected module consumers after changing a shared support package.
- Capture and compare benchmarks for significant changes.
- Re-run affected validation after the last review-driven implementation change.
- Inspect the complete final diff after all corrections.
- For documentation-only changes, tests are normally unnecessary; verify referenced paths and commands, run `git diff --check`, and inspect the final diff.
- Report missing tools, network, Chromium, Lab, Postgres, Redis, or sandbox access separately from code failures.

### Expectations for non-trivial changes

When proposing or implementing non-trivial changes:

- identify the owning module, package, or repository surface first;
- preserve public contracts and invariants unless the task explicitly changes them;
- prefer local, comprehensible changes before new abstractions;
- distinguish correctness work from performance work;
- complete the mandatory final self-review;
- inspect the final diff after review-driven corrections;
- re-run affected validation after the last implementation change;
- summarize completed and skipped evidence accurately.

## Secondary references

- Root `README.md` for the module inventory, development entrypoints, manifests, and release overview.
- Module README files for public module behavior and FQL examples.
- `ferret.yaml` files and the pinned Specs validator for publication contracts.
- `go.work` and each `go.mod` for workspace and dependency truth.
- `.github/workflows` for CI and service-backed integration behavior.
- The Ferret repository for core runtime and extension API ownership; do not apply its core command matrix or package map to this repository.
