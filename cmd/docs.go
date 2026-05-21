package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Breee/outline-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	docsOutputDir string
	docsDir       string
	docsSiteURL   string
)

func init() {
	docsCmd := &cobra.Command{
		Use:    "docs",
		Short:  "Generate documentation",
		Hidden: true,
	}

	genCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate all docs: CLI reference, llms.txt, LLM instructions",
		RunE:  runDocsGenerate,
	}
	genCmd.Flags().StringVarP(&docsOutputDir, "output", "o", "docs/content/commands", "Output directory for CLI reference")
	genCmd.Flags().StringVar(&docsDir, "docs-dir", "docs/content", "Docs content directory")
	genCmd.Flags().StringVar(&docsSiteURL, "site-url", "https://breee.github.io/outline-cli", "Site base URL for links")

	docsCmd.AddCommand(genCmd)
	rootCmd.AddCommand(docsCmd)
}

func runDocsGenerate(_ *cobra.Command, _ []string) error {
	// 1. Generate CLI reference from cobra
	fmt.Println("Generating CLI reference...")
	if err := generateCLIReference(); err != nil {
		return err
	}

	// 2. Generate llms.txt and llms-full.txt
	fmt.Println("Generating llms.txt...")
	if err := generateLLMsTxt(); err != nil {
		return err
	}

	// 3. Generate LLM tool instructions
	fmt.Println("Generating LLM tool instructions...")
	if err := generateLLMInstructions(); err != nil {
		return err
	}

	fmt.Println("Done.")
	return nil
}

// --- CLI Reference Generation ---

func generateCLIReference() error {
	if err := os.MkdirAll(docsOutputDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	rootCmd.DisableAutoGenTag = true
	// Use stable default so generated docs don't vary by environment
	rootCmd.PersistentFlags().Lookup("config").DefValue = "$HOME/.outline-cli/config.yaml"
	if err := doc.GenMarkdownTree(rootCmd, docsOutputDir); err != nil {
		return fmt.Errorf("generating docs: %w", err)
	}

	return addFrontmatter(docsOutputDir)
}

func addFrontmatter(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		base := strings.TrimSuffix(filepath.Base(path), ".md")
		title := strings.ReplaceAll(base, "_", " ")

		frontmatter := fmt.Sprintf(`---
title: "%s"
description: "CLI reference for %s"
llmsDescription: "Auto-generated CLI reference for %s. Contains usage, flags, and examples."
generated: true
---

`, title, title, title)

		return os.WriteFile(path, []byte(frontmatter+string(content)), 0o644)
	})
}

// --- llms.txt Generation ---

type docPage struct {
	title       string
	description string
	llmsDesc    string
	relPath     string
	content     string
}

func generateLLMsTxt() error {
	pages, err := collectPages(docsDir)
	if err != nil {
		return err
	}

	// llms.txt
	var llms strings.Builder
	llms.WriteString("# outline-cli\n\n")
	llms.WriteString("> A CLI for pushing markdown files to Outline wiki. Supports OIDC auth, API tokens, metadata headers, directory hierarchy, image upload, and OS keyring credential storage.\n\n")
	llms.WriteString("## Docs\n\n")
	for _, p := range pages {
		desc := p.llmsDesc
		if desc == "" {
			desc = p.description
		}
		url := docsSiteURL + "/" + strings.TrimSuffix(p.relPath, ".md") + "/"
		url = strings.TrimSuffix(url, "index/") // index.md -> directory root
		llms.WriteString(fmt.Sprintf("- [%s](%s): %s\n", p.title, url, desc))
	}

	if err := os.WriteFile("llms.txt", []byte(llms.String()), 0o644); err != nil {
		return err
	}

	// llms-full.txt
	var full strings.Builder
	full.WriteString("# outline-cli\n\n")
	full.WriteString("> A CLI for pushing markdown files to Outline wiki.\n\n")
	for _, p := range pages {
		full.WriteString("## " + p.title + "\n\n")
		full.WriteString(p.content + "\n\n---\n\n")
	}

	return os.WriteFile("llms-full.txt", []byte(full.String()), 0o644)
}

func collectPages(dir string) ([]docPage, error) {
	var pages []docPage
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(dir, path)
		text := string(content)
		fm := parseFrontmatter(text)
		body := stripFrontmatter(text)

		title := fm["title"]
		if title == "" {
			title = strings.TrimSuffix(filepath.Base(path), ".md")
		}

		pages = append(pages, docPage{
			title:       title,
			description: fm["description"],
			llmsDesc:    fm["llmsDescription"],
			relPath:     rel,
			content:     body,
		})
		return nil
	})
	return pages, err
}

func parseFrontmatter(content string) map[string]string {
	fm := make(map[string]string)
	if !strings.HasPrefix(content, "---") {
		return fm
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Scan() // skip opening ---
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			break
		}
		if idx := strings.Index(line, ":"); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			val = strings.Trim(val, `"'`)
			fm[key] = val
		}
	}
	return fm
}

func stripFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---") {
		return content
	}
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return content
	}
	return strings.TrimSpace(parts[2])
}

// --- LLM Tool Instructions ---

func generateLLMInstructions() error {
	instructions := buildInstructionContent()
	outputs := []string{".github/copilot-instructions.md", "CLAUDE.md", ".cursor/rules"}
	for _, path := range outputs {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(instructions), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func buildInstructionContent() string {
	data := instructionData{
		Targets: readMakefileTargets(),
		Options: config.Registry,
	}

	tmplPath := filepath.Join(filepath.Dir(docsDir), "templates", "instructions.md.tmpl")
	tmplContent, err := os.ReadFile(tmplPath)
	if err != nil {
		return fmt.Sprintf("# outline-cli\n\nError reading template %s: %v\n", tmplPath, err)
	}

	var buf bytes.Buffer
	tmpl := template.Must(template.New("instructions").Parse(string(tmplContent)))
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("# outline-cli\n\nError generating instructions: %v\n", err)
	}
	return buf.String()
}

type instructionData struct {
	Targets []makeTarget
	Options []config.Option
}

type makeTarget struct {
	Name    string
	Comment string
}

func readMakefileTargets() []makeTarget {
	content, err := os.ReadFile("Makefile")
	if err != nil {
		return nil
	}

	var targets []makeTarget
	inHelp := false
	for _, line := range strings.Split(string(content), "\n") {
		if strings.Contains(line, "help:") {
			inHelp = true
			continue
		}
		if inHelp {
			if !strings.HasPrefix(line, "\t") {
				break
			}
			trimmed := strings.TrimSpace(line)
			trimmed = strings.TrimPrefix(trimmed, `@echo "`)
			trimmed = strings.TrimSuffix(trimmed, `"`)
			if parts := strings.SplitN(strings.TrimSpace(trimmed), " - ", 2); len(parts) == 2 {
				targets = append(targets, makeTarget{Name: strings.TrimSpace(parts[0]), Comment: strings.TrimSpace(parts[1])})
			}
		}
	}
	return targets
}




