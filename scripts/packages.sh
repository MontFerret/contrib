#!/usr/bin/env bash
set -euo pipefail

DIR_PACKAGES="./pkg"

get_packages() {
  find "$DIR_PACKAGES" -mindepth 2 -maxdepth 2 -type f -name go.mod \
    -exec dirname {} \; \
    | sed "s|^$DIR_PACKAGES/||" \
    | sort
}

usage() {
  echo "Usage:"
  echo "  $0 <list|build|test-unit|lint|fmt> [package ...]"
}

package_exists() {
  local target="$1"
  local package

  while IFS= read -r package; do
    if [ "$package" = "$target" ]; then
      return 0
    fi
  done < <(get_packages)

  return 1
}

run_fieldalignment() {
  local output
  local status

  set +e
  output="$(fieldalignment --fix ./... 2>&1)"
  status="$?"
  set -e

  if [ "$status" -eq 0 ]; then
    if [ -n "$output" ]; then
      echo "$output"
    fi

    return 0
  fi

  if [[ "$output" == *"matched no packages"* ]]; then
    return 0
  fi

  echo "$output" >&2

  return "$status"
}

main() {
  if [ "$#" -lt 1 ]; then
    usage
    exit 1
  fi

  local command="$1"
  shift

  if [ "$command" = "list" ]; then
    get_packages
    exit 0
  fi

  if [ "$#" -eq 0 ]; then
    while IFS= read -r package; do
      set -- "$@" "$package"
    done < <(get_packages)
  else
    for package in "$@"; do
      if ! package_exists "$package"; then
        echo "Unknown package: $package" >&2
        echo "Available packages:" >&2
        get_packages >&2
        exit 1
      fi
    done
  fi

  local root_dir
  root_dir="$(pwd)"

  for package in "$@"; do
    case "$command" in
      build)
        echo "Building package '$package'"
        ( cd "$DIR_PACKAGES/$package" && go build ./... )
        ;;
      test-unit)
        echo "Testing package '$package'"
        ( cd "$DIR_PACKAGES/$package" && go test ./... )
        ;;
      lint)
        echo "Linting package '$package'"
        ( cd "$DIR_PACKAGES/$package" && \
          staticcheck -tests=false -checks=all,-U1000 ./... && \
          revive -config "$root_dir/revive.toml" -formatter stylish \
            -exclude './*_test.go' \
            ./...
        )
        ;;
      fmt)
        echo "Formatting package '$package'"
        ( cd "$DIR_PACKAGES/$package" && \
          run_fieldalignment && \
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
}

main "$@"
