---
title: "Quick Start"
weight: 20
description: "Push your first document to Outline in under 2 minutes."
llmsDescription: "Quick start for outline-cli: 1) Set OUTLINE_SERVER_URL and OUTLINE_API_TOKEN env vars. 2) Run `outline push --collection-id <name-or-uuid> --path ./file.md` to push a single file. 3) Run `outline push --collection-id <name> --path ./docs/` to push a directory tree. Files are upserted by title. Use `--create-collection` flag to auto-create the target collection."
---


Push your first document to Outline in under 2 minutes.

## Prerequisites

- An Outline instance (self-hosted or cloud)
- An API token (Settings → API in Outline)

## 1. Configure credentials

```bash
export OUTLINE_SERVER_URL=https://outline.example.com
export OUTLINE_API_TOKEN=sk-your-api-token
```

{{< callout type="info" >}}
**For persistent config:**
```bash
outline config set server_url https://outline.example.com
outline config set api_token sk-your-api-token  # stored in OS keyring
```
{{< /callout >}}

## 2. Push a single file

```bash
echo '# Hello World
This is my first doc pushed from the terminal.' > hello.md

outline push --collection-id "Engineering" --path hello.md
```

Output:
```
created "Hello World" in collection "Engineering"
```

## 3. Push a directory

```bash
outline push --collection-id "Engineering" --path ./docs/
```

All `.md` files are pushed. `index.md` and `README.md` become parent documents. Subdirectories become nested children.

## 4. Verify

Check your Outline instance — the documents are live!

## What's next?

- [Authentication options]({{< relref "auth" >}}) — OIDC, API token, or basic auth
- [Metadata headers]({{< relref "metadata" >}}) — control title, collection, icon per file
- [CI/CD integration]({{< relref "cicd" >}}) — automate on every git push
