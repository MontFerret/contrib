#!/usr/bin/env bash
set -euo pipefail

DIR_MODULES="./modules"

usage() {
  echo "Usage: $0 <major|minor|patch> <module>"
}

get_latest_tag() {
  local module="$1"
  git tag --list "$DIR_MODULES/$module/v*" --sort=-version:refname | head -n 1
}

bump_version() {
  local bump="$1"
  local version="$2"

  local major minor patch
  IFS='.' read -r major minor patch <<< "$version"

  case "$bump" in
    major)
      ((major += 1))
      minor=0
      patch=0
      ;;
      minor)
      ((minor += 1))
      patch=0
      ;;
    patch)
      ((patch += 1))
      ;;
    *)
      echo "Invalid bump type: $bump" >&2
      exit 1
      ;;
  esac

  echo "$major.$minor.$patch"
}

main() {
  if [[ $# -ne 2 ]]; then
    usage
    exit 1
  fi

  local bump="$1"
  local module="$2"

  if [[ ! -d "$DIR_MODULES/$module" ]]; then
    echo "Unknown module: $module" >&2
    exit 1
  fi

  local latest_tag current_version new_version new_tag
  latest_tag="$(get_latest_tag "$module")"

  if [[ -z "$latest_tag" ]]; then
    current_version="0.0.0"
  else
    current_version="${latest_tag##*/v}"
  fi

  new_version="$(bump_version "$bump" "$current_version")"
  new_tag="$DIR_MODULES/$module/v$new_version"

  echo "Module:          $module"
  echo "Current version: v$current_version"
  echo "Next version:    v$new_version"
  echo "Tag:             $new_tag"

  git tag -a "$new_tag" -m "Release $new_tag"

  echo "Created tag: $new_tag"
  echo "Push with:"
  echo "  git push origin $new_tag"
}

main "$@"