#!/usr/bin/env bash
set -euo pipefail

DIR_MODULES="./modules"
DIR_PACKAGES="./pkg"

usage() {
  echo "Usage: $0 <package> <semver> [module ...]"
  echo "Examples:"
  echo "  $0 common 0.1.1"
  echo "  $0 common v0.1.1 csv xml"
}

get_packages() {
  find "$DIR_PACKAGES" -mindepth 2 -maxdepth 2 -type f -name go.mod \
    -exec dirname {} \; \
    | sed "s|^$DIR_PACKAGES/||" \
    | sort
}

get_modules() {
  find "$DIR_MODULES" -type f -name go.mod \
    -exec dirname {} \; \
    | sed "s|^$DIR_MODULES/||" \
    | sort
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

normalize_version() {
  local version="$1"
  echo "${version#v}"
}

is_semver() {
  local version="$1"
  [[ "$version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$ ]]
}

get_package_module_path() {
  local package="$1"
  local go_mod="$DIR_PACKAGES/$package/go.mod"

  awk '$1 == "module" { print $2; exit }' "$go_mod"
}

get_required_version() {
  local module="$1"
  local package_path="$2"
  local go_mod="$DIR_MODULES/$module/go.mod"

  awk -v path="$package_path" '
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

main() {
  if [[ $# -lt 2 ]]; then
    usage
    exit 1
  fi

  local package="$1"
  local input_version="$2"
  shift 2

  local version
  version="$(normalize_version "$input_version")"

  if ! is_semver "$version"; then
    echo "Invalid version: $input_version" >&2
    usage
    exit 1
  fi

  if ! package_exists "$package"; then
    echo "Unknown package: $package" >&2
    echo "Available packages:" >&2
    get_packages >&2
    exit 1
  fi

  local package_path
  package_path="$(get_package_module_path "$package")"

  if [[ -z "$package_path" ]]; then
    echo "Package has no module path: $package" >&2
    exit 1
  fi

  local modules=()
  local module current_version

  if [[ $# -eq 0 ]]; then
    while IFS= read -r module; do
      current_version="$(get_required_version "$module" "$package_path")"
      if [[ -n "$current_version" ]]; then
        modules+=("$module")
      fi
    done < <(get_modules)
  else
    for module in "$@"; do
      if ! module_exists "$module"; then
        echo "Unknown module: $module" >&2
        echo "Available modules:" >&2
        get_modules >&2
        exit 1
      fi

      current_version="$(get_required_version "$module" "$package_path")"
      if [[ -z "$current_version" ]]; then
        echo "Module '$module' does not require package '$package' ($package_path)" >&2
        exit 1
      fi

      modules+=("$module")
    done
  fi

  if [[ "${#modules[@]}" -eq 0 ]]; then
    echo "No modules require package '$package' ($package_path)" >&2
    exit 1
  fi

  echo "Package: $package_path"
  echo "Version: v$version"

  for module in "${modules[@]}"; do
    current_version="$(get_required_version "$module" "$package_path")"

    ( cd "$DIR_MODULES/$module" && \
      go mod edit -require="$package_path@v$version" && \
      GOWORK=off go mod tidy
    )

    if [[ "$current_version" == "v$version" ]]; then
      echo "Refreshed module '$module': already uses v$version"
    else
      echo "Updated module '$module': $current_version -> v$version"
    fi
  done
}

main "$@"
