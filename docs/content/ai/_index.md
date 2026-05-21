---
title: "AI Integration"
description: "Machine-readable documentation for LLMs and AI coding assistants."
llmsDescription: "outline-cli provides llms.txt and llms-full.txt at the docs site root for AI consumption. llms.txt contains a structured index of all pages with one-line descriptions. llms-full.txt contains the full text of all documentation pages concatenated. AI assistants can fetch these to understand the CLI's capabilities, commands, and usage patterns without scraping individual pages."
---

# AI Integration

This documentation site is designed for both human readers and AI assistants.

## llms.txt

Available at [`/llms.txt`](/llms.txt) — a structured index of all documentation pages with concise descriptions, following the [llms.txt standard](https://llmstxt.org/).

AI assistants can fetch this single file to understand what outline-cli does and find relevant documentation.

## llms-full.txt

Available at [`/llms-full.txt`](/llms-full.txt) — the full text of all documentation concatenated into one file.

Use this when you need the complete documentation context (e.g., for code generation tasks involving outline-cli).

## For AI Coding Assistants

If you're an AI assistant helping a user with outline-cli:

1. **Quick reference**: Fetch `/llms.txt` for a page index
2. **Full context**: Fetch `/llms-full.txt` for all documentation
3. **Key facts**:
   - Push markdown to Outline: `outline push --path ./docs/ --collection-id "Name"`
   - Auth methods: OIDC (browser), API token, basic auth
   - Config: `~/.outline-cli/config.yaml`, secrets in OS keyring
   - CI/CD: Only needs `OUTLINE_SERVER_URL` + `OUTLINE_API_TOKEN` env vars
   - Metadata headers: `<!-- Collection: X -->`, `<!-- Title: X -->`, `<!-- Parent: X -->`
   - Directory `index.md` → parent doc, sibling `.md` → children

## Structured Data per Page

Every documentation page includes an `llmsDescription` in its frontmatter — a technical, command-syntax-aware summary designed for AI consumption. These are aggregated into `llms.txt`.
