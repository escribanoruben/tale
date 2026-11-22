package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/creativeprojects/go-selfupdate"
)

func TestNewerThan(t *testing.T) {
	cases := []struct {
		current, latest string
		want            bool
	}{
		{"0.2.0", "v0.3.0", true},
		{"v0.2.0", "0.2.1", true},
		{"0.3.0", "v0.2.0", false},
		{"0.2.0", "v0.2.0", false},
		{"0.2.1", "v0.2.0", false},
		{"dev", "v9.9.9", false},
		{"0.2.0", "", false},
		{"0.2.0", "not-a-version", false},
		{"0.2.0", "1.0.0-beta", true},
	}
	for _, c := range cases {
		if got := newerThan(c.current, c.latest); got != c.want {
			t.Errorf("newerThan(%q, %q) = %v, want %v", c.current, c.latest, got, c.want)
		}
	}
}

// stubGitHub serves a GitHub-like API: releases/latest returns a
// release with the given version and two assets (the platform archive
// and checksums.txt); any other path serves a requested asset.
func stubGitHub(t *testing.T, version string, corruptChecksum bool) selfupdate.Source {
	t.Helper()
	archive := tarchive(t, "tale", []byte("new tale payload"))
	sum := fmt.Sprintf("%x", sha256.Sum256(archive))
	if corruptChecksum {
		sum = strings.Repeat("0", len(sum))
	}
	checksums := sum + "  tale_" + version + "_linux_amd64.tar.gz\n"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/releases"):
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `[{
				"tag_name": "v%s",
				"assets": [
					{"id": 1, "name": "tale_%s_linux_amd64.tar.gz",
					 "browser_download_url": "%s/download/tale_%s_linux_amd64.tar.gz",
					 "size": %d},
					{"id": 2, "name": "checksums.txt",
					 "browser_download_url": "%s/download/checksums.txt"}
				]}]`,
				version, version, srvURL(r), version, len(archive), srvURL(r))
		case strings.HasSuffix(r.URL.Path, "/checksums.txt"):
			fmt.Fprint(w, checksums)
		case strings.HasSuffix(r.URL.Path, "/assets/2"):
			fmt.Fprint(w, checksums)
		case strings.HasSuffix(r.URL.Path, "/assets/1"):
			w.Write(archive)
		default:
			w.Write(archive)
		}
	}))
	t.Cleanup(srv.Close)

	src, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{EnterpriseBaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	return src
}

func srvURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// tarchive builds a tar.gz containing one file with the given content.
func tarchive(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(content))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestCheckDetectsNewerRelease(t *testing.T) {
	src := stubGitHub(t, "99.0.0", false)
	res, err := Check(context.Background(), src, "0.2.0")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Found || !res.Newer {
		t.Fatalf("found=%v newer=%v, want true/true", res.Found, res.Newer)
	}
	if res.Latest != "99.0.0" {
		t.Fatalf("latest %q, want 99.0.0", res.Latest)
	}
}

func TestCheckUpToDate(t *testing.T) {
	src := stubGitHub(t, "0.2.0", false)
	res, err := Check(context.Background(), src, "0.2.0")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Found || res.Newer {
		t.Fatalf("found=%v newer=%v, want true/false", res.Found, res.Newer)
	}
}

func TestCheckDevBuildNeverNewer(t *testing.T) {
	src := stubGitHub(t, "99.0.0", false)
	res, err := Check(context.Background(), src, "dev")
	if err != nil {
		t.Fatal(err)
	}
	if res.Newer {
		t.Fatal("dev build should never report newer")
	}
}

func TestApplySwapsBinaryWithChecksumValidation(t *testing.T) {
	src := stubGitHub(t, "99.0.0", false)
	exe := filepath.Join(t.TempDir(), "tale")
	if err := os.WriteFile(exe, []byte("old binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Point the swap at a temp copy: ExecutablePath inside Apply would
	// resolve to the running test binary.
	from, to, err := ApplyAt(context.Background(), src, "0.2.0", exe)
	if err != nil {
		t.Fatal(err)
	}
	if from != "0.2.0" || to != "99.0.0" {
		t.Fatalf("from=%q to=%q, want 0.2.0 -> 99.0.0", from, to)
	}
	got, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new tale payload" {
		t.Fatalf("binary not swapped: %q", got)
	}
}

func TestApplyRejectsCorruptArtifact(t *testing.T) {
	src := stubGitHub(t, "99.0.0", true)
	exe := filepath.Join(t.TempDir(), "tale")
	if err := os.WriteFile(exe, []byte("old binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ApplyAt(context.Background(), src, "0.2.0", exe); err == nil {
		t.Fatal("expected checksum validation to fail")
	}
	got, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "old binary" {
		t.Fatal("corrupt download must not replace the binary")
	}
}
