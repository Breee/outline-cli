<!-- Collection: test -->
<!-- Title: API Reference -->

# API Reference

## Authentication

All requests require a Bearer token:

```
Authorization: Bearer <token>
```

## Endpoints

### POST /api/documents.create

Create a new document.

**Request body:**
```json
{
  "collectionId": "uuid",
  "title": "My Document",
  "text": "# Markdown content",
  "publish": true
}
```

**Response:**
```json
{
  "ok": true,
  "data": {
    "id": "document-uuid",
    "title": "My Document",
    "text": "# Markdown content"
  }
}
```

### POST /api/collections.list

List all collections accessible to the authenticated user.

**Request body:**
```json
{}
```
