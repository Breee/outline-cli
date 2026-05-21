---
title: "Environment Variables"
weight: 65
description: "All environment variables supported by outline-cli."
llmsDescription: "outline-cli environment variables: OUTLINE_HOST or OUTLINE_SERVER_URL (server URL, OUTLINE_HOST takes priority), OUTLINE_API_TOKEN (API token for authentication), OUTLINE_USERNAME (basic auth username), OUTLINE_PASSWORD (basic auth password), OUTLINE_OIDC_ACCESS_TOKEN (OIDC access token, rarely set manually), OUTLINE_AUTH_METHOD (auto-auth method: oidc|api-token|basic). All override config file values. CLI flags override env vars."
---

# Environment Variables

All environment variables recognized by outline-cli.

## Authentication

| Variable | Description |
|----------|-------------|
| `OUTLINE_API_TOKEN` | API token for Outline |
| `OUTLINE_USERNAME` | Basic auth username |
| `OUTLINE_PASSWORD` | Basic auth password |
| `OUTLINE_OIDC_ACCESS_TOKEN` | OIDC access token (set automatically by `auth oidc-login`) |

## Server

| Variable | Description |
|----------|-------------|
| `OUTLINE_HOST` | Outline server URL (preferred in CI/CD) |
| `OUTLINE_SERVER_URL` | Outline server URL (alternative) |

!!! note
    If both `OUTLINE_HOST` and `OUTLINE_SERVER_URL` are set, `OUTLINE_SERVER_URL` takes priority (it maps directly to the `--server-url` flag).

## Behavior

| Variable | Description |
|----------|-------------|
| `OUTLINE_AUTH_METHOD` | Auto-auth method: `oidc`, `api-token`, or `basic` |

## Priority

```
CLI flags > Environment variables > OS keyring > Config file
```

## CI/CD Minimal Setup

Only two variables are needed for CI:

```bash
export OUTLINE_HOST=https://outline.example.com
export OUTLINE_API_TOKEN=sk-your-token
outline push --collection-id "Docs" --path ./docs/
```
