#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=release-common.sh
source "$SCRIPT_DIR/release-common.sh"

usage() {
  echo "Usage:"
  echo "  make release-major <module>"
  echo "  make release-minor <module>"
  echo "  make release-patch <module>"
  echo "  make release-pre <module> <semver|preid>"
  echo "Examples:"
  echo "  make release-patch xml"
  echo "  make release-pre xml 1.0.0-rc.1"
  echo "  make release-pre xml rc"
  echo ""
  echo "Direct script usage:"
  echo "  $0 <major|minor|patch> <module>"
  echo "  $0 <module> <semver>"
  echo "  $0 <module> <preid>"
}

main() {
  if [[ $# -ne 2 ]]; then
    usage
    exit 1
  fi

  local mode="$1"
  local target="$2"
  local module request preid current_version new_version new_tag

  case "$mode" in
    major|minor|patch)
      module="$target"
      request="$mode"
      ;;
    *)
      module="$mode"
      request="$target"
      if release_is_preid "$request"; then
        preid="$request"
      else
        new_version="$(release_normalize_version "$request")"
        if ! release_is_semver "$new_version"; then
          echo "Invalid version or prerelease identifier: $request" >&2
          usage
          exit 1
        fi
      fi
      ;;
  esac

  if [[ ! -d "$DIR_MODULES/$module" ]]; then
    release_fail "Unknown module: $module"
  fi

  release_require_repository_state
  current_version="$(release_current_version "$module")"

  case "$request" in
    major|minor|patch)
      new_version="$(release_bump_version "$request" "$current_version")"
      ;;
    *)
      if [[ -n "${preid:-}" ]]; then
        new_version="$(release_bump_prerelease_version "$preid" "$current_version" "$module")"
      fi
      ;;
  esac

  release_reset_plan
  release_add_plan_entry "$module" "$new_version"
  new_tag="${RELEASE_TAGS[0]}"

  echo "Module:          $module"
  echo "Current version: v$current_version"
  echo "Next version:    v$new_version"
  echo "Manifest:        ${RELEASE_MANIFESTS[0]}"
  echo "Tag:             $new_tag"

  if [[ "${RELEASE_CHECK_ONLY:-}" == "1" ]]; then
    release_check_plan
    echo "Release check passed; no commit or tag created."
    exit 0
  fi

  release_execute_plan "chore(release): publish $module v$new_version"
}

main "$@"
