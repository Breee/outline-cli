---
title: "Environment Variables"
weight: 65
description: "All environment variables supported by outline-cli."
generated: true
---

All environment variables recognized by outline-cli.

| Variable | Config Key | Description |
|----------|-----------|-------------|
| `OUTLINE_SERVER_URL` | `server_url` | Outline server base URL |
| `OUTLINE_AUTH_METHOD` | `auth_method` | Auth method: api_token, oidc, basic |
| `OUTLINE_TOKEN_STORAGE` | `token_storage` | Secret storage backend: keyring, file |
| `OUTLINE_OIDC_PORT` | `oidc_port` | Local port for OIDC callback |
| `OUTLINE_API_TOKEN` | `api_token` | API bearer token |
| `OUTLINE_PASSWORD` | `password` | Basic auth password |
| `OUTLINE_USERNAME` | `username` | Basic auth username |
| `OUTLINE_OIDC_ACCESS_TOKEN` | `oidc_access_token` | OIDC access token (set by auth oidc-login) |

## Priority

```
CLI flags > Environment variables > OS keyring > Config file
```

## CI/CD Minimal Setup

Only two variables are needed:

```bash
export OUTLINE_SERVER_URL=https://outline.example.com
export OUTLINE_API_TOKEN=sk-your-token
outline push --collection-id "Docs" --path ./docs/
```
