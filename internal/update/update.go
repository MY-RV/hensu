// Package update checks GitHub Releases and can replace the running binary.
//
// Until github.com/my-rv/hensu publishes releases, Latest fails with a clear error
// unless APIBase is overridden (tests / future mirrors).
package update

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// maxBinary bounds what self-update will write to disk before giving up.
const maxBinary = 256 << 20

// ChecksumsName is the file GoReleaser publishes alongside the binaries.
const ChecksumsName = "checksums.txt"

// DefaultAPIBase is the GitHub API root for this project when it has its own repo.
const DefaultAPIBase = "https://api.github.com/repos/my-rv/hensu"

// Client talks to a GitHub-like releases API.
type Client struct {
	HTTP    *http.Client
	APIBase string // e.g. DefaultAPIBase; override in tests
	GOOS    string
	GOARCH  string
}

func (c *Client) http() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 30 * time.Second}
}

func (c *Client) apiBase() string {
	if c.APIBase != "" {
		return strings.TrimRight(c.APIBase, "/")
	}
	if v := strings.TrimSpace(os.Getenv("HENSU_RELEASES_API")); v != "" {
		return strings.TrimRight(v, "/")
	}
	return DefaultAPIBase
}

func (c *Client) goos() string {
	if c.GOOS != "" {
		return c.GOOS
	}
	return runtime.GOOS
}

func (c *Client) goarch() string {
	if c.GOARCH != "" {
		return c.GOARCH
	}
	return runtime.GOARCH
}

// Release is a trimmed GitHub release payload.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset is a release binary.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Candidate is the asset self-update would install, together with the
// checksums file that vouches for it. Both travel as one value so no caller
// can download the first while forgetting the second.
type Candidate struct {
	Tag          string
	AssetName    string
	DownloadURL  string
	ChecksumsURL string
}

// Latest fetches /releases/latest and picks the asset for this OS/arch.
// Prefers bare binaries (scripts/release-local.sh); also accepts archive names
// that contain hensu + os + arch (GoReleaser).
func (c *Client) Latest() (Candidate, error) {
	url := c.apiBase() + "/releases/latest"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Candidate{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	res, err := c.http().Do(req)
	if err != nil {
		return Candidate{}, fmt.Errorf("check releases: %w (repo may not be published yet)", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode == http.StatusNotFound {
		return Candidate{}, fmt.Errorf("no GitHub releases at %s (publish the repo + a release, or set HENSU_RELEASES_API)", c.apiBase())
	}
	if res.StatusCode != http.StatusOK {
		return Candidate{}, fmt.Errorf("releases API %s: %s", res.Status, truncate(body, 200))
	}
	var rel Release
	if err := json.Unmarshal(body, &rel); err != nil {
		return Candidate{}, err
	}
	if rel.TagName == "" {
		return Candidate{}, fmt.Errorf("latest release has no tag_name")
	}
	cand := Candidate{Tag: rel.TagName}
	want := assetName(rel.TagName, c.goos(), c.goarch())
	var fallback Asset
	for _, a := range rel.Assets {
		if a.Name == ChecksumsName {
			cand.ChecksumsURL = a.BrowserDownloadURL
			continue
		}
		if a.Name == want {
			cand.AssetName, cand.DownloadURL = a.Name, a.BrowserDownloadURL
			continue
		}
		if fallback.Name == "" && matchBareOrArchive(a.Name, c.goos(), c.goarch()) {
			fallback = a
		}
	}
	if cand.DownloadURL == "" && fallback.Name != "" {
		cand.AssetName, cand.DownloadURL = fallback.Name, fallback.BrowserDownloadURL
	}
	if cand.DownloadURL == "" {
		return cand, fmt.Errorf("release %s has no asset for %s/%s (want %q)", rel.TagName, c.goos(), c.goarch(), want)
	}
	return cand, nil
}

// assetName is the convention used by scripts/release-local.sh (bare binary).
func assetName(tag, goos, goarch string) string {
	tag = strings.TrimPrefix(tag, "v")
	return fmt.Sprintf("hensu_%s_%s_%s", tag, goos, goarch)
}

func matchBareOrArchive(name, goos, goarch string) bool {
	n := strings.ToLower(name)
	if strings.Contains(n, ".tar.gz") || strings.Contains(n, ".zip") || strings.HasSuffix(n, ".txt") {
		// Archives need extract; self-update expects a raw binary asset.
		return false
	}
	return strings.Contains(n, "hensu") && strings.Contains(n, goos) && strings.Contains(n, goarch)
}

// Newer reports whether latest is a newer semver than current (v-prefix optional).
func Newer(current, latest string) (bool, error) {
	c, err := parseSemver(current)
	if err != nil {
		return false, fmt.Errorf("current version %q: %w", current, err)
	}
	l, err := parseSemver(latest)
	if err != nil {
		return false, fmt.Errorf("latest version %q: %w", latest, err)
	}
	return cmpSemver(l, c) > 0, nil
}

type semver struct{ major, minor, patch int }

func parseSemver(s string) (semver, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i] // drop -dev / +build
	}
	parts := strings.Split(s, ".")
	if len(parts) < 1 || len(parts) > 3 {
		return semver{}, fmt.Errorf("want major.minor.patch")
	}
	var v semver
	var err error
	if v.major, err = atoi(parts[0]); err != nil {
		return semver{}, err
	}
	if len(parts) > 1 {
		if v.minor, err = atoi(parts[1]); err != nil {
			return semver{}, err
		}
	}
	if len(parts) > 2 {
		if v.patch, err = atoi(parts[2]); err != nil {
			return semver{}, err
		}
	}
	return v, nil
}

func atoi(s string) (int, error) {
	n := 0
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("not a number %q", s)
		}
		n = n*10 + int(r-'0')
	}
	return n, nil
}

