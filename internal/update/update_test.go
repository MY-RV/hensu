package update_test

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/my-rv/hensu/internal/update"
)

func TestNewer(t *testing.T) {
	ok, err := update.Newer("0.2.0-dev", "v0.2.1")
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	ok, err = update.Newer("v0.2.0", "v0.1.9")
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	ok, err = update.Newer("v0.2.0", "0.2.0")
	if err != nil || ok {
		t.Fatalf("same should not be newer: %v %v", ok, err)
	}
}

func TestClient_Latest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/releases/latest" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "tag_name": "v0.2.0",
  "assets": [
    {"name": "hensu_0.2.0_darwin_arm64", "browser_download_url": "http://example.com/hensu"},
    {"name": "checksums.txt", "browser_download_url": "http://example.com/sums"}
  ]
}`))
	}))
	defer srv.Close()

	c := &update.Client{HTTP: srv.Client(), APIBase: srv.URL, GOOS: "darwin", GOARCH: "arm64"}
	cand, err := c.Latest()
	if err != nil {
		t.Fatal(err)
	}
	if cand.Tag != "v0.2.0" || cand.AssetName != "hensu_0.2.0_darwin_arm64" {
		t.Fatalf("%+v", cand)
	}
	if !strings.Contains(cand.DownloadURL, "example.com") || !strings.Contains(cand.ChecksumsURL, "sums") {
		t.Fatalf("%+v", cand)
	}
}

func TestClient_LatestSkipsArchives(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
  "tag_name": "v0.2.0",
  "assets": [
    {"name": "hensu_0.2.0_darwin_arm64.tar.gz", "browser_download_url": "http://example.com/archive"},
    {"name": "checksums.txt", "browser_download_url": "http://example.com/sum"}
  ]
}`))
	}))
	defer srv.Close()
	c := &update.Client{HTTP: srv.Client(), APIBase: srv.URL, GOOS: "darwin", GOARCH: "arm64"}
	cand, err := c.Latest()
	if err == nil || cand.DownloadURL != "" {
		t.Fatalf("expected no bare binary, got %+v err=%v", cand, err)
	}
}

func TestClient_LatestNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	c := &update.Client{HTTP: srv.Client(), APIBase: srv.URL}
	_, err := c.Latest()
	if err == nil || !strings.Contains(err.Error(), "no GitHub releases") {
		t.Fatalf("%v", err)
	}
}

// releaseServer serves one binary plus a checksums.txt describing it. When
// corrupt is set, the served bytes no longer match the advertised digest.
func releaseServer(t *testing.T, body string, corrupt bool) (*httptest.Server, update.Candidate) {
	t.Helper()
	sum := sha256.Sum256([]byte(body))
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	served := body
	if corrupt {
		served = body + "tampered"
	}
	mux.HandleFunc("/bin", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(served))
	})
	mux.HandleFunc("/sums", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(hex.EncodeToString(sum[:]) + "  hensu_9.9.9_linux_amd64\n"))
	})
	return srv, update.Candidate{
		Tag:          "v9.9.9",
		AssetName:    "hensu_9.9.9_linux_amd64",
		DownloadURL:  srv.URL + "/bin",
		ChecksumsURL: srv.URL + "/sums",
	}
}

func TestDownload_verifiesChecksum(t *testing.T) {
	srv, cand := releaseServer(t, "#!/bin/sh\necho ok\n", false)
	dest := filepath.Join(t.TempDir(), "hensu")
	if err := update.Download(srv.Client(), cand, dest); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(dest)
	if err != nil || !strings.Contains(string(data), "echo ok") {
		t.Fatalf("%q %v", data, err)
	}
}

func TestDownload_rejectsTamperedBytes(t *testing.T) {
	srv, cand := releaseServer(t, "#!/bin/sh\necho ok\n", true)
	dest := filepath.Join(t.TempDir(), "hensu")
	if err := os.WriteFile(dest, []byte("original"), 0o755); err != nil {
		t.Fatal(err)
	}
	err := update.Download(srv.Client(), cand, dest)
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("want checksum mismatch, got %v", err)
	}
	data, _ := os.ReadFile(dest)
	if string(data) != "original" {
		t.Fatalf("a failed update must leave the binary alone, got %q", data)
	}
}

func TestDownload_refusesWithoutChecksums(t *testing.T) {
	srv, cand := releaseServer(t, "x", false)
	cand.ChecksumsURL = ""
	err := update.Download(srv.Client(), cand, filepath.Join(t.TempDir(), "hensu"))
	if err == nil || !strings.Contains(err.Error(), "unverified") {
		t.Fatalf("%v", err)
	}
}

func TestDownload_refusesWhenAssetIsNotListed(t *testing.T) {
	srv, cand := releaseServer(t, "x", false)
	cand.AssetName = "hensu_9.9.9_plan9_arm"
	err := update.Download(srv.Client(), cand, filepath.Join(t.TempDir(), "hensu"))
	if err == nil || !strings.Contains(err.Error(), "no entry for") {
		t.Fatalf("%v", err)
	}
}
