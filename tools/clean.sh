#!/bin/bash

if [ $# -lt 1 ] || [ $# -gt 2 ]; then
  echo "Usage: $0 <days> [owner/repo]"
  exit 1
fi

DAYS="$1"
SECS_AGO=$((DAYS * 24 * 60 * 60))

# If a repo is provided as $2, use it; otherwise derive from git remote.
if [ $# -eq 2 ]; then
  REPO="$2"
else
  REPO=$(git remote get-url origin | sed -n 's#.*github.com[:/]\([^/]\+\)/\([^/.]\+\).*#\1/\2#p')
fi

if [ -z "$REPO" ]; then
  echo "Could not determine GitHub repo. Provide as [owner/repo] or set a valid git remote." >&2
  exit 2
fi

echo "== Deleting releases older than $DAYS days in $REPO =="

gh release list --repo "$REPO" --limit 1000 \
  --json tagName,createdAt \
  --jq "map(select((now - (.createdAt | fromdateiso8601)) > $SECS_AGO)) | .[] | .tagName" \
  | xargs -r -I{} gh release delete --repo "$REPO" --cleanup-tag --yes "{}"
