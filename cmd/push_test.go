package cmd

import "testing"

func TestDocumentTitle(t *testing.T) {
	t.Parallel()

	if got := documentTitle("/tmp/example.md", "# My Title\ncontent"); got != "My Title" {
		t.Fatalf("unexpected markdown title: %s", got)
	}
	if got := documentTitle("/tmp/another-file.md", "no heading"); got != "another-file" {
		t.Fatalf("unexpected fallback title: %s", got)
	}
}
