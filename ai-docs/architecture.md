<!-- Collection: test -->
<!-- Title: Architecture Overview -->

# Architecture Overview

This document describes the high-level architecture of the outline-cli tool.

## Components

- **CLI (cmd/)**: Cobra command handlers for push, auth, etc.
- **OIDC (internal/oidc/)**: OAuth2/OIDC browser login flow.
- **Outline Client (internal/outline/)**: HTTP client for the Outline API.
- **Config (internal/config/)**: Persistent configuration storage.

## Auth Flow

1. User runs `outline auth oidc-login`
2. CLI discovers Outline's OAuth2 server metadata
3. CLI registers a public OAuth client via Dynamic Client Registration
4. CLI starts local HTTP server and opens browser to Outline's `/oauth/authorize`
5. User authenticates (via Dex/OIDC) and authorizes the CLI
6. Outline redirects to local callback with authorization code
7. CLI exchanges code for access token using PKCE
8. Token is stored in `~/.config/outline-cli/config.json`
