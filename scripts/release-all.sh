#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=release-common.sh
source "$SCRIPT_DIR/release-common.sh"

DEFAULT_RELEASE_PRE_BASE_VERSION="1.0.0"

usage() {
  echo "Usage: make release-pre-all <semver|preid>" >&2
  echo "Examples:" >&2
  echo "  make release-pre-all 1.0.0-rc.1" >&2
  echo "  make release-pre-all rc" >&2
  echo "" >&2
  echo "For modules without an initial release, interactive runs prompt for a base version." >&2
  echo "Non-interactive runs can set the initial base version, for example:" >&2
  echo "  RELEASE_PRE_BASE_VERSION=1.0.0 make release-pre-all rc" >&2
  echo "" >&2
  echo "Direct script usage: $0 <semver|preid>" >&2
}

prompt_base_version() {
  local base_version

  echo "Modules without an initial release tag:" >&2
  printf "  %s\n" "$@" >&2
  echo "" >&2

  read -r -p "Initial base version [${DEFAULT_RELEASE_PRE_BASE_VERSION}]: " base_version
  echo "${base_version:-$DEFAULT_RELEASE_PRE_BASE_VERSION}"
}

main() {
  if [[ $# -ne 1 ]]; then
    usage
    exit 1
  fi

  local version_or_preid="$1"
  local preid=""
  local explicit_semver=0
  local base_version initial_version module current_version new_version i
  local modules=()
  local uninitialized_modules=()

  if release_is_semver "$version_or_preid"; then
    explicit_semver=1
  elif release_is_preid "$version_or_preid"; then
    preid="$version_or_preid"
  else
    echo "Invalid version or prerelease identifier: $version_or_preid" >&2
    usage
    exit 1
  fi

  release_require_repository_state

  while IFS= read -r module; do
    modules+=("$module")
  done < <("$SCRIPT_DIR/modules.sh" list)

  if [[ "${#modules[@]}" -eq 0 ]]; then
    release_fail "No modules found"
  fi

  if [[ "$explicit_semver" -eq 0 ]]; then
    for module in "${modules[@]}"; do
      if [[ "$(release_current_version "$module")" == "0.0.0" ]]; then
        uninitialized_modules+=("$module")
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

      if ! release_is_base_version "$base_version"; then
        release_fail "Invalid RELEASE_PRE_BASE_VERSION: $base_version"
        echo "Expected a base semantic version without prerelease or build metadata, for example: ${DEFAULT_RELEASE_PRE_BASE_VERSION}" >&2
        exit 1
      fi

      initial_version="$base_version-$preid.1"
    fi
  fi

  release_reset_plan
  for module in "${modules[@]}"; do
    current_version="$(release_current_version "$module")"
    if [[ "$explicit_semver" -eq 1 ]]; then
      new_version="$version_or_preid"
    elif [[ "$current_version" == "0.0.0" ]]; then
      new_version="$initial_version"
    else
      new_version="$(release_bump_prerelease_version "$preid" "$current_version" "$module")"
    fi

    release_add_plan_entry "$module" "$new_version"
  done

  for i in "${!RELEASE_MODULES[@]}"; do
    echo "Module:          ${RELEASE_MODULES[$i]}"
    echo "Next version:    v${RELEASE_VERSIONS[$i]}"
    echo "Manifest:        ${RELEASE_MANIFESTS[$i]}"
    echo "Tag:             ${RELEASE_TAGS[$i]}"
    echo ""
  done

  if [[ "${RELEASE_CHECK_ONLY:-}" == "1" ]]; then
    release_check_plan
    echo "Release check passed for all modules; no commit or tag created."
    exit 0
  fi

  release_execute_plan "chore(release): publish module versions"
}

main "$@"
