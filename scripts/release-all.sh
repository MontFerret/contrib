#!/usr/bin/env bash
set -euo pipefail

TAG_MODULES="modules"
DEFAULT_RELEASE_PRE_BASE_VERSION="1.0.0"

usage() {
  echo "Usage: make release-pre-all <semver|preid>" >&2
  echo "Examples:" >&2
  echo "  make release-pre-all 1.0.0-rc.1" >&2
  echo "  make release-pre-all rc" >&2
  echo "" >&2
  echo "For modules without an initial release, interactive runs prompt for a base version." >&2
  echo "Non-interactive runs can set RELEASE_PRE_BASE_VERSION, for example:" >&2
  echo "  RELEASE_PRE_BASE_VERSION=1.0.0 make release-pre-all rc" >&2
  echo "" >&2
  echo "Direct script usage: $0 <semver|preid>" >&2
}

is_semver() {
  local version="$1"
  [[ "$version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$ ]]
}

is_base_version() {
  local version="$1"
  [[ "$version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]
}

is_preid() {
  local preid="$1"
  [[ "$preid" =~ ^[A-Za-z][0-9A-Za-z-]*$ ]]
}

get_latest_tag() {
  local module="$1"
  git tag --list "$TAG_MODULES/$module/v*" --sort=-version:refname | head -n 1
}

prompt_base_version() {
  local base_version

  echo "Modules without an initial release tag:" >&2
  printf "  %s\n" "$@" >&2
  echo "" >&2

  read -r -p "Initial base version [${DEFAULT_RELEASE_PRE_BASE_VERSION}]: " base_version
  echo "${base_version:-$DEFAULT_RELEASE_PRE_BASE_VERSION}"
}

if [ "$#" -ne 1 ]; then
  usage
  exit 1
fi

VERSION_OR_PREID="$1"
preid=""
explicit_semver=0

if is_semver "$VERSION_OR_PREID"; then
  explicit_semver=1
elif is_preid "$VERSION_OR_PREID"; then
  preid="$VERSION_OR_PREID"
else
  echo "Invalid version or prerelease identifier: $VERSION_OR_PREID" >&2
  usage
  exit 1
fi

modules=()
while IFS= read -r module; do
  modules+=("$module")
done < <(make modules)

if [ "${#modules[@]}" -eq 0 ]; then
  echo "No modules found" >&2
  exit 1
fi

resolved_args=()
uninitialized_modules=()

if [[ "$explicit_semver" -eq 1 ]]; then
  for _ in "${modules[@]}"; do
    resolved_args+=("$VERSION_OR_PREID")
  done
else
  for module in "${modules[@]}"; do
    if [[ -z "$(get_latest_tag "$module")" ]]; then
      resolved_args+=("")
      uninitialized_modules+=("$module")
    else
      resolved_args+=("$VERSION_OR_PREID")
    fi
  done

  if [[ "${#uninitialized_modules[@]}" -gt 0 ]]; then
    base_version="${RELEASE_PRE_BASE_VERSION:-}"

    if [[ -z "$base_version" ]]; then
      if [[ ! -t 0 || ! -t 1 ]]; then
        echo "Modules without an initial release tag require an initial base version:" >&2
        printf "  %s\n" "${uninitialized_modules[@]}" >&2
        echo "Use RELEASE_PRE_BASE_VERSION, for example: RELEASE_PRE_BASE_VERSION=${DEFAULT_RELEASE_PRE_BASE_VERSION} make release-pre-all $preid" >&2
        exit 1
      fi

      base_version="$(prompt_base_version "${uninitialized_modules[@]}")"
    fi

    if ! is_base_version "$base_version"; then
      echo "Invalid RELEASE_PRE_BASE_VERSION: $base_version" >&2
      echo "Expected a base semantic version without prerelease or build metadata, for example: ${DEFAULT_RELEASE_PRE_BASE_VERSION}" >&2
      exit 1
    fi

    initial_version="$base_version-$preid.1"

    for i in "${!modules[@]}"; do
      if [[ -z "${resolved_args[$i]}" ]]; then
        resolved_args[$i]="$initial_version"
      fi
    done
  fi
fi

for i in "${!modules[@]}"; do
  module="${modules[$i]}"
  resolved_arg="${resolved_args[$i]}"

  echo "Checking module '$module' with '$resolved_arg'"
  RELEASE_CHECK_ONLY=1 ./scripts/release.sh "$module" "$resolved_arg"
done

if [[ "${RELEASE_CHECK_ONLY:-}" == "1" ]]; then
  echo "Release check passed for all modules; no tags created."
  exit 0
fi

for i in "${!modules[@]}"; do
  module="${modules[$i]}"
  resolved_arg="${resolved_args[$i]}"

  echo "Releasing module '$module' with '$resolved_arg'"
  make release-pre "$module" "$resolved_arg"
done
