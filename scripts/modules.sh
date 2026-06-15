#!/usr/bin/env bash
set -euo pipefail

DIR_BIN="./bin"
DIR_MODULES="./modules"
DIR_TESTS="./tests"
DIR_RUNTIME="$DIR_TESTS/runtime"
DIR_TESTDATA="$DIR_TESTS/data"
DIR_MODULE_TESTS="$DIR_TESTS/modules"
TAG_MODULES="modules"

get_modules() {
  find "$DIR_MODULES" -type f -name go.mod \
    -exec dirname {} \; \
    | sed "s|^$DIR_MODULES/||" \
    | sort
}

usage() {
  echo "Usage:"
  echo "  $0 <list|build|test|lint|fmt|versions|deps> [module ...]"
}

get_direct_dependencies() {
  local module="$1"
  local go_mod="$DIR_MODULES/$module/go.mod"

  if [[ ! -f "$go_mod" ]]; then
    return
  fi

  # Extract the module path prefix from go.mod
  local module_prefix
  module_prefix=$(grep -E '^module ' "$go_mod" | awk '{print $2}' | sed 's|/modules/.*||')

  # Find all require statements that reference internal modules
  grep -E '^\s+'"$module_prefix"'/modules/' "$go_mod" 2>/dev/null | \
    awk '{print $1}' | \
    sed 's|^.*/modules/||' || true
}

get_all_dependencies() {
  local module="$1"
  local indent="${2:-}"
  local visited_list="${3:-}"

  # Check if already visited (avoid cycles)
  if echo "$visited_list" | grep -q "^${module}$"; then
    return
  fi

  # Add to visited list
  visited_list="${visited_list}${module}"$'\n'

  # Get direct dependencies and process them
  while IFS= read -r dep; do
    # Skip empty entries
    if [[ -z "$dep" ]]; then
      continue
    fi

    echo "${indent}${dep}"

    # Recursively get dependencies of this dependency
    get_all_dependencies "$dep" "  ${indent}" "$visited_list"
  done < <(get_direct_dependencies "$module")
}

module_exists() {
  local target="$1"
  local module

  while IFS= read -r module; do
    if [ "$module" = "$target" ]; then
      return 0
    fi
  done < <(get_modules)

  return 1
}

get_versions() {
  if [[ $# -ne 1 ]]; then
    usage
    exit 1
  fi

  local module="$1"

  # Get all tags for the module, sort by version (descending), and limit to 10
  git tag --list "$TAG_MODULES/$module/v*" --sort=-version:refname | head -n 10 | while IFS= read -r tag; do
    # Extract just the version from the tag (remove the prefix)
    echo "${tag##*/}"
  done
}

main() {
  if [ "$#" -lt 1 ]; then
    usage
    exit 1
  fi

  local command="$1"
  shift

  if [ "$command" = "list" ]; then
    get_modules
    exit 0
  fi

  if [ "$#" -eq 0 ]; then
    while IFS= read -r module; do
      set -- "$@" "$module"
    done < <(get_modules)
  else
    for module in "$@"; do
      if ! module_exists "$module"; then
        echo "Unknown module: $module" >&2
        echo "Available modules:" >&2
        get_modules >&2
        exit 1
      fi
    done
  fi

  local root_dir
  root_dir="$(pwd)"

  for module in "$@"; do
    case "$command" in
      build)
        echo "Building module '$module'"
        ( cd "$DIR_MODULES/$module" && go build ./... )
        ;;
      test-unit)
        echo "Testing module '$module'"
        ( cd "$DIR_MODULES/$module" && go test ./... )
        ;;
      test-integration)
        echo "Running integration tests for module '$module'"

        local runtime_uri="bin://$root_dir/${DIR_BIN#./}/runtime"
        local module_test_files="$root_dir/${DIR_MODULE_TESTS#./}/$module"
        local serve_dynamic="$root_dir/${DIR_TESTDATA#./}/pages/dynamic"
        local serve_static="$root_dir/${DIR_TESTDATA#./}/pages/static"

        # If the module test folder doesn't exist, skip it
        if [ ! -d "$module_test_files" ]; then
          echo "No integration tests found for module '$module', skipping."
          continue
        fi

        lab run \
          --runtime="$runtime_uri" \
          --timeout=120 \
          --attempts=5 \
          --concurrency=1 \
          --wait=http://127.0.0.1:9222/json/version \
          --files="$module_test_files" \
          --serve="$serve_dynamic" \
          --serve="$serve_static"
        ;;
      lint)
        echo "Linting module '$module'"
        ( cd "$DIR_MODULES/$module" && \
          staticcheck -tests=false -checks=all,-U1000 ./... && \
          revive -config "$root_dir/revive.toml" -formatter stylish \
            -exclude './*_test.go' \
            ./...
        )
        ;;
      fmt)
        echo "Formatting module '$module'"
        ( cd "$DIR_MODULES/$module" && \
          fieldalignment --fix ./... && \
          go fmt ./... && \
          goimports -w -local github.com/MontFerret .
        )
        ;;
      versions)
        echo "Listing versions for module '$module'"
        get_versions $module
        ;;
      deps)
        echo "Listing module dependencies for '$module'"
        get_all_dependencies "$module" "" ""
        ;;
      *)
        echo "Unknown command: $command" >&2
        usage
        exit 1
        ;;
    esac
  done

  if [ "$command" = "build" ]; then
    echo "Building runtime..."
    mkdir -p "$DIR_BIN"
    go build -v -o "$DIR_BIN/runtime" "$DIR_RUNTIME/runtime.go"
  fi
}

main "$@"
