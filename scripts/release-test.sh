#!/usr/bin/env bash
set -euo pipefail

REPOSITORY_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/contrib-release-tests.XXXXXX")"
FIXTURE_ROOT=""
FIXTURE_WORK=""
FIXTURE_ORIGIN=""
FIXTURE_VALIDATOR=""

cleanup() {
  if [[ -n "$TEST_ROOT" && -d "$TEST_ROOT" ]]; then
    rm -rf "$TEST_ROOT"
  fi
}
trap cleanup EXIT

fail() {
  echo "FAIL: $*" >&2
  if [[ -n "$FIXTURE_ROOT" && -f "$FIXTURE_ROOT/release.log" ]]; then
    echo "Release output:" >&2
    sed -n '1,240p' "$FIXTURE_ROOT/release.log" >&2
  fi
  exit 1
}

assert_equal() {
  local expected="$1"
  local actual="$2"
  local message="$3"

  if [[ "$actual" != "$expected" ]]; then
    fail "$message: got '$actual', want '$expected'"
  fi
}

assert_not_equal() {
  local unexpected="$1"
  local actual="$2"
  local message="$3"

  if [[ "$actual" == "$unexpected" ]]; then
    fail "$message: unexpectedly got '$actual'"
  fi
}

assert_clean() {
  local status

  status="$(git -C "$FIXTURE_WORK" status --porcelain --untracked-files=all)"
  assert_equal "" "$status" "fixture working tree is not clean"
}

new_fixture() {
  local name="$1"

  FIXTURE_ROOT="$TEST_ROOT/$name"
  FIXTURE_WORK="$FIXTURE_ROOT/work"
  FIXTURE_ORIGIN="$FIXTURE_ROOT/origin.git"
  FIXTURE_VALIDATOR="$FIXTURE_ROOT/bin/ferret-spec"

  mkdir -p "$FIXTURE_ROOT/bin"
  git init --bare --initial-branch=main "$FIXTURE_ORIGIN" >/dev/null
  git clone "$FIXTURE_ORIGIN" "$FIXTURE_WORK" >/dev/null 2>&1
  git -C "$FIXTURE_WORK" config user.name "Release Test"
  git -C "$FIXTURE_WORK" config user.email "release-test@example.com"

  mkdir -p "$FIXTURE_WORK/scripts"
  cp "$REPOSITORY_ROOT/scripts/release-common.sh" "$FIXTURE_WORK/scripts/release-common.sh"
  cp "$REPOSITORY_ROOT/scripts/release.sh" "$FIXTURE_WORK/scripts/release.sh"
  cp "$REPOSITORY_ROOT/scripts/release-all.sh" "$FIXTURE_WORK/scripts/release-all.sh"
  cp "$REPOSITORY_ROOT/scripts/modules.sh" "$FIXTURE_WORK/scripts/modules.sh"
  chmod +x \
    "$FIXTURE_WORK/scripts/release.sh" \
    "$FIXTURE_WORK/scripts/release-all.sh" \
    "$FIXTURE_WORK/scripts/modules.sh"

  printf '# release fixture\n' > "$FIXTURE_WORK/README.md"
  write_fake_validator
}

write_fake_validator() {
  printf '%s\n' \
    '#!/usr/bin/env bash' \
    'set -euo pipefail' \
    'if [[ "${FAKE_VALIDATOR_FAIL:-}" == "1" ]]; then' \
    '  echo "forced validator failure" >&2' \
    '  exit 9' \
    'fi' \
    '[[ "$1" == "validate" && "$2" == "module" ]]' \
    'shift 2' \
    'for manifest in "$@"; do' \
    '  count="$(grep -c "^version:[[:space:]]" "$manifest" || true)"' \
    '  [[ "$count" == "1" ]] || exit 10' \
    '  version="$(awk "/^version:[[:space:]]/ { print \$2 }" "$manifest")"' \
    '  [[ "$version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$ ]] || exit 11' \
    'done' > "$FIXTURE_VALIDATOR"
  chmod +x "$FIXTURE_VALIDATOR"
}

