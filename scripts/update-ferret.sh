#!/usr/bin/env bash
set -euo pipefail

DIR_MODULES="./modules"
DIR_PACKAGES="./pkg"
DIR_RUNTIME="./tests/runtime"
FERRET_MODULE="github.com/MontFerret/ferret/v2"

usage() {
  echo "Usage: $0 <semver>"
  echo "Examples:"
  echo "  $0 2.0.0-alpha.26"
  echo "  $0 v2.0.0-alpha.26"
}

normalize_version() {
  local version="$1"
  echo "${version#v}"
}

is_semver() {
  local version="$1"
  [[ "$version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$ ]]
}

get_go_mods() {
  find "$DIR_MODULES" "$DIR_PACKAGES" -mindepth 2 -maxdepth 3 -type f -name go.mod

  if [[ -f "$DIR_RUNTIME/go.mod" ]]; then
    echo "$DIR_RUNTIME/go.mod"
  fi
}

get_required_version() {
  local go_mod="$1"

  awk -v path="$FERRET_MODULE" '
    $1 == path && $2 != "" {
      print $2
      exit
    }
    $1 == "require" && $2 == path && $3 != "" {
      print $3
      exit
    }
  ' "$go_mod"
}

module_label() {
  local go_mod="$1"
  dirname "${go_mod#./}"
}

main() {
  if [[ $# -ne 1 ]]; then
    usage
    exit 1
  fi

  local input_version="$1"
  local version
  version="$(normalize_version "$input_version")"

  if ! is_semver "$version"; then
    echo "Invalid version: $input_version" >&2
    usage
    exit 1
  fi

  local go_mods=()
  local go_mod current_version

  while IFS= read -r go_mod; do
    current_version="$(get_required_version "$go_mod")"
    if [[ -n "$current_version" ]]; then
      go_mods+=("$go_mod")
    fi
  done < <(get_go_mods | sort)

  if [[ "${#go_mods[@]}" -eq 0 ]]; then
    echo "No go.mod files require $FERRET_MODULE" >&2
    exit 1
  fi

  echo "Package: $FERRET_MODULE"
  echo "Version: v$version"

  for go_mod in "${go_mods[@]}"; do
    current_version="$(get_required_version "$go_mod")"

    if [[ "$current_version" == "v$version" ]]; then
      echo "Skipped $(module_label "$go_mod"): already uses v$version"
      continue
    fi

    if [[ "$go_mod" == "$DIR_RUNTIME/go.mod" ]]; then
      ( cd "$(dirname "$go_mod")" && \
        go mod edit -require="$FERRET_MODULE@v$version" && \
        go mod tidy
      )
    else
      ( cd "$(dirname "$go_mod")" && \
        go mod edit -require="$FERRET_MODULE@v$version" && \
        GOWORK=off go mod tidy
      )
    fi

    echo "Updated $(module_label "$go_mod"): $current_version -> v$version"
  done

  go work vendor
}

main "$@"
