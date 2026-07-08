#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "Usage: make release-pre-all <semver|preid>" >&2
  echo "Examples:" >&2
  echo "  make release-pre-all 1.0.0-rc.1" >&2
  echo "  make release-pre-all rc" >&2
  echo "" >&2
  echo "Direct script usage: $0 <semver|preid>" >&2
  exit 1
fi

VERSION_OR_PREID="$1"

modules=()
while IFS= read -r module; do
  modules+=("$module")
done < <(make modules)

if [ "${#modules[@]}" -eq 0 ]; then
  echo "No modules found" >&2
  exit 1
fi

for module in "${modules[@]}"; do
  echo "Checking module '$module' with '$VERSION_OR_PREID'"
  RELEASE_CHECK_ONLY=1 ./scripts/release.sh "$module" "$VERSION_OR_PREID"
done

if [[ "${RELEASE_CHECK_ONLY:-}" == "1" ]]; then
  echo "Release check passed for all modules; no tags created."
  exit 0
fi

for module in "${modules[@]}"; do
  echo "Releasing module '$module' with '$VERSION_OR_PREID'"
  make release-pre "$module" "$VERSION_OR_PREID"
done