add_module() {
  local module="$1"
  local version="$2"
  local directory="$FIXTURE_WORK/modules/$module"
  local name="${module##*/}"

  mkdir -p "$directory"
  printf 'module example.test/modules/%s\n\ngo 1.25.0\n' "$module" > "$directory/go.mod"
  printf '%s\n' \
    '$schema: https://schemas.ferretlang.org/module/v1.json' \
    "name: montferret/$name" \
    'namespace: TEST' \
    "version: $version" \
    'description: Release fixture.' \
    'license: Apache-2.0' \
    'authors:' \
    '  - name: MontFerret Authors' \
    'repository:' \
    '  url: https://github.com/MontFerret/contrib' \
    "  directory: modules/$module" \
    'compatibility:' \
    '  ferret: ">=2.0.0-alpha.43 <3.0.0"' \
    'exports:' \
    '  namespaces:' \
    '    - name: TEST' \
    '      functions:' \
    '        - OPEN' > "$directory/ferret.yaml"
}

commit_fixture() {
  git -C "$FIXTURE_WORK" add .
  git -C "$FIXTURE_WORK" commit -m "initial fixture" >/dev/null
  git -C "$FIXTURE_WORK" push --set-upstream origin main >/dev/null 2>&1
}

tag_fixture() {
  local module="$1"
  local version="$2"
  local tag="modules/$module/v$version"

  git -C "$FIXTURE_WORK" tag -a "$tag" -m "Release $tag"
  git -C "$FIXTURE_WORK" push origin "refs/tags/$tag" >/dev/null 2>&1
}

run_single_release() {
  (
    cd "$FIXTURE_WORK"
    FERRET_SPEC="$FIXTURE_VALIDATOR" ./scripts/release.sh "$@"
  ) > "$FIXTURE_ROOT/release.log" 2>&1
}

run_bulk_release() {
  (
    cd "$FIXTURE_WORK"
    FERRET_SPEC="$FIXTURE_VALIDATOR" ./scripts/release-all.sh "$@"
  ) > "$FIXTURE_ROOT/release.log" 2>&1
}

remote_main() {
  git --git-dir "$FIXTURE_ORIGIN" rev-parse refs/heads/main
}

remote_tag_commit() {
  local tag="$1"
  git --git-dir "$FIXTURE_ORIGIN" rev-parse "refs/tags/$tag^{}"
}

remote_manifest_version() {
  local ref="$1"
  local module="$2"

  git --git-dir "$FIXTURE_ORIGIN" show "$ref:modules/$module/ferret.yaml" |
    awk '/^version:[[:space:]]/ { print $2 }'
}

test_initial_and_consequent_release() {
  local initial release_commit

  new_fixture initial-and-consequent
  add_module db/redis 1.0.0-rc.1
  commit_fixture
  initial="$(remote_main)"

  run_single_release db/redis 1.0.0-rc.1 || fail "initial release failed"
  assert_equal "$initial" "$(remote_main)" "initial release unexpectedly created a commit"
  assert_equal "$initial" "$(remote_tag_commit modules/db/redis/v1.0.0-rc.1)" "initial tag commit"

  run_single_release db/redis rc || fail "consequent release failed"
  release_commit="$(remote_main)"
  assert_not_equal "$initial" "$release_commit" "consequent release did not create a commit"
  assert_equal "$initial" "$(git --git-dir "$FIXTURE_ORIGIN" rev-parse "$release_commit^")" "consequent release commit parent"
  assert_equal "$release_commit" "$(remote_tag_commit modules/db/redis/v1.0.0-rc.2)" "consequent tag commit"
  assert_equal "1.0.0-rc.2" "$(remote_manifest_version "$release_commit" db/redis)" "consequent manifest version"
  assert_clean
}

test_stale_manifest_recovery() {
  local initial rc2_commit release_commit

  new_fixture stale-manifest
  add_module db/redis 1.0.0-rc.1
  commit_fixture
  initial="$(remote_main)"
  tag_fixture db/redis 1.0.0-rc.1
  tag_fixture db/redis 1.0.0-rc.2
  rc2_commit="$(remote_tag_commit modules/db/redis/v1.0.0-rc.2)"

  run_single_release db/redis rc || fail "stale-manifest recovery failed"
  release_commit="$(remote_main)"
  assert_equal "$initial" "$rc2_commit" "failed rc.2 tag moved before recovery"
  assert_equal "$rc2_commit" "$(remote_tag_commit modules/db/redis/v1.0.0-rc.2)" "failed rc.2 tag was rewritten"
  assert_equal "$release_commit" "$(remote_tag_commit modules/db/redis/v1.0.0-rc.3)" "recovery tag commit"
  assert_equal "1.0.0-rc.3" "$(remote_manifest_version "$release_commit" db/redis)" "recovery manifest version"
  assert_clean
}

