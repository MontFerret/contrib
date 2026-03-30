#!/usr/bin/env bash
set -euo pipefail

DIR_MODULES="./modules"

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
    local modules=()
    mapfile -t modules < <(get_modules)
    set -- "${modules[@]}"
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
      test)
        echo "Testing module '$module'"
        go test "$DIR_MODULES/$module/..."
        ;;
      lint)
        echo "Linting module '$module'"
        staticcheck -tests=false -checks=all "$DIR_MODULES/$module/..."
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
}

main "$@"