#!/usr/bin/env bash
set -euo pipefail

DIR_BIN="./bin"
DIR_MODULES="./modules"
DIR_TESTS="./tests"
DIR_RUNTIME="$DIR_TESTS/runtime"
DIR_TESTDATA="$DIR_TESTS/data"
DIR_MODULE_TESTS="$DIR_TESTS/modules"

get_modules() {
  find "$DIR_MODULES" -type f -name go.mod \
    -exec dirname {} \; \
    | sed "s|^$DIR_MODULES/||" \
    | sort
}

usage() {
  echo "Usage:"
  echo "  $0 <list|build|test|lint|fmt> [module ...]"
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
