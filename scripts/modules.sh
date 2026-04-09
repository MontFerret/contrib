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

  for module in "$@"; do
    case "$command" in
      build)
        echo "Building module '$module'"
        go build "$DIR_MODULES/$module/..."
        ;;
      test-unit)
        echo "Testing module '$module'"
        go test "$DIR_MODULES/$module/..."
        ;;
      test-integration)
        echo "Running integration tests for module '$module'"

        # If the module test folder doesn't exist, skip it
        if [ ! -d "$DIR_MODULE_TESTS/$module" ]; then
          echo "No integration tests found for module '$module', skipping."
          continue
        fi

        lab run \
          --runtime=bin://${DIR_BIN}/runtime \
          --timeout=120 \
          --attempts=5 \
          --concurrency=1 \
          --wait=http://127.0.0.1:9222/json/version \
          --files="$DIR_MODULE_TESTS/$module" \
          --serve=${DIR_TESTDATA}/pages/dynamic \
          --serve=${DIR_TESTDATA}/pages/static
        ;;
      lint)
        echo "Linting module '$module'"
        staticcheck -tests=false -checks=all,-U1000 "$DIR_MODULES/$module/..."
        revive -config revive.toml -formatter stylish \
          -exclude ./vendor/... \
          -exclude ./*_test.go \
          "$DIR_MODULES/$module/..."
        ;;
      fmt)
        echo "Formatting module '$module'"
        fieldalignment --fix "$DIR_MODULES/$module/..."
        go fmt "$DIR_MODULES/$module/..."
        goimports -w -local github.com/MontFerret "$DIR_MODULES/$module"
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
 	  go build -v -o ${DIR_BIN}/runtime ${DIR_RUNTIME}/runtime.go
  fi
}

main "$@"