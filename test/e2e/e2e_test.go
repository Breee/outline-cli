package e2e

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPushMarkdownViaAPI(t *testing.T) {
	serverURL := os.Getenv("OUTLINE_SERVER_URL")
	if serverURL == "" {
		t.Skip("OUTLINE_SERVER_URL not set")
	}

	tmp := t.TempDir()
	md := filepath.Join(tmp, "doc.md")
	if err := os.WriteFile(md, []byte("# E2E Title\nbody"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "run", ".", "--server-url", serverURL, "--api-token", "e2e-token", "push", "--collection-id", "col-1", "--path", md)
	cmd.Dir = filepath.Clean("../..")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("push command failed: %v\n%s", err, string(output))
	}
	if !strings.Contains(string(output), "created") && !strings.Contains(string(output), "updated") {
		t.Fatalf("expected push output, got: %s", output)
	}

	resp, err := http.Get(serverURL + "/api/documents/E2E Title")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	var payload struct {
		Data struct {
			Title string `json:"title"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.Title != "E2E Title" {
		t.Fatalf("unexpected title: %s", payload.Data.Title)
	}
}
