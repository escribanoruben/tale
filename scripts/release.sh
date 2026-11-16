#!/usr/bin/env bash
# Cut a SemVer release: tests, tag, push main + tag to Forgejo (origin).
# GitHub receives the tag via the Forgejo push mirror, which starts the
# GoReleaser Release workflow (binaries + checksums + GitHub Release).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

VERSION="${1:-}"
ORIGIN_REMOTE="${ORIGIN_REMOTE:-origin}"

if [[ -z "$VERSION" ]]; then
	echo "Usage: release.sh VERSION   (e.g. 0.1.4, without the v prefix)" >&2
	exit 1
fi

if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$ ]]; then
	echo "VERSION must be semver (e.g. 0.1.4)" >&2
	exit 1
fi

TAG="v${VERSION}"

if git rev-parse "$TAG" >/dev/null 2>&1; then
	echo "ERROR: tag $TAG already exists" >&2
	exit 1
fi

echo "==> Tests"
go test ./...

echo "==> Tag $TAG"
git tag -a "$TAG" -m "Release ${TAG}"

echo "==> Push Forgejo ($ORIGIN_REMOTE)"
branch="$(git branch --show-current)"
if [[ "$branch" == "main" ]]; then
	git push "$ORIGIN_REMOTE" main
else
	git push "$ORIGIN_REMOTE" HEAD:main
fi
git push "$ORIGIN_REMOTE" "$TAG"

echo "==> Syncing GitHub mirror (Forgejo push mirror)"
# The mirror syncs automatically when commits are pushed; a tag-only push
# may not trigger it: request a manual sync via the API using the stored
# credentials for the origin host (~/.git-credentials or credential helper).
origin_url="$(git remote get-url "$ORIGIN_REMOTE")"
host=""
repo_path=""
if [[ "$origin_url" =~ ssh://git@([^/]+)/([^?]+)$ ]]; then
	host="${BASH_REMATCH[1]}"
	repo_path="${BASH_REMATCH[2]}"
elif [[ "$origin_url" =~ git@([^:]+):([^?]+)$ ]]; then
	host="${BASH_REMATCH[1]}"
	repo_path="${BASH_REMATCH[2]}"
elif [[ "$origin_url" =~ https://([^/]+)/([^?]+)$ ]]; then
	host="${BASH_REMATCH[1]}"
	repo_path="${BASH_REMATCH[2]}"
fi
repo_path="${repo_path%.git}"

if [[ -n "$host" && -n "$repo_path" ]]; then
	creds="$(env GIT_TERMINAL_PROMPT=0 GIT_ASKPASS=/bin/true \
		git credential fill <<EOF 2>/dev/null || true
protocol=https
host=$host

EOF
)"
	token="$(sed -n 's/^password=//p' <<<"$creds")"
	if [[ -n "$token" ]]; then
		curl -sf -X POST \
			-H "Authorization: token $token" \
			-H "Content-Type: application/json" \
			"https://$host/api/v1/repos/$repo_path/push_mirrors-sync" >/dev/null \
			|| echo "WARNING: mirror sync failed; GitHub will catch up on the next push."
	else
		echo "WARNING: no stored token for $host; GitHub will sync on the next commit push."
	fi
else
	echo "WARNING: could not parse origin host ($origin_url); GitHub will sync on the next commit push."
fi

GH_URL="https://github.com/escribanoruben/tale"
echo ""
echo "Release in progress on GitHub Actions:"
echo "  ${GH_URL}/actions"
echo "  ${GH_URL}/releases/tag/${TAG}"
