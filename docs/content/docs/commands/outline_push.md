---
title: "outline push"
description: "CLI reference for outline push"
llmsDescription: "Auto-generated CLI reference for outline push. Contains usage, flags, and examples."
generated: true
---

## outline push

Push markdown files to Outline

```
outline push [flags]
```

### Examples

```
  # Push a single file
  outline push --path ./README.md --collection-id "Engineering"

  # Push a directory tree
  outline push --path ./docs/ --collection-id "Docs"

  # Create the collection if it doesn't exist
  outline push --path ./docs/ --collection-id "New Docs" --create-collection
```

### Options

```
      --collection-id string   Default Outline collection (name, slug, or UUID)
      --create-collection      Create collection if it does not exist
      --diff                   Show content diff for changed documents
  -h, --help                   help for push
  -p, --path string            Path to markdown file or directory (default ".")
      --publish                Publish created documents (default true)
  -y, --yes                    Skip confirmation prompt
```

### Options inherited from parent commands

```
      --api-token string           Outline API token
      --config string              config file (default "$HOME/.outline-cli/config.yaml")
      --oidc-access-token string   OIDC access token
      --password string            Basic auth password
      --server-url string          Outline server URL
      --username string            Basic auth username
```

### SEE ALSO

* [outline](outline.md)	 - Push markdown documents to Outline

