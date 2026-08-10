#!/usr/bin/env bash

DIR_MODULES="modules"
TAG_MODULES="modules"
RELEASE_REMOTE="origin"
RELEASE_BRANCH="main"
RELEASE_REMOTE_TAG_ROOT="refs/release-remote-tags"

release_fail() {
  echo "$*" >&2
  return 1
}

release_normalize_version() {
  local version="$1"
  echo "${version#v}"
}

release_is_semver() {
  local version="$1"
  [[ "$version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$ ]]
}

release_is_preid() {
  local preid="$1"
  [[ "$preid" =~ ^[A-Za-z][0-9A-Za-z-]*$ ]]
}

release_is_base_version() {
  local version="$1"
  [[ "$version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]
}

release_extract_core_version() {
  local version="$1"
  if [[ "$version" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$ ]]; then
    echo "${BASH_REMATCH[1]}.${BASH_REMATCH[2]}.${BASH_REMATCH[3]}"
    return
  fi

  release_fail "Invalid semantic version: $version"
}

release_bump_version() {
  local bump="$1"
  local version="$2"
  local core_version
  local major minor patch

  core_version="$(release_extract_core_version "$version")"
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
      release_fail "Invalid bump type: $bump"
      ;;
  esac

  echo "$major.$minor.$patch"
}

release_bump_prerelease_version() {
  local preid="$1"
  local version="$2"
  local module="$3"

  if [[ "$version" =~ ^([0-9]+\.[0-9]+\.[0-9]+)-$preid\.([0-9]+)$ ]]; then
    echo "${BASH_REMATCH[1]}-$preid.$((10#${BASH_REMATCH[2]} + 1))"
    return
  fi

  release_fail "Latest version for module '$module' is not a matching prerelease: v$version"
  echo "Use an explicit semantic version first, for example: make release-pre $module 1.0.0-$preid.1" >&2
  return 1
}

release_require_repository_state() {
  local repository_root current_branch upstream head remote_head

  repository_root="$(git rev-parse --show-toplevel 2>/dev/null)" ||
    release_fail "Module releases must run from a Git repository."
  if [[ "$(pwd -P)" != "$(cd "$repository_root" && pwd -P)" ]]; then
    release_fail "Module releases must run from the repository root: $repository_root"
  fi

  if [[ -n "$(git status --porcelain --untracked-files=all)" ]]; then
    release_fail "Module releases require a clean working tree."
  fi

  current_branch="$(git symbolic-ref --quiet --short HEAD 2>/dev/null)" ||
    release_fail "Module releases cannot run from a detached HEAD."
  if [[ "$current_branch" != "$RELEASE_BRANCH" ]]; then
    release_fail "Module releases must run from '$RELEASE_BRANCH', not '$current_branch'."
  fi

  git fetch --no-tags --prune "$RELEASE_REMOTE" \
    "+refs/heads/$RELEASE_BRANCH:refs/remotes/$RELEASE_REMOTE/$RELEASE_BRANCH" \
    "+refs/tags/$TAG_MODULES/*:$RELEASE_REMOTE_TAG_ROOT/$TAG_MODULES/*"

  upstream="$(git rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' 2>/dev/null || true)"
  if [[ "$upstream" != "$RELEASE_REMOTE/$RELEASE_BRANCH" ]]; then
    release_fail "Branch '$RELEASE_BRANCH' must track '$RELEASE_REMOTE/$RELEASE_BRANCH'."
  fi

  head="$(git rev-parse HEAD)"
  remote_head="$(git rev-parse "refs/remotes/$RELEASE_REMOTE/$RELEASE_BRANCH")"
  if [[ "$head" != "$remote_head" ]]; then
    release_fail "Branch '$RELEASE_BRANCH' must be synchronized with '$RELEASE_REMOTE/$RELEASE_BRANCH'."
  fi
}

release_latest_tag() {
  local module="$1"
  local ref

  ref="$(
    git for-each-ref \
      --count=1 \
      --sort=-version:refname \
      --format='%(refname)' \
      "$RELEASE_REMOTE_TAG_ROOT/$TAG_MODULES/$module/v*"
  )"
  if [[ -n "$ref" ]]; then
    echo "${ref#"$RELEASE_REMOTE_TAG_ROOT/"}"
  fi
}

release_current_version() {
  local module="$1"
  local latest_tag

  latest_tag="$(release_latest_tag "$module")"
  if [[ -z "$latest_tag" ]]; then
    echo "0.0.0"
  else
    release_normalize_version "${latest_tag##*/}"
  fi
}

release_assert_tag_available() {
  local tag="$1"

  if git show-ref --verify --quiet "refs/tags/$tag"; then
    release_fail "Tag already exists locally: $tag"
  fi
  if git show-ref --verify --quiet "$RELEASE_REMOTE_TAG_ROOT/$tag"; then
    release_fail "Tag already exists on $RELEASE_REMOTE: $tag"
  fi
}

release_reset_plan() {
  RELEASE_MODULES=()
  RELEASE_VERSIONS=()
  RELEASE_MANIFESTS=()
  RELEASE_TAGS=()
}

release_add_plan_entry() {
  local module="$1"
  local version="$2"
  local manifest tag

  if [[ ! "$module" =~ ^[a-z0-9][a-z0-9._-]*(/[a-z0-9][a-z0-9._-]*)*$ ]]; then
    release_fail "Invalid module path: $module"
  fi
  if ! release_is_semver "$version"; then
    release_fail "Invalid semantic version: $version"
  fi

  manifest="$DIR_MODULES/$module/ferret.yaml"
  if [[ ! -f "$manifest" ]]; then
    release_fail "Missing module manifest: $manifest"
  fi
  if ! git ls-files --error-unmatch "$manifest" >/dev/null 2>&1; then
    release_fail "Module manifest is not tracked: $manifest"
  fi

  tag="$TAG_MODULES/$module/v$version"
  git check-ref-format "refs/tags/$tag" >/dev/null ||
    release_fail "Invalid release tag: $tag"
  release_assert_tag_available "$tag"

  RELEASE_MODULES+=("$module")
  RELEASE_VERSIONS+=("$version")
  RELEASE_MANIFESTS+=("$manifest")
  RELEASE_TAGS+=("$tag")
}

release_write_manifest_version() {
  local manifest="$1"
  local version="$2"
  local temporary

  temporary="$(mktemp "${TMPDIR:-/tmp}/contrib-release-manifest.XXXXXX")"
  if ! awk -v version="$version" '
    BEGIN { count = 0 }
    /^version:[[:space:]]*/ {
      count++
      print "version: " version
      next
    }
    { print }
    END {
      if (count != 1) {
        exit 42
      }
    }
  ' "$manifest" > "$temporary"; then
    rm -f "$temporary"
    release_fail "Manifest must contain exactly one top-level version field: $manifest"
    return 1
  fi

  cp "$temporary" "$manifest"
  rm -f "$temporary"
}

release_read_manifest_version() {
  local manifest="$1"

  awk '
    /^version:[[:space:]]*/ {
      count++
      value = $0
      sub(/^version:[[:space:]]*/, "", value)
    }
    END {
      if (count != 1) {
        exit 42
      }
      print value
    }
  ' "$manifest"
}

release_validate_manifests() {
  local validator

  validator="${FERRET_SPEC:-ferret-spec}"
  if ! command -v "$validator" >/dev/null 2>&1; then
    release_fail "Manifest validator not found: $validator. Run 'make install-manifest-validator'."
  fi

  "$validator" validate module "$@"
}

release_verify_planned_versions() {
  local i actual

  for i in "${!RELEASE_MANIFESTS[@]}"; do
    actual="$(release_read_manifest_version "${RELEASE_MANIFESTS[$i]}")" || {
      release_fail "Manifest must contain exactly one top-level version field: ${RELEASE_MANIFESTS[$i]}"
      return 1
    }
    if [[ "$actual" != "${RELEASE_VERSIONS[$i]}" ]]; then
      release_fail "Manifest version '$actual' does not match planned version '${RELEASE_VERSIONS[$i]}': ${RELEASE_MANIFESTS[$i]}"
    fi
  done
}

release_path_is_planned() {
  local candidate="$1"
  local manifest

  for manifest in "${RELEASE_MANIFESTS[@]}"; do
    if [[ "$candidate" == "$manifest" ]]; then
      return 0
    fi
  done

  return 1
}

release_assert_only_planned_changes() {
  local line changed_path

  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    changed_path="${line:3}"
    if ! release_path_is_planned "$changed_path"; then
      release_fail "Release preparation changed an unexpected path: $changed_path"
      return 1
    fi
  done < <(git status --porcelain --untracked-files=all)
}

release_assert_only_planned_staged_paths() {
  local staged_path

  while IFS= read -r staged_path; do
    [[ -z "$staged_path" ]] && continue
    if ! release_path_is_planned "$staged_path"; then
      release_fail "Release preparation staged an unexpected path: $staged_path"
      return 1
    fi
  done < <(git diff --cached --name-only)

  if ! git diff --quiet; then
    release_fail "Release preparation left unstaged tracked changes."
  fi
  if [[ -n "$(git ls-files --others --exclude-standard)" ]]; then
    release_fail "Release preparation created unexpected untracked files."
  fi
}

release_check_plan() (
  set -euo pipefail

  local temporary_root copy i
  local copies=()

  temporary_root="$(mktemp -d "${TMPDIR:-/tmp}/contrib-release-check.XXXXXX")"
  trap 'rm -rf "$temporary_root"' EXIT

  for i in "${!RELEASE_MANIFESTS[@]}"; do
    copy="$temporary_root/$i.yaml"
    cp "${RELEASE_MANIFESTS[$i]}" "$copy"
    release_write_manifest_version "$copy" "${RELEASE_VERSIONS[$i]}"
    copies+=("$copy")
  done

  release_validate_manifests "${copies[@]}"
)

RELEASE_TRANSACTION_ACTIVE=0
RELEASE_START_COMMIT=""
RELEASE_CREATED_TAGS=()

release_rollback_transaction() {
  local status="$1"
  local current_head tag

  if [[ "$RELEASE_TRANSACTION_ACTIVE" != "1" || "$status" -eq 0 ]]; then
    return
  fi

  set +e
  if [[ "${#RELEASE_CREATED_TAGS[@]}" -gt 0 ]]; then
    for tag in "${RELEASE_CREATED_TAGS[@]}"; do
      git tag -d "$tag" >/dev/null 2>&1
    done
  fi

  current_head="$(git rev-parse HEAD 2>/dev/null)"
  if [[ -n "$RELEASE_START_COMMIT" && "$current_head" != "$RELEASE_START_COMMIT" ]]; then
    git reset --hard "$RELEASE_START_COMMIT" >/dev/null 2>&1
  elif [[ "${#RELEASE_MANIFESTS[@]}" -gt 0 ]]; then
    git restore --staged --worktree -- "${RELEASE_MANIFESTS[@]}" >/dev/null 2>&1
  fi

  echo "Release transaction failed; restored its local commit, tags, and manifest edits." >&2
}

release_remote_transaction_landed() {
  local commit="$1"
  local remote_branch tag remote_commit

  remote_branch="$(
    git ls-remote "$RELEASE_REMOTE" "refs/heads/$RELEASE_BRANCH" 2>/dev/null |
      awk 'NR == 1 { print $1 }'
  )"
  if [[ "$remote_branch" != "$commit" ]]; then
    return 1
  fi

  for tag in "${RELEASE_TAGS[@]}"; do
    remote_commit="$(
      git ls-remote "$RELEASE_REMOTE" "refs/tags/$tag^{}" 2>/dev/null |
        awk 'NR == 1 { print $1 }'
    )"
    if [[ "$remote_commit" != "$commit" ]]; then
      return 1
    fi
  done

  return 0
}

release_execute_plan() {
  local commit_message="$1"
  local i tag release_commit
  local push_args=(--atomic "$RELEASE_REMOTE" "HEAD:refs/heads/$RELEASE_BRANCH")

  if [[ -z "$(git config user.name)" || -z "$(git config user.email)" ]]; then
    release_fail "Git user.name and user.email must be configured for module releases."
  fi

  RELEASE_TRANSACTION_ACTIVE=1
  RELEASE_START_COMMIT="$(git rev-parse HEAD)"
  RELEASE_CREATED_TAGS=()
  trap 'release_rollback_transaction $?' EXIT

  for i in "${!RELEASE_MANIFESTS[@]}"; do
    release_write_manifest_version "${RELEASE_MANIFESTS[$i]}" "${RELEASE_VERSIONS[$i]}"
  done

  release_validate_manifests "${RELEASE_MANIFESTS[@]}"
  release_verify_planned_versions
  release_assert_only_planned_changes

  git add -- "${RELEASE_MANIFESTS[@]}"
  release_assert_only_planned_staged_paths

  if ! git diff --cached --quiet; then
    git commit -m "$commit_message"
    echo "Created release commit: $(git rev-parse --short HEAD)"
  else
    echo "Manifests already contain the planned versions; no release commit was needed."
  fi

  for tag in "${RELEASE_TAGS[@]}"; do
    git tag -a "$tag" -m "Release $tag"
    RELEASE_CREATED_TAGS+=("$tag")
    push_args+=("refs/tags/$tag:refs/tags/$tag")
    echo "Created tag: $tag"
  done

  release_commit="$(git rev-parse HEAD)"
  if ! git push "${push_args[@]}"; then
    if release_remote_transaction_landed "$release_commit"; then
      echo "Atomic push completed despite a local transport error; confirmed remote refs."
    else
      release_fail "Atomic release push failed."
      return 1
    fi
  fi

  RELEASE_TRANSACTION_ACTIVE=0
  trap - EXIT
  echo "Atomically pushed $RELEASE_BRANCH and ${#RELEASE_TAGS[@]} module tag(s)."
}
