package e2e

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	serverURL string
	apiToken  string
)

func TestMain(m *testing.M) {
	serverURL = os.Getenv("OUTLINE_SERVER_URL")
	if serverURL == "" {
		fmt.Println("OUTLINE_SERVER_URL not set, skipping e2e tests")
		os.Exit(0)
	}

	// If an API token is provided directly, use it.
	apiToken = os.Getenv("OUTLINE_API_TOKEN")
	if apiToken == "" {
		// Bootstrap: perform OIDC login via Dex to get a session, then create an API key.
		var err error
		apiToken, err = bootstrapAPIToken(serverURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to bootstrap API token: %v\n", err)
			os.Exit(1)
		}
	}

	os.Exit(m.Run())
}

// bootstrapAPIToken performs a programmatic OIDC login through Dex and creates an API key.
func bootstrapAPIToken(baseURL string) (string, error) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
		// Follow redirects but track them.
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 20 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	// Step 1: Hit Outline's OIDC auth endpoint to get redirected to Dex.
	resp, err := client.Get(baseURL + "/auth/oidc")
	if err != nil {
		return "", fmt.Errorf("initiating OIDC: %w", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	fmt.Fprintf(os.Stderr, "OIDC step1: status=%d url=%s bodyLen=%d\n", resp.StatusCode, resp.Request.URL, len(body))

	// We should now be at Dex's login page. Extract the action URL from the form.
	loginURL := extractFormAction(string(body), resp.Request.URL)
	if loginURL == "" {
		return "", fmt.Errorf("could not find login form action in Dex response (status %d, url %s, body: %s)", resp.StatusCode, resp.Request.URL, string(body[:min(len(body), 500)]))
	}
	fmt.Fprintf(os.Stderr, "OIDC step2: posting to %s\n", loginURL)

	// Step 2: Submit the login form to Dex.
	resp, err = client.PostForm(loginURL, url.Values{
		"login":    {"admin@example.com"},
		"password": {"password"},
	})
	if err != nil {
		return "", fmt.Errorf("submitting Dex login: %w", err)
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()

	fmt.Fprintf(os.Stderr, "OIDC step2 response: status=%d url=%s bodyLen=%d\n", resp.StatusCode, resp.Request.URL, len(body))

	// Check if there's an approval form we need to submit
	if strings.Contains(string(body), `<form`) {
		approvalURL := extractFormAction(string(body), resp.Request.URL)
		if approvalURL != "" {
			fmt.Fprintf(os.Stderr, "OIDC step2b: submitting approval form to %s\n", approvalURL)
			resp, err = client.PostForm(approvalURL, url.Values{"approval": {"approve"}})
			if err != nil {
				return "", fmt.Errorf("submitting approval: %w", err)
			}
			body, _ = io.ReadAll(resp.Body)
			resp.Body.Close()
			fmt.Fprintf(os.Stderr, "OIDC step2b response: status=%d url=%s\n", resp.StatusCode, resp.Request.URL)
		}
	}

	// After following redirects, we should be back at Outline with a session cookie.
	// Verify we have a valid session.
	req, _ := http.NewRequest("POST", baseURL+"/api/auth.info", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		return "", fmt.Errorf("checking auth: %w", err)
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()

	fmt.Fprintf(os.Stderr, "OIDC step3 auth.info: status=%d body=%s\n", resp.StatusCode, string(body[:min(len(body), 300)]))

	var authResp struct {
		OK   bool `json:"ok"`
		Data struct {
			User struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &authResp); err != nil {
		return "", fmt.Errorf("decoding auth.info: %w", err)
	}
	if !authResp.OK {
		return "", fmt.Errorf("auth.info returned ok=false after login")
	}

	// Get the CSRF token from the "csrfToken" cookie set by Outline on GET requests.
	// Outline uses the double-submit cookie pattern: the cookie value must also be
	// sent in the "x-csrf-token" header on mutating requests.
	csrfToken := readCSRFCookie(jar, baseURL)
	if csrfToken == "" {
		// Make a GET request to trigger Outline's attachCSRFToken middleware.
		getResp, getErr := client.Get(baseURL)
		if getErr == nil {
			io.Copy(io.Discard, getResp.Body)
			getResp.Body.Close()
			csrfToken = readCSRFCookie(jar, baseURL)
		}
	}
	fmt.Fprintf(os.Stderr, "CSRF token found: %v\n", csrfToken != "")

	// Step 3: Create an API key using the session.
	req, _ = http.NewRequest("POST", baseURL+"/api/apiKeys.create", strings.NewReader(`{"name":"e2e-test"}`))
	req.Header.Set("Content-Type", "application/json")
	if csrfToken != "" {
		req.Header.Set("x-csrf-token", csrfToken)
	}
	resp, err = client.Do(req)
	if err != nil {
		return "", fmt.Errorf("creating API key: %w", err)
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)

	var keyResp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
		Data  struct {
			Value string `json:"value"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &keyResp); err != nil {
		return "", fmt.Errorf("decoding apiKeys.create: %w", err)
	}
	if !keyResp.OK {
		return "", fmt.Errorf("apiKeys.create returned ok=false (status %d, error: %s)", resp.StatusCode, keyResp.Error)
	}
	if keyResp.Data.Value == "" {
		return "", fmt.Errorf("apiKeys.create returned empty value (status %d, body: %s)", resp.StatusCode, string(body[:min(len(body), 500)]))
	}

	return keyResp.Data.Value, nil
}

// readCSRFCookie returns the value of the "csrfToken" cookie from the jar for
// the given base URL. Outline's CSRF uses a double-submit cookie pattern where
// the cookie value must also be echoed in the "x-csrf-token" request header.
func readCSRFCookie(jar http.CookieJar, baseURL string) string {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	for _, c := range jar.Cookies(parsed) {
		if c.Name == "csrfToken" {
			return c.Value
		}
	}
	return ""
}

// extractFormAction finds the action attribute of the first <form> in HTML.
func extractFormAction(htmlBody string, baseURL *url.URL) string {
	// Simple extraction — find action="..." in form tag.
	idx := strings.Index(htmlBody, "<form")
	if idx == -1 {
		return ""
	}
	formTag := htmlBody[idx:]
	end := strings.Index(formTag, ">")
	if end == -1 {
		return ""
	}
	formTag = formTag[:end]

	actionIdx := strings.Index(formTag, `action="`)
	if actionIdx == -1 {
		return ""
	}
	actionIdx += len(`action="`)
	actionEnd := strings.Index(formTag[actionIdx:], `"`)
	if actionEnd == -1 {
		return ""
	}
	action := formTag[actionIdx : actionIdx+actionEnd]
	action = html.UnescapeString(action)

	// Resolve relative URL.
	parsed, err := url.Parse(action)
	if err != nil {
		return ""
	}
	return baseURL.ResolveReference(parsed).String()
}

func TestExtractFormActionUnescapesHTML(t *testing.T) {
	baseURL, err := url.Parse("http://dex:5556/dex/auth/local/login?back=&state=abc")
	if err != nil {
		t.Fatal(err)
	}

	got := extractFormAction(`<form action="/dex/auth/local/login?back=&amp;state=is4ydxmpqi2dm7uysxlmjjk7c" method="post">`, baseURL)
	want := "http://dex:5556/dex/auth/local/login?back=&state=is4ydxmpqi2dm7uysxlmjjk7c"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	allArgs := append([]string{"run", ".", "--server-url", serverURL, "--api-token", apiToken}, args...)
	cmd := exec.Command("go", allArgs...)
	cmd.Dir = filepath.Clean("../..")
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func TestAuthCheck(t *testing.T) {
	out, err := runCLI(t, "auth", "check")
	if err != nil {
		t.Fatalf("auth check failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "admin") {
		t.Fatalf("expected username 'admin' in output, got: %s", out)
	}
}

func TestPushSingleFile(t *testing.T) {
	tmp := t.TempDir()
	md := filepath.Join(tmp, "e2e-single.md")
	content := "# E2E Single\n\nThis is a test document."
	if err := os.WriteFile(md, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a collection first.
	out, err := runCLI(t, "push", "--yes", "--path", md, "--collection-id", "E2E Tests", "--create-collection")
	if err != nil {
		t.Fatalf("push failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "created") {
		t.Fatalf("expected 'created' in output, got: %s", out)
	}
}

func TestPushDirectory(t *testing.T) {
	tmp := t.TempDir()

	// Create index.md (parent) and child.md.
	if err := os.WriteFile(filepath.Join(tmp, "index.md"), []byte("# E2E Parent\n\nParent doc."), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "child.md"), []byte("# E2E Child\n\nChild doc."), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCLI(t, "push", "--yes", "--path", tmp, "--collection-id", "E2E Tests", "--create-collection")
	if err != nil {
		t.Fatalf("push failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "created") && !strings.Contains(out, "updated") {
		t.Fatalf("expected push results, got: %s", out)
	}
}

func TestPushThenUpdate(t *testing.T) {
	tmp := t.TempDir()
	md := filepath.Join(tmp, "e2e-update.md")

	// Push initial version.
	if err := os.WriteFile(md, []byte("# E2E Update\n\nVersion 1."), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, "push", "--yes", "--path", md, "--collection-id", "E2E Tests", "--create-collection")
	if err != nil {
		t.Fatalf("initial push failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "created") {
		t.Fatalf("expected 'created', got: %s", out)
	}

	// Push updated version.
	if err := os.WriteFile(md, []byte("# E2E Update\n\nVersion 2."), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err = runCLI(t, "push", "--yes", "--path", md, "--collection-id", "E2E Tests")
	if err != nil {
		t.Fatalf("update push failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "updated") {
		t.Fatalf("expected 'updated', got: %s", out)
	}
}

func TestSearch(t *testing.T) {
	// Ensure there's something to find — push a doc with a unique keyword.
	tmp := t.TempDir()
	md := filepath.Join(tmp, "searchable.md")
	if err := os.WriteFile(md, []byte("# E2E Searchable\n\nUniqueKeyword12345 content."), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, "push", "--yes", "--path", md, "--collection-id", "E2E Tests", "--create-collection")
	if err != nil {
		t.Fatalf("push failed: %v\n%s", err, out)
	}

	// Search for it.
	out, err = runCLI(t, "search", "UniqueKeyword12345")
	if err != nil {
		t.Fatalf("search failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "E2E Searchable") {
		t.Fatalf("expected search result, got: %s", out)
	}
}

func TestSearchJSON(t *testing.T) {
	out, err := runCLI(t, "search", "E2E", "--format", "json")
	if err != nil {
		t.Fatalf("search json failed: %v\n%s", err, out)
	}
	var results []json.RawMessage
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("expected valid JSON array, got: %s", out)
	}
}

func TestGetDocuments(t *testing.T) {
	// List documents in the E2E Tests collection.
	out, err := runCLI(t, "get", "documents", "--collection", "E2E Tests")
	if err != nil {
		t.Fatalf("get documents failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "E2E") {
		t.Fatalf("expected E2E documents in output, got: %s", out)
	}
}

func TestGetDocumentByTitle(t *testing.T) {
	out, err := runCLI(t, "get", "documents", "E2E Single", "-o", "raw")
	if err != nil {
		t.Fatalf("get document by title failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "This is a test document") {
		t.Fatalf("expected document content, got: %s", out)
	}
}

func TestPullCollection(t *testing.T) {
	tmp := t.TempDir()
	out, err := runCLI(t, "pull", "--collection", "E2E Tests", "--output", tmp)
	if err != nil {
		t.Fatalf("pull failed: %v\n%s", err, out)
	}

	// Verify at least one file was created.
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected pulled files, got empty directory")
	}
}

func TestPullSingleDoc(t *testing.T) {
	tmp := t.TempDir()
	outFile := filepath.Join(tmp, "pulled.md")
	out, err := runCLI(t, "pull", "--doc", "E2E Single", "--output", outFile)
	if err != nil {
		t.Fatalf("pull doc failed: %v\n%s", err, out)
	}

	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("reading pulled file: %v", err)
	}
	if !strings.Contains(string(content), "test document") {
		t.Fatalf("expected document content in pulled file, got: %s", string(content))
	}
}

func TestPushWithMetadata(t *testing.T) {
	tmp := t.TempDir()
	md := filepath.Join(tmp, "metadata.md")
	content := `<!-- Title: E2E Metadata Title -->
<!-- Collection: E2E Tests -->

# Ignored H1

This document uses metadata headers for title and collection.`
	if err := os.WriteFile(md, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCLI(t, "push", "--yes", "--path", md, "--collection-id", "E2E Tests", "--create-collection")
	if err != nil {
		t.Fatalf("push with metadata failed: %v\n%s", err, out)
	}

	// The document should be created with "E2E Metadata Title", not "Ignored H1".
	out, err = runCLI(t, "get", "documents", "E2E Metadata Title", "-o", "json")
	if err != nil {
		t.Fatalf("get metadata doc failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "E2E Metadata Title") {
		t.Fatalf("expected metadata title in output, got: %s", out)
	}
}

func TestVersion(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "version")
	cmd.Dir = filepath.Clean("../..")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version failed: %v\n%s", err, string(out))
	}
	if !strings.Contains(string(out), "v") {
		t.Fatalf("expected version string, got: %s", string(out))
	}
}

func TestVersionJSON(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "version", "-o", "json")
	cmd.Dir = filepath.Clean("../..")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version -o json failed: %v\n%s", err, string(out))
	}
	var v map[string]string
	if err := json.Unmarshal(out, &v); err != nil {
		t.Fatalf("expected valid JSON, got: %s", string(out))
	}
	if v["goVersion"] == "" {
		t.Fatal("expected goVersion in JSON output")
	}
}
