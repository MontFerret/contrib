#!/usr/bin/env bash
set -euo pipefail

DIR_PACKAGES="./pkg"
TAG_PACKAGES="pkg"

usage() {
  echo "Usage: $0 <major|minor|patch> <package>"
  echo "       $0 <package> <semver>"
  echo "Examples:"
  echo "  $0 patch common"
  echo "  $0 common 0.2.0-rc.1"
}

get_packages() {
  find "$DIR_PACKAGES" -mindepth 2 -maxdepth 2 -type f -name go.mod \
    -exec dirname {} \; \
    | sed "s|^$DIR_PACKAGES/||" \
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

get_latest_tag() {
  local package="$1"
  git tag --list "$TAG_PACKAGES/$package/v*" --sort=-version:refname | head -n 1
}

normalize_version() {
  local version="$1"
  echo "${version#v}"
}

is_semver() {
  local version="$1"
  [[ "$version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$ ]]
}

extract_core_version() {
  local version="$1"
  if [[ "$version" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$ ]]; then
    echo "${BASH_REMATCH[1]}.${BASH_REMATCH[2]}.${BASH_REMATCH[3]}"
    return
  fi

  echo "Invalid semantic version: $version" >&2
  exit 1
}

bump_version() {
  local bump="$1"
  local version="$2"
  local core_version

  local major minor patch
  core_version="$(extract_core_version "$version")"
  IFS='.' read -r major minor patch <<< "$core_version"

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

  local mode="$1"
  local target="$2"
  local package new_version

  case "$mode" in
    major|minor|patch)
      package="$target"
      ;;
    *)
      package="$mode"
      new_version="$(normalize_version "$target")"
      if ! is_semver "$new_version"; then
        echo "Invalid version: $target" >&2
        usage
        exit 1
      fi
      ;;
  esac

  if ! package_exists "$package"; then
    echo "Unknown package: $package" >&2
    echo "Available packages:" >&2
    get_packages >&2
    exit 1
  fi

  local latest_tag current_version new_tag
  latest_tag="$(get_latest_tag "$package")"

  if [[ -z "$latest_tag" ]]; then
    current_version="0.0.0"
  else
    current_version="$(normalize_version "${latest_tag##*/}")"
  fi

  case "$mode" in
    major|minor|patch)
      new_version="$(bump_version "$mode" "$current_version")"
      ;;
  esac

  new_tag="$TAG_PACKAGES/$package/v$new_version"

  if git rev-parse -q --verify "refs/tags/$new_tag" >/dev/null; then
    echo "Tag already exists: $new_tag" >&2
    exit 1
  fi

  echo "Package:         $package"
  echo "Current version: v$current_version"
  echo "Next version:    v$new_version"
  echo "Tag:             $new_tag"

  git tag -a "$new_tag" -m "Release $new_tag"
  echo "Created tag: $new_tag"

  git push origin "$new_tag"
  echo "Pushed tag:  $new_tag"
}

main "$@"
