<!-- Collection: test -->
<!-- Title: Feature: CI/CD Integration & Test Infrastructure -->

# Feature: CI/CD Integration & Test Infrastructure

## Why This Matters

For `outline-cli` to be a serious tool, it needs:
1. **Confidence in changes** — comprehensive tests that catch regressions
2. **Easy adoption** — users need ready-made CI/CD templates to publish docs from their repos
3. **Quality gates** — lint, validate, and check docs before they hit the wiki

This feature covers both internal testing infrastructure AND user-facing CI/CD integration.

---

## Part 1: Internal Test Infrastructure

### Current State

- Unit tests: `cmd/push_test.go`, `internal/config/*_test.go`, `internal/outline/client_test.go`
- E2E tests: `test/e2e/` with mock Outline server
- CI: `release.yaml` runs `make test` and `make e2e`

### What's Missing

| Gap | Impact |
|-----|--------|
| No test coverage tracking | Don't know what's untested |
| No integration tests for OIDC flow | Auth bugs only caught manually |
| No test for config set/get keyring round-trip | Keyring bugs missed |
| No golden file tests for push output | Formatting regressions unnoticed |
| No race detector in CI | Concurrency bugs |
| No linting (golangci-lint) | Style drift, potential bugs |
| No PR workflow (only release) | Broken code can merge |

### New: `.github/workflows/ci.yaml`

Runs on every push and PR to main:

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  contents: read

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version: "1.25.0"
      - uses: golangci/golangci-lint-action@v7
        with:
          version: latest

  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version: "1.25.0"
      - run: go test -race -coverprofile=coverage.out ./...
      - uses: codecov/codecov-action@v5
        with:
          files: coverage.out

  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version: "1.25.0"
      - run: make e2e
```

### New Tests to Write

| Test | Package | Type |
|------|---------|------|
| Push with metadata headers (all combinations) | `cmd` | Unit |
| Push directory tree ordering | `cmd` | Unit |
| Image upload and URL rewrite | `cmd` | Unit |
| Collection resolution (name, slug, UUID) | `internal/outline` | Unit |
| Config set/get all keys round-trip | `internal/config` | Unit |
| OIDC browser flow state machine | `internal/oidc` | Unit |
| Auto-auth triggers when token expired | `cmd` | Integration |
| Push with `--create-collection` | E2E | E2E |
| Push same file twice (upsert) | E2E | E2E |
| Push directory with index.md parent | E2E | E2E |
| OUTLINE_SERVER_URL env var resolves | `cmd` | Unit |
| Keyring fallback to file | `internal/config` | Unit |

### Golden File Tests

For commands that produce structured output, use golden files:

```go
func TestPushOutput(t *testing.T) {
    got := captureOutput(pushCmd, args)
    golden := filepath.Join("testdata", t.Name()+".golden")
    if *update {
        os.WriteFile(golden, got, 0644)
    }
    want, _ := os.ReadFile(golden)
    if diff := cmp.Diff(string(want), string(got)); diff != "" {
        t.Fatalf("output mismatch (-want +got):\n%s", diff)
    }
}
```

### `.golangci.yml`

```yaml
linters:
  enable:
    - errcheck
    - govet
    - staticcheck
    - unused
    - ineffassign
    - gosimple
    - gocritic
    - revive
    - misspell

issues:
  exclude-dirs:
    - test/e2e/mock
```

---

## Part 2: User-Facing CI/CD Integration

### GitHub Action: `outline-cli/push-action`

A reusable GitHub Action for users to publish docs from their repos:

```yaml
# In user's repo: .github/workflows/docs.yaml
name: Publish Docs

on:
  push:
    branches: [main]
    paths: ['docs/**']

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: Breee/outline-cli/action@v1
        with:
          server-url: ${{ vars.OUTLINE_URL }}
          api-token: ${{ secrets.OUTLINE_API_TOKEN }}
          collection: "Engineering"
          path: ./docs/
          create-collection: true
```

### Action Implementation (`action.yml`)

```yaml
name: 'Outline Push'
description: 'Push markdown files to Outline wiki'
inputs:
  server-url:
    description: 'Outline server URL'
    required: true
  api-token:
    description: 'Outline API token'
    required: true
  collection:
    description: 'Target collection (name, slug, or UUID)'
    required: true
  path:
    description: 'Path to markdown files'
    required: true
    default: './docs/'
  create-collection:
    description: 'Create collection if it does not exist'
    default: 'false'
  publish:
    description: 'Publish created documents'
    default: 'true'
runs:
  using: 'composite'
  steps:
    - name: Install outline-cli
      shell: bash
      run: |
        curl -sL "https://github.com/Breee/outline-cli/releases/latest/download/outline_linux_amd64.tar.gz" | tar xz -C /usr/local/bin outline
    - name: Push docs
      shell: bash
      env:
        OUTLINE_SERVER_URL: ${{ inputs.server-url }}
        OUTLINE_API_TOKEN: ${{ inputs.api-token }}
      run: |
        outline push \
          --collection-id "${{ inputs.collection }}" \
          --path "${{ inputs.path }}" \
          --publish=${{ inputs.publish }} \
          --create-collection=${{ inputs.create-collection }}
```

### GitLab CI Template

```yaml
# Include in .gitlab-ci.yml:
# include:
#   - remote: 'https://raw.githubusercontent.com/Breee/outline-cli/main/ci/gitlab-template.yml'

.outline-push:
  image: alpine:latest
  before_script:
    - apk add --no-cache curl
    - curl -sL "https://github.com/Breee/outline-cli/releases/latest/download/outline_linux_amd64.tar.gz" | tar xz -C /usr/local/bin
  script:
    - outline push --collection-id "$OUTLINE_COLLECTION" --path "$OUTLINE_DOCS_PATH"
  variables:
    OUTLINE_DOCS_PATH: ./docs/
```

### Doc Validation in CI (Pre-push Checks)

```bash
# Validate docs before pushing (future command)
outline lint --path ./docs/

# Checks:
# - All metadata headers are valid
# - Referenced collections exist
# - No broken internal links (between docs in same push)
# - Images referenced actually exist locally
# - No empty documents
# - Title uniqueness within collection
```

---

## Part 3: Implementation Steps

1. **Create `.github/workflows/ci.yaml`** — PR/push workflow with lint, test (multi-OS), e2e
2. **Add `.golangci.yml`** — linter config
3. **Write missing unit tests** (see table above)
4. **Add `-race` flag** to `make test`
5. **Add coverage reporting** (codecov or coveralls)
6. **Create `action.yml`** — GitHub Action for users
7. **Create `ci/gitlab-template.yml`** — GitLab CI template
8. **Add `outline lint` command** — pre-push validation (future)

## Testing the CI/CD Feature Itself

- Test that the GitHub Action installs and runs correctly (workflow_call test)
- Test that `outline lint` catches expected validation errors
- Verify coverage doesn't regress (set minimum threshold)
