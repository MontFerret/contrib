#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/contrib-publications-tests.XXXXXX")"
trap 'rm -rf "$TEST_ROOT"' EXIT

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

expect_matrix() {
  local expected="$1"
  local actual
  shift

  actual="$(bash "$SCRIPT_DIR/discover-publications.sh" "$@")" ||
    fail "discovery failed for $*"
  if [[ "$(jq -Sc . <<< "$actual")" != "$(jq -Sc . <<< "$expected")" ]]; then
    fail "$*: got $actual, want $expected"
  fi
}

expect_failure() {
  if bash "$SCRIPT_DIR/discover-publications.sh" "$@" > "$TEST_ROOT/output" 2> "$TEST_ROOT/error"; then
    fail "accepted invalid request: $*"
  fi
  [[ -s "$TEST_ROOT/error" ]] || fail "missing error for $*"
}

cd "$TEST_ROOT"
git init --initial-branch=main --quiet
git config user.name "Publication Test"
git config user.email "publication-test@example.com"
git commit --allow-empty -m initial --quiet
initial="$(git rev-parse HEAD)"
zero='0000000000000000000000000000000000000000'

expect_matrix '{"include":[]}' range "$zero" "$initial"
expect_matrix '{"include":[]}' range "$initial" "$initial"
expect_failure ref "$initial"

git tag -a modules/old/v1.0.0 -m old
git commit --allow-empty -m release --quiet
release="$(git rev-parse HEAD)"
records=''
for i in $(seq 1 18); do
  tag="modules/category/module-$i/v1.0.0-rc.1"
  git tag -a "$tag" -m "$tag"
  records+="$(jq -cn --arg tag "$tag" --arg commit "$release" '{tag: $tag, commit: $commit}')"$'\n'
done
batch="$(jq -sc '{include: sort_by(.tag)}' <<< "$records")"

# Unrelated and malformed tag names must not become matrix entries.
git tag pkg/common/v1.0.0
git tag other/v1.0.0
git tag modules/invalid/v1.0
git tag modules/invalid/v1.0.0-01
git tag modules/invalid_/v1.0.0

expect_matrix "$batch" range "$initial" "$release"
expect_matrix "$batch" ref "$release"
single="$(jq -c '{include: [.include[] | select(.tag == "modules/category/module-1/v1.0.0-rc.1")]}' <<< "$batch")"
expect_matrix "$single" ref modules/category/module-1/v1.0.0-rc.1

git commit --allow-empty -m second-release --quiet
second="$(git rev-parse HEAD)"
git tag modules/xml/v1.2.3+build.1
lightweight="$(jq -cn --arg commit "$second" '{include: [{tag: "modules/xml/v1.2.3+build.1", commit: $commit}]}')"
expect_matrix "$lightweight" ref modules/xml/v1.2.3+build.1
expect_matrix "$lightweight" ref "$second"
combined="$(jq -cn --argjson first "$batch" --argjson second "$lightweight" '{include: ($first.include + $second.include | sort_by(.tag))}')"
expect_matrix "$combined" range "$initial" "$second"

# New-branch pushes include all introduced history; later ordinary commits do not.
all="$(jq -c --arg commit "$initial" '.include += [{tag: "modules/old/v1.0.0", commit: $commit}] | .include |= sort_by(.tag)' <<< "$combined")"
expect_matrix "$all" range "$zero" "$second"
git commit --allow-empty -m ordinary --quiet
ordinary="$(git rev-parse HEAD)"
expect_matrix '{"include":[]}' range "$second" "$ordinary"
expect_matrix '{"include":[]}' range "$ordinary" "$zero"
expect_failure ref "$ordinary"

for invalid in main HEAD "${release:0:7}" "$zero" modules/missing/v1.0.0 \
  modules/invalid/v1.0 modules/invalid/v1.0.0-01 modules/invalid_/v1.0.0 \
  pkg/common/v1.0.0 'modules/../xml/v1.0.0' '--help'; do
  expect_failure ref "$invalid"
done
expect_failure range invalid "$release"
expect_failure range "$release" invalid
expect_failure range ffffffffffffffffffffffffffffffffffffffff "$release"
expect_failure ref
expect_failure range "$initial"
expect_failure unexpected

# A release-shaped tag that targets a blob cannot supply module source.
blob="$(printf 'not a commit' | git hash-object -w --stdin)"
git tag modules/blob/v1.0.0 "$blob"
expect_failure ref modules/blob/v1.0.0
expect_matrix "$combined" range "$initial" "$second"
expect_matrix "$batch" ref "$release"

echo "All publication discovery tests passed."