func cmpSemver(a, b semver) int {
	if a.major != b.major {
		return a.major - b.major
	}
	if a.minor != b.minor {
		return a.minor - b.minor
	}
	return a.patch - b.patch
}

// Download replaces destPath with cand's asset, and only if its SHA-256
// matches the release checksums file.
//
// hensu handles secrets; an update channel that replaces the binary on
// whatever bytes arrive would be a worse hole than any it closes. A release
// without checksums is refused rather than trusted.
func Download(httpClient *http.Client, cand Candidate, destPath string) error {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 120 * time.Second}
	}
	if cand.DownloadURL == "" {
		return fmt.Errorf("no download URL for this platform")
	}
	if cand.ChecksumsURL == "" {
		return fmt.Errorf("release %s publishes no %s; refusing to replace the binary unverified", cand.Tag, ChecksumsName)
	}
	want, err := fetchChecksum(httpClient, cand.ChecksumsURL, cand.AssetName)
	if err != nil {
		return err
	}

	res, err := httpClient.Get(cand.DownloadURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: %s", cand.DownloadURL, res.Status)
	}
	dir := filepath.Dir(destPath)
	tmp, err := os.CreateTemp(dir, ".hensu-update-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	sum := sha256.New()
	n, err := io.Copy(io.MultiWriter(tmp, sum), io.LimitReader(res.Body, maxBinary+1))
	if err != nil {
		_ = tmp.Close()
		return err
	}
	if n > maxBinary {
		_ = tmp.Close()
		return fmt.Errorf("download %s: larger than %d bytes", cand.AssetName, maxBinary)
	}
	if got := hex.EncodeToString(sum.Sum(nil)); got != want {
		_ = tmp.Close()
		return fmt.Errorf("checksum mismatch for %s: got %s, %s says %s", cand.AssetName, got, ChecksumsName, want)
	}
	if err := tmp.Chmod(0o755); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, destPath)
}

// fetchChecksum returns the expected SHA-256 for name from a GoReleaser-style
// checksums file ("<hex>  <name>" per line).
func fetchChecksum(httpClient *http.Client, url, name string) (string, error) {
	res, err := httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", ChecksumsName, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch %s: %s", ChecksumsName, res.Status)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(body), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		if strings.TrimPrefix(fields[1], "*") == name {
			sum := strings.ToLower(fields[0])
			if len(sum) != sha256.Size*2 {
				return "", fmt.Errorf("%s: %q is not a SHA-256 digest", ChecksumsName, fields[0])
			}
			return sum, nil
		}
	}
	return "", fmt.Errorf("%s has no entry for %s", ChecksumsName, name)
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "…"
}
