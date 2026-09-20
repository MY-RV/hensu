package cli_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/my-rv/hensu"
	"github.com/my-rv/hensu/internal/cli"
	"github.com/my-rv/hensu/internal/update"
)

func TestApp_updateCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/releases/latest":
			_, _ = w.Write([]byte(`{
  "tag_name": "v9.9.9",
  "assets": [{"name":"hensu_9.9.9_darwin_arm64","browser_download_url":"REPLACE"}]
}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	prev := hensu.Version
	hensu.Version = "0.2.0"
	t.Cleanup(func() { hensu.Version = prev })

	var stdout bytes.Buffer
	app := cli.New()
	app.Stdout = &stdout
	app.UpdateClient = &update.Client{
		HTTP:    srv.Client(),
		APIBase: srv.URL,
		GOOS:    "darwin",
		GOARCH:  "arm64",
	}
	if err := app.Run([]string{"--update-check"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "update available") {
		t.Fatalf("%q", stdout.String())
	}
}

func TestApp_updateDownloads(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "hensu-bin")
	if err := os.WriteFile(dest, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()
	downloadURL := srv.URL + "/bin"
	sum := sha256.Sum256([]byte("new-binary"))
	mux.HandleFunc("/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
  "tag_name": "v9.9.9",
  "assets": [
    {"name":"hensu_9.9.9_linux_amd64","browser_download_url":"` + downloadURL + `"},
    {"name":"checksums.txt","browser_download_url":"` + srv.URL + `/sums"}
  ]
}`))
	})
	mux.HandleFunc("/bin", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("new-binary"))
	})
	mux.HandleFunc("/sums", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(hex.EncodeToString(sum[:]) + "  hensu_9.9.9_linux_amd64\n"))
	})

	prev := hensu.Version
	hensu.Version = "0.2.0"
	t.Cleanup(func() { hensu.Version = prev })

	var stdout bytes.Buffer
	app := cli.New()
	app.Stdout = &stdout
	app.UpdateClient = &update.Client{
		HTTP:    srv.Client(),
		APIBase: srv.URL,
		GOOS:    "linux",
		GOARCH:  "amd64",
	}
	app.Executable = func() (string, error) { return dest, nil }
	if err := app.Run([]string{"--update"}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(dest)
	if string(data) != "new-binary" {
		t.Fatalf("%q", data)
	}
}
