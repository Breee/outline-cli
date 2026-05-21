# Feature: MCP Server Mode (Read-Only)

## Summary

`outline mcp` starts a Model Context Protocol (MCP) server that exposes your Outline wiki as read-only tools for AI assistants (Copilot, Claude, Cursor, etc.).

## Usage

```bash
outline mcp --server-url https://wiki.example.com --api-token <token>
```

Starts an MCP server on stdio (default) or SSE transport. AI tools can then search, read, and list documents without write access.

## MCP Tools Exposed

| Tool | Description |
|------|-------------|
| `search` | Full-text search across documents |
| `get_document` | Get document content by ID or title |
| `list_collections` | List all collections |
| `list_documents` | List documents in a collection |
| `get_collection` | Get collection metadata |

## Design

- **Read-only**: No create/update/delete operations exposed
- **Transport**: stdio (default, for editor integration) or `--transport sse` for remote
- **Auth**: Same auth stack as other commands (flags > env > keyring > config)
- **Protocol**: MCP spec (JSON-RPC over stdio/SSE)

## Architecture

```
cmd/mcp.go          — cobra command, starts server
internal/mcp/       — MCP server implementation
  server.go         — JSON-RPC handler, tool registry
  tools.go          — tool implementations (search, get, list)
  transport.go      — stdio / SSE transport
```

## Dependencies

- `github.com/mark3labs/mcp-go` — Go MCP SDK (handles protocol, tool registration)

## Configuration

Uses existing `config.Registry` options (`server_url`, `api_token`, etc.). No new config keys needed.

## Flags

```
--transport string   MCP transport: stdio (default), sse
--port int           Port for SSE transport (default 3333)
```

## Example: VS Code Integration

```jsonc
// .vscode/mcp.json
{
  "servers": {
    "outline": {
      "command": "outline",
      "args": ["mcp"],
      "env": {
        "OUTLINE_SERVER_URL": "https://wiki.example.com",
        "OUTLINE_API_TOKEN": "${input:outlineToken}"
      }
    }
  }
}
```

## Example: Claude Desktop

```jsonc
// claude_desktop_config.json
{
  "mcpServers": {
    "outline": {
      "command": "outline",
      "args": ["mcp"]
    }
  }
}
```

## Non-Goals

- No write operations (push stays as a separate command with explicit intent)
- No streaming/subscription (documents are static reads)
- No pagination in tool responses (Outline API handles limits internally)
