#!/usr/bin/env bash
set -euo pipefail

# The caller supplies a checkout with complete history and tags. Discovery is
# read-only so the same selection can be used for automatic runs and backfills.
fail() {
  echo "Discover publications: $*" >&2
  return 1
}

is_commit_id() {
  [[ "$1" =~ ^[0-9a-f]{40}$ ]]
}

is_module_tag() {
  local segment='[a-z0-9]([a-z0-9._-]*[a-z0-9])?'
  local number='(0|[1-9][0-9]*)'
  local prerelease='(0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*)'
  local build='[0-9A-Za-z-]+'
  local version="$number\\.$number\\.$number(-$prerelease(\\.$prerelease)*)?(\\+$build(\\.$build)*)?"

  [[ "$1" =~ ^modules/($segment/)+v$version$ ]]
}

resolve_commit() {
  git rev-parse --verify --end-of-options "$1^{commit}" 2>/dev/null ||
    fail "reference does not resolve to a commit: $1"
}

emit_record() {
  jq -cn --arg tag "$1" --arg commit "$2" '{tag: $tag, commit: $commit}'
}

select_tags() {
  local commits="$1"
  local required="$2"
  local tags tag commit found=0

  tags="$(git for-each-ref --format='%(refname:strip=2)' refs/tags/modules/)"
  while IFS= read -r tag; do
    if ! is_module_tag "$tag"; then
      continue
    fi

    # Tags targeting trees or blobs are not module releases. An explicit
    # request for one still fails through resolve_commit in select_ref.
    if ! commit="$(git rev-parse --verify --end-of-options "refs/tags/$tag^{commit}" 2>/dev/null)"; then
      continue
    fi
    case $'\n'"$commits"$'\n' in
      *$'\n'"$commit"$'\n'*)
        emit_record "$tag" "$commit"
        found=1
        ;;
    esac
  done <<< "$tags"

  if [[ "$required" == 1 && "$found" == 0 ]]; then
    fail "no module release tags point to the requested commit"
  fi
}

select_ref() {
  local ref="$1"
  local commit

  if is_module_tag "$ref"; then
    commit="$(resolve_commit "refs/tags/$ref")"
    emit_record "$ref" "$commit"
  elif is_commit_id "$ref"; then
    commit="$(resolve_commit "$ref")"
    select_tags "$commit" 1
  else
    fail "expected a modules/<module>/v<version> tag or a full commit SHA: $ref"
  fi
}

select_range() {
  local before="$1"
  local after="$2"
  local commits
  local zero='0000000000000000000000000000000000000000'

  if ! is_commit_id "$before" || ! is_commit_id "$after"; then
    fail "push boundaries must be full commit SHAs"
    return 1
  fi

  if [[ "$after" == "$zero" ]]; then
    return
  fi

  after="$(resolve_commit "$after")"
  if [[ "$before" == "$zero" ]]; then
    commits="$(git rev-list "$after")"
  else
    before="$(resolve_commit "$before")"
    commits="$(git rev-list "$before..$after")"
  fi

  select_tags "$commits" 0
}

main() {
  case "${1:-}" in
    ref)
      [[ $# -eq 2 ]] || { fail "usage: $0 ref <tag|commit>"; return 1; }
      select_ref "$2"
      ;;
    range)
      [[ $# -eq 3 ]] || { fail "usage: $0 range <before> <after>"; return 1; }
      select_range "$2" "$3"
      ;;
    *)
      fail "usage: $0 ref <tag|commit> | range <before> <after>"
      ;;
  esac
}

main "$@" | jq -sc '{include: (unique_by(.tag) | sort_by(.tag))}'