test_bulk_release_uses_one_commit() {
  local initial release_commit changed

  new_fixture bulk
  add_module xml 1.0.0-rc.1
  add_module yaml 1.0.0-rc.1
  commit_fixture
  tag_fixture xml 1.0.0-rc.1
  tag_fixture yaml 1.0.0-rc.1
  initial="$(remote_main)"

  run_bulk_release rc || fail "bulk release failed"
  release_commit="$(remote_main)"
  assert_equal "1" "$(git --git-dir "$FIXTURE_ORIGIN" rev-list --count "$initial..$release_commit")" "bulk release commit count"
  assert_equal "$release_commit" "$(remote_tag_commit modules/xml/v1.0.0-rc.2)" "bulk XML tag commit"
  assert_equal "$release_commit" "$(remote_tag_commit modules/yaml/v1.0.0-rc.2)" "bulk YAML tag commit"
  assert_equal "1.0.0-rc.2" "$(remote_manifest_version "$release_commit" xml)" "bulk XML manifest version"
  assert_equal "1.0.0-rc.2" "$(remote_manifest_version "$release_commit" yaml)" "bulk YAML manifest version"
  changed="$(git --git-dir "$FIXTURE_ORIGIN" diff-tree --no-commit-id --name-only -r "$release_commit" | sort)"
  assert_equal $'modules/xml/ferret.yaml\nmodules/yaml/ferret.yaml' "$changed" "bulk release changed paths"
  assert_clean
}

test_check_only_is_non_mutating() {
  local initial

  new_fixture check-only
  add_module xml 1.0.0-rc.1
  commit_fixture
  tag_fixture xml 1.0.0-rc.1
  initial="$(remote_main)"

  if ! (
    cd "$FIXTURE_WORK"
    RELEASE_CHECK_ONLY=1 FERRET_SPEC="$FIXTURE_VALIDATOR" ./scripts/release.sh xml rc
  ) > "$FIXTURE_ROOT/release.log" 2>&1; then
    fail "check-only release failed"
  fi

  assert_equal "$initial" "$(remote_main)" "check-only remote main"
  assert_equal "$initial" "$(git -C "$FIXTURE_WORK" rev-parse HEAD)" "check-only local main"
  if git --git-dir "$FIXTURE_ORIGIN" show-ref --verify --quiet refs/tags/modules/xml/v1.0.0-rc.2; then
    fail "check-only created a remote tag"
  fi
  if git -C "$FIXTURE_WORK" show-ref --verify --quiet refs/tags/modules/xml/v1.0.0-rc.2; then
    fail "check-only created a local tag"
  fi
  assert_clean
}

test_repository_state_rejections() {
  local original ahead

  new_fixture dirty
  add_module xml 1.0.0-rc.1
  commit_fixture
  tag_fixture xml 1.0.0-rc.1
  printf 'dirty\n' >> "$FIXTURE_WORK/README.md"
  if run_single_release xml rc; then
    fail "dirty working tree was accepted"
  fi
  [[ -n "$(git -C "$FIXTURE_WORK" status --porcelain)" ]] || fail "dirty user change was removed"

  new_fixture detached
  add_module xml 1.0.0-rc.1
  commit_fixture
  tag_fixture xml 1.0.0-rc.1
  git -C "$FIXTURE_WORK" checkout --detach >/dev/null 2>&1
  if run_single_release xml rc; then
    fail "detached HEAD was accepted"
  fi

  new_fixture ahead
  add_module xml 1.0.0-rc.1
  commit_fixture
  tag_fixture xml 1.0.0-rc.1
  printf 'ahead\n' >> "$FIXTURE_WORK/README.md"
  git -C "$FIXTURE_WORK" add README.md
  git -C "$FIXTURE_WORK" commit -m ahead >/dev/null
  ahead="$(git -C "$FIXTURE_WORK" rev-parse HEAD)"
  if run_single_release xml rc; then
    fail "branch ahead of origin was accepted"
  fi
  assert_equal "$ahead" "$(git -C "$FIXTURE_WORK" rev-parse HEAD)" "ahead commit was disturbed"

  new_fixture behind
  add_module xml 1.0.0-rc.1
  commit_fixture
  tag_fixture xml 1.0.0-rc.1
  original="$(git -C "$FIXTURE_WORK" rev-parse HEAD)"
  printf 'remote advance\n' >> "$FIXTURE_WORK/README.md"
  git -C "$FIXTURE_WORK" add README.md
  git -C "$FIXTURE_WORK" commit -m advance >/dev/null
  git -C "$FIXTURE_WORK" push origin main >/dev/null 2>&1
  git -C "$FIXTURE_WORK" reset --hard "$original" >/dev/null
  if run_single_release xml rc; then
    fail "branch behind origin was accepted"
  fi
  assert_equal "$original" "$(git -C "$FIXTURE_WORK" rev-parse HEAD)" "behind branch was disturbed"
}

