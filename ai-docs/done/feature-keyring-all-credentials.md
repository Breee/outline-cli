<!-- Collection: test -->
<!-- Title: Feature: Keyring for All Credentials -->

# Feature: Keyring for All Credentials

## Motivation

Per [OWASP Secrets Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html):

- Environment variables are "generally accessible to all processes and may be included in logs or system dumps" (Section 5.1)
- Desktop CLIs should use "appropriate secure storage mechanisms" — the OS keyring (Section 6.5)
- Static tokens should be stored encrypted at rest, not in plaintext config files (Section 2.2)

Currently only the OIDC access token uses the OS keyring. API tokens set via `config set` are stored in plaintext YAML. This feature extends keyring storage to all persisted credentials.

## Scope

| Credential | Current Storage | Target Storage |
|---|---|---|
| OIDC access token | Keyring (with file fallback) | No change |
| API token (via `config set`) | Plaintext in `config.yaml` | Keyring (with file fallback) |
| Basic auth password (via `config set`) | Plaintext in `config.yaml` | Keyring (with file fallback) |

**Out of scope:** Credentials passed via flags/env vars remain ephemeral (not persisted). This is acceptable per OWASP for CI/pipeline use cases.

## Design

### Keyring Keys

Each credential gets a distinct keyring user under the existing `outline-cli` service:

| Keyring User | Credential |
|---|---|
| `oidc_access_token` | OIDC token (existing) |
| `api_token` | API token |
| `basic_password` | Basic auth password |

### Config Keys

Add `api_token` and `password` to `validKeys` in `internal/config/config.go`.

### Token Functions

Generalize `internal/config/token.go`:

1. Rename/refactor `SaveToken` → generic `SaveSecret(cfgPath, cfg, keyringUser, value)`.
2. Rename/refactor `LoadToken` → generic `LoadSecret(cfg, keyringUser, configField)`.
3. Rename/refactor `DeleteToken` → generic `DeleteSecret(cfg, keyringUser)`.
4. Keep existing `SaveToken`/`LoadToken`/`DeleteToken` as thin wrappers for backward compat.
5. Add `SaveAPIToken`, `LoadAPIToken`, `SavePassword`, `LoadPassword` wrappers.

### Config Set Integration

In `cmd/config.go`, intercept `config set api_token <value>` and `config set password <value>`:
- Route through `SaveAPIToken` / `SavePassword` instead of writing plaintext to YAML.
- On `config get api_token` / `config get password`, read from keyring.
- On `config list`, show `api_token: ***` (masked) if set in keyring.

### Auth Resolution

In `cmd/root.go` (or wherever credentials are resolved), add keyring lookup for API token:

```
Priority (highest wins):
1. --api-token flag
2. OUTLINE_API_TOKEN env var
3. Keyring (api_token)
4. Config file (fallback)
```

For server URL:

```
Priority (highest wins):
1. --server-url flag
2. OUTLINE_SERVER_URL env var (new alias)
3. OUTLINE_SERVER_URL env var
4. Config file
```

Same pattern for password.

### File Fallback

Same behavior as OIDC: if keyring is unavailable (headless, CI, containers), fall back to plaintext config with a warning on stderr.

## Implementation Steps

1. **Generalize token.go** — extract `SaveSecret`/`LoadSecret`/`DeleteSecret` accepting a keyring user string.
2. **Add wrapper functions** — `SaveAPIToken`, `LoadAPIToken`, `SavePassword`, `LoadPassword`.
3. **Update validKeys** — add `api_token`, `password` to the allowed config keys.
4. **Update config set/get** — route secret keys through keyring functions.
5. **Update config list** — mask secret values.
6. **Update credential resolution in root.go** — check keyring for api_token/password after flags/env.
7. **Add unit tests** — mock keyring (already used by existing tests).
8. **Update README** — document `config set api_token` stores in keyring.
9. **Document CI/CD usage** — show how to use env vars (`OUTLINE_API_TOKEN`, `OUTLINE_PASSWORD`) in pipelines where keyring is unavailable, and note that file fallback with `token_storage=file` is the alternative for headless environments.

## Default Credential Resolution

The primary use case for many users is CI/CD. The default resolution order should prioritize environment variables so that pipelines work out of the box with no config file or keyring:

```
Priority (highest wins):
1. --api-token / --server-url flags
2. OUTLINE_API_TOKEN / OUTLINE_SERVER_URL env vars
3. Keyring lookup
4. Config file (plaintext fallback)
```

This means a CI job only needs:

```yaml
env:
  OUTLINE_SERVER_URL: https://outline.example.com
  OUTLINE_API_TOKEN: ${{ secrets.OUTLINE_API_TOKEN }}
```

No config file, no keyring, no `auth oidc-login` step required.

**Note:** `OUTLINE_SERVER_URL` should be added as an alias for `OUTLINE_SERVER_URL` since "host" is the more natural term in CI/CD contexts. Both should be accepted.

## Security Notes

- Keyring access is scoped to the current user's session (libsecret/Keychain/Credential Manager).
- No secrets are ever logged; `config list` masks them.
- File fallback emits a warning: `"keyring unavailable, storing secret in plaintext config"`.
- The `config.yaml` file should be `0600` (already enforced by `Save()`).

## Status: Implemented

All implementation steps are complete. Tests are in:
- `internal/config/token_test.go` — 8 tests covering Save/Load/Delete for all credential types, file permissions, multi-secret independence, and error cases.
- `internal/config/config_test.go` — 8 tests covering ValidKeys, SecretKeys, Get/Set for new keys.

### Test Coverage Checklist

- [x] SaveAPIToken + LoadAPIToken round-trip (file storage)
- [x] SavePassword + LoadPassword round-trip (file storage)
- [x] SaveToken + LoadToken round-trip (backward compat)
- [x] Unknown storage returns error
- [x] Unknown storage in Load falls back to file field
- [x] DeleteSecret with file storage is no-op
- [x] Config file written with 0600 permissions
- [x] Multiple secrets stored independently
- [x] ValidKeys contains api_token and password
- [x] SecretKeys map identifies secret keys correctly
- [x] Set/Get api_token and password via File struct
- [x] Unknown key in Get/Set returns error
- [ ] E2E: `config set api_token` + `push` uses keyring token (manual test with dev env)
- [ ] E2E: `OUTLINE_SERVER_URL` env var resolves correctly (manual test)

