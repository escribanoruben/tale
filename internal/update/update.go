// Package update checks for new tale releases on GitHub and swaps the
// running binary in place. It is network-only code: the password and
// mnemonic paths never touch this package.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/creativeprojects/go-selfupdate"

	"github.com/escribanoruben/tale/internal/paths"
)

// Repository (on GitHub) that hosts the GoReleaser release artifacts.
// Forgejo stays the primary remote; the binaries are published to the
// GitHub mirror by the release workflow.
const (
	DefaultOwner = "escribanoruben"
	DefaultRepo  = "tale"
)

// Latest is the outcome of an update check.
type Latest struct {
	Found   bool   // a published release exists
	Current string // running version
	Latest  string // latest published version
	Newer   bool   // latest is newer than current
}

// source builds the release source (nil = public GitHub API; tests pass
// a stub source pointed at a local server).
func source(src selfupdate.Source) (selfupdate.Source, error) {
	if src != nil {
		return src, nil
	}
	return selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
}

// updater builds an updater validating artifacts against the GoReleaser
// checksums.txt asset.
func updater(src selfupdate.Source) (*selfupdate.Updater, error) {
	s, err := source(src)
	if err != nil {
		return nil, err
	}
	return selfupdate.NewUpdater(selfupdate.Config{
		Source:    s,
		Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"},
	})
}

// Check compares the running version against the latest GitHub release.
func Check(ctx context.Context, src selfupdate.Source, current string) (Latest, error) {
	up, err := updater(src)
	if err != nil {
		return Latest{}, err
	}
	rel, found, err := up.DetectLatest(ctx, selfupdate.NewRepositorySlug(DefaultOwner, DefaultRepo))
	if err != nil {
		return Latest{}, err
	}
	out := Latest{Found: found, Current: current}
	if !found || rel == nil {
		return out, nil
	}
	out.Latest = rel.Version()
	out.Newer = newerThan(current, out.Latest)
	return out, nil
}

// Apply downloads the latest release and replaces the running binary.
// It reports the from → to versions. The artifact is verified against
// checksums.txt before the swap; on failure the old binary is kept.
func Apply(ctx context.Context, src selfupdate.Source, current string) (from, to string, err error) {
	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		return current, "", fmt.Errorf("locate executable: %w", err)
	}
	return ApplyAt(ctx, src, current, exe)
}

// ApplyAt is Apply targeting an explicit binary path (the tests use it
// to swap a temp copy instead of the running executable).
func ApplyAt(ctx context.Context, src selfupdate.Source, current, exePath string) (from, to string, err error) {
	up, err := updater(src)
	if err != nil {
		return current, "", err
	}
	rel, found, err := up.DetectLatest(ctx, selfupdate.NewRepositorySlug(DefaultOwner, DefaultRepo))
	if err != nil {
		return current, "", err
	}
	if !found || rel == nil {
		return current, "", fmt.Errorf("no release found")
	}
	if err := up.UpdateTo(ctx, rel, exePath); err != nil {
		return current, "", err
	}
	return current, rel.Version(), nil
}

// cacheFile stores the last successful check so the opt-in TUI notice
// does not hit the API on every start.
type cacheFile struct {
	CheckedAt time.Time `json:"checked_at"`
	Latest    string    `json:"latest"`
}

func cachePath() (string, error) {
	dir, err := paths.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "update-check.json"), nil
}

// Notice is the cached, throttled check behind the opt-in TUI banner.
// It returns "vX.Y.Z" when a newer release exists, "" otherwise (no
// release, up to date, or any error — silent by design).
func Notice(ctx context.Context, current string, maxAge time.Duration) string {
	path, err := cachePath()
	if err != nil {
		return ""
	}
	var cached cacheFile
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &cached)
	}

	latest := cached.Latest
	if cached.CheckedAt.IsZero() || time.Since(cached.CheckedAt) > maxAge {
		res, err := Check(ctx, nil, current)
		if err != nil {
			return ""
		}
		latest = res.Latest
		cached = cacheFile{CheckedAt: time.Now(), Latest: latest}
		if data, err := json.Marshal(cached); err == nil {
			_ = os.WriteFile(path, data, 0o644)
		}
	}
	if !newerThan(current, latest) {
		return ""
	}
	return latest
}

// newerThan reports whether latest is a strictly greater X.Y.Z release
// than current. Anything non-semver (e.g. "dev") yields false.
func newerThan(current, latest string) bool {
	c, ok := parseSemver(current)
	if !ok {
		return false
	}
	l, ok := parseSemver(latest)
	if !ok {
		return false
	}
	for i := range c {
		if l[i] != c[i] {
			return l[i] > c[i]
		}
	}
	return false // equal
}

// parseSemver extracts [major, minor, patch] from a version string,
// tolerating the v prefix, prerelease suffixes and build metadata.
func parseSemver(v string) ([3]int, bool) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return [3]int{}, false
	}
	var out [3]int
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return [3]int{}, false
		}
		out[i] = n
	}
	return out, true
}
