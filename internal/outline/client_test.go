package outline

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAuthConfig(t *testing.T) {
	t.Parallel()

	tokenAuth, err := NewAuthConfig("token", "", "", "")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if tokenAuth.HeaderValue != "Bearer token" {
		t.Fatalf("unexpected token header: %s", tokenAuth.HeaderValue)
	}

	basicAuth, err := NewAuthConfig("", "user", "pass", "")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	wantBasic := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass"))
	if basicAuth.HeaderValue != wantBasic {
		t.Fatalf("unexpected basic header: %s", basicAuth.HeaderValue)
	}

	_, err = NewAuthConfig("", "", "", "")
	if err == nil {
		t.Fatal("expected error for missing auth")
	}
}

func TestCreateDocument(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		if r.URL.Path != "/api/documents.create" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":   true,
			"data": map[string]any{"id": "doc-1", "title": "hello", "text": "# hello"},
		})
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL, AuthConfig{HeaderValue: "Bearer test"}, ts.Client())
	if err != nil {
		t.Fatal(err)
	}
	doc, err := client.CreateDocument(context.Background(), "col-1", "hello", "# hello", true, DocumentOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if doc.ID != "doc-1" {
		t.Fatalf("unexpected id: %s", doc.ID)
	}
}
