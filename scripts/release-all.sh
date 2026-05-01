#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-1.0.0-rc.1}"
REMOTE="${2:-origin}"

modules=()
while IFS= read -r module; do
  modules+=("$module")
done < <(make modules)

if [ "${#modules[@]}" -eq 0 ]; then
  echo "No modules found" >&2
  exit 1
fi

tags=()

for module in "${modules[@]}"; do
  tag="modules/$module/v$VERSION"
  echo "Creating tag: $tag"
  make release-pre "$VERSION" "$module"
  tags+=("$tag")
done

echo "Pushing ${#tags[@]} tags to remote '$REMOTE'"
git push "$REMOTE" "${tags[@]}"
