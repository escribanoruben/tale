#!/usr/bin/env bash
# Next SemVer version.
# - Without BUMP: computed by svu (declared as a Go tool in go.mod) from
#   Conventional Commits since the last v* tag
#   (fix/perf/revert -> patch, feat -> minor, ! or BREAKING CHANGE -> major).
#   Fails when no commit warrants a release: use BUMP= or VERSION=.
# - With BUMP=patch|minor|major: forces the bump on top of the last tag
#   (svu cannot force minor/major; express the intent in the commits).
set -euo pipefail

case "${BUMP:-}" in
	""|patch|minor|major) ;;
	*) echo "BUMP must be patch|minor|major" >&2; exit 1 ;;
esac

cur="$(go tool svu current | sed 's/^v//')"

if [[ -n "${BUMP:-}" ]]; then
	awk -v b="$BUMP" -v v="$cur" 'BEGIN {
		split(v, a, ".")
		if (b == "major") printf "%d.0.0\n", a[1]+1
		else if (b == "minor") printf "%d.%d.0\n", a[1], a[2]+1
		else printf "%d.%d.%d\n", a[1], a[2], a[3]+1
	}'
	exit 0
fi

next="$(go tool svu next)"
if [[ "$next" == "v$cur" ]]; then
	echo "No commits warrant a release. Use BUMP=patch|minor|major or VERSION=x.y.z" >&2
	exit 1
fi
echo "$next"