test_existing_tag_and_malformed_manifest_rejections() {
  local original malformed

  new_fixture existing-tag
  add_module xml 1.0.0-rc.1
  commit_fixture
  tag_fixture xml 1.0.0-rc.1
  if run_single_release xml 1.0.0-rc.1; then
    fail "existing release tag was accepted"
  fi
  assert_clean

  new_fixture malformed
  add_module xml 1.0.0-rc.1
  printf 'version: 1.0.0-rc.1\n' >> "$FIXTURE_WORK/modules/xml/ferret.yaml"
  commit_fixture
  original="$(remote_main)"
  malformed="$(git -C "$FIXTURE_WORK" show HEAD:modules/xml/ferret.yaml)"
  if run_single_release xml 1.0.0-rc.1; then
    fail "manifest with duplicate version fields was accepted"
  fi
  assert_equal "$original" "$(git -C "$FIXTURE_WORK" rev-parse HEAD)" "malformed-manifest HEAD"
  assert_equal "$malformed" "$(git -C "$FIXTURE_WORK" show HEAD:modules/xml/ferret.yaml)" "malformed manifest was changed"
  assert_clean
}

test_validation_and_push_failures_roll_back() {
  local original

  new_fixture validator-failure
  add_module xml 1.0.0-rc.1
  commit_fixture
  tag_fixture xml 1.0.0-rc.1
  original="$(remote_main)"
  if (
    cd "$FIXTURE_WORK"
    FAKE_VALIDATOR_FAIL=1 FERRET_SPEC="$FIXTURE_VALIDATOR" ./scripts/release.sh xml rc
  ) > "$FIXTURE_ROOT/release.log" 2>&1; then
    fail "validator failure was accepted"
  fi
  assert_equal "$original" "$(git -C "$FIXTURE_WORK" rev-parse HEAD)" "validator-failure HEAD"
  assert_equal "1.0.0-rc.1" "$(remote_manifest_version refs/heads/main xml)" "validator-failure manifest"
  if git -C "$FIXTURE_WORK" show-ref --verify --quiet refs/tags/modules/xml/v1.0.0-rc.2; then
    fail "validator failure left a local tag"
  fi
  assert_clean

  new_fixture push-failure
  add_module xml 1.0.0-rc.1
  commit_fixture
  tag_fixture xml 1.0.0-rc.1
  original="$(remote_main)"
  printf '%s\n' '#!/usr/bin/env bash' 'exit 1' > "$FIXTURE_ORIGIN/hooks/pre-receive"
  chmod +x "$FIXTURE_ORIGIN/hooks/pre-receive"
  if run_single_release xml rc; then
    fail "rejected atomic push was accepted"
  fi
  assert_equal "$original" "$(remote_main)" "push-failure remote HEAD"
  assert_equal "$original" "$(git -C "$FIXTURE_WORK" rev-parse HEAD)" "push-failure local HEAD"
  assert_equal "1.0.0-rc.1" "$(remote_manifest_version refs/heads/main xml)" "push-failure manifest"
  if git -C "$FIXTURE_WORK" show-ref --verify --quiet refs/tags/modules/xml/v1.0.0-rc.2; then
    fail "push failure left a local tag"
  fi
  if git --git-dir "$FIXTURE_ORIGIN" show-ref --verify --quiet refs/tags/modules/xml/v1.0.0-rc.2; then
    fail "push failure created a remote tag"
  fi
  assert_clean
}

tests=(
  test_initial_and_consequent_release
  test_stale_manifest_recovery
  test_bulk_release_uses_one_commit
  test_check_only_is_non_mutating
  test_repository_state_rejections
  test_existing_tag_and_malformed_manifest_rejections
  test_validation_and_push_failures_roll_back
)

for test_name in "${tests[@]}"; do
  echo "Running $test_name"
  "$test_name"
done

echo "All release transaction tests passed."
