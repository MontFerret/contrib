#!/usr/bin/env bash
set -euo pipefail

DIR_MODULES="./modules"

get_modules() {
  find "$DIR_MODULES" -mindepth 1 -maxdepth 1 -type d -exec basename {} \; | sort
}

usage() {
  echo "Usage:"
  echo "  $0 <build|test|lint|fmt> [module ...]"
}

main() {
  if [[ $# -lt 1 ]]; then
    usage
    exit 1
  fi

  local command="$1"
  shift

  all_modules="$(get_modules)"

  local selected_modules=()
  if [[ $# -eq 0 ]]; then
    selected_modules=("${all_modules[@]}")
  else
    for module in "$@"; do
      if [[ ! -d "$DIR_MODULES/$module" ]]; then
        echo "Unknown module: $module" >&2
        echo "Available modules: ${all_modules[*]}" >&2
        exit 1
      fi
      selected_modules+=("$module")
    done
  fi

  for module in "${selected_modules[@]}"; do
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