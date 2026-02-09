# Link Manager — API Contract

All endpoints require authentication unless noted. All responses are JSON.

**Base URL**: `/api/v1`

---

## Authentication

### POST /auth/login
Login and receive JWT token.

**Request**:
```json
{ "username": "bob", "password": "secret" }
```

**Response** (200):
```json
{ "token": "eyJ...", "expires_at": "2024-01-01T00:00:00Z" }
```

### POST /auth/logout
Invalidate current token.

### GET /auth/me
Get current user info.

**Response** (200):
```json
{ "id": "uuid", "username": "bob", "role": "user" }
```

---

## Admin (role=admin required)

### GET /admin/users
List all users.

### POST /admin/users
Create new user.
```json
{ "username": "alice", "password": "secret", "role": "user" }
```

### POST /admin/users/{id}/reset-password
Reset user password.
```json
{ "new_password": "newsecret" }
```

### DELETE /admin/users/{id}
Delete user and all their data.

---

## Projects

### GET /projects
List user's projects (ordered by display_order).

**Response**:
```json
[
  { "id": "uuid", "name": "Default", "is_default": true, "category_count": 3, "link_count": 42 },
  { "id": "uuid", "name": "Pedals", "is_default": false, "category_count": 5, "link_count": 128 }
]
```

### POST /projects
Create new project.
```json
{ "name": "Synths", "description": "Synthesizer research" }
```

### GET /projects/{id}
Get project with categories and link counts.

### PATCH /projects/{id}
Update project.

### DELETE /projects/{id}
Delete project (cascades to categories and links). Cannot delete default project.

### PUT /projects/{id}/reorder
Update display_order.

---

## Categories

### GET /projects/{project_id}/categories
List categories in project.

### POST /projects/{project_id}/categories
Create category.
```json
{ "name": "Schematics" }
```

### PATCH /categories/{id}
Update category.

### DELETE /categories/{id}
Delete category. Links move to project's default category.

---

## Links

### GET /links
List links with filtering, sorting, pagination.

**Query Parameters**:
| Param | Type | Description |
|-------|------|-------------|
| project_id | uuid | Filter by project |
| category_id | uuid | Filter by category |
| tag | string | Filter by tag name |
| cart | boolean | Filter cart items |
| q | string | Full-text search |
| sort | string | `stars`, `clicks`, `recent`, `created` (default: stars) |
| limit | int | Default 50, max 200 |
| offset | int | Pagination offset |

**Response**:
```json
{
  "links": [
    {
      "id": "uuid",
      "url": "https://example.com",
      "title": "Example",
      "description": "Short desc",
      "icon_url": "https://example.com/favicon.ico",
      "stars": 8,
      "click_count": 42,
      "last_clicked_at": "2024-01-01T00:00:00Z",
      "cart": false,
      "project": { "id": "uuid", "name": "Pedals" },
      "category": { "id": "uuid", "name": "Schematics" },
      "tags": ["fuzz", "delay"]
    }
  ],
  "total": 128,
  "limit": 50,
  "offset": 0
}
```

### POST /links
Create link.
```json
{
  "url": "https://example.com",
  "title": "Example Site",
  "description": "A great resource",
  "project_id": "uuid",
  "category_id": "uuid",
  "tags": ["fuzz"],
  "stars": 5
}
```

If `project_id` or `category_id` omitted, uses defaults.

### GET /links/{id}
Get full link details including notes.

### PATCH /links/{id}
Update link fields.

### DELETE /links/{id}
Delete link.

### POST /links/{id}/click
Record a click. Increments click_count, updates last_clicked_at. Returns redirect URL.

**Response** (200):
```json
{ "redirect_url": "https://example.com" }
```

### PATCH /links/{id}/stars
Update star rating.
```json
{ "stars": 8 }
```

### PATCH /links/{id}/cart
Toggle cart status.
```json
{ "cart": true }
```

### POST /links/{id}/move
Move link to different project/category.
```json
{ "project_id": "uuid", "category_id": "uuid" }
```

---

## Tags

### GET /tags
List user's tags with usage counts.

**Response**:
```json
[
  { "id": "uuid", "name": "fuzz", "color": "#ff0000", "link_count": 15 },
  { "id": "uuid", "name": "delay", "color": null, "link_count": 8 }
]
```

### POST /tags
Create tag.
```json
{ "name": "reverb", "color": "#00ff00" }
```

### PATCH /tags/{id}
Update tag.

### DELETE /tags/{id}
Delete tag (removes from all links).

### POST /links/{link_id}/tags
Add tags to link.
```json
{ "tags": ["fuzz", "new-tag"] }
```
Creates tags that don't exist.

### DELETE /links/{link_id}/tags/{tag_id}
Remove tag from link.

---

## Generated Notes

### POST /links/{id}/generate
Generate AI notes for link.

**Query Parameters**:
- `size`: tiny | short | medium | long (default: short)

**Response** (202 Accepted):
```json
{ "status": "processing" }
```

Or (200) if cached:
```json
{ "generated_notes": "## Summary\n...", "size": "short" }
```

---

## Export

### GET /export/links.json
Export all links as JSON.

**Query Parameters**: Same as GET /links for filtering.

### GET /export/cart.json
Export cart links as JSON.

### GET /export/obsidian.zip
Export as Obsidian-compatible markdown bundle.

**Query Parameters**: Same as GET /links for filtering.

Each file contains:
```markdown
---
id: uuid
url: "..."
project: "Pedals"
category: "Schematics"
tags: [fuzz, delay]
stars: 8
cart: true
created_at: "..."
---

# Title

## Description
...

## User Notes
...

## Generated Notes
...
```

---

## Favicon

### POST /links/{id}/fetch-icon
Attempt to fetch and store favicon URL.

**Response** (200):
```json
{ "icon_url": "https://example.com/favicon.ico" }
```

---

## Health

### GET /healthz (no auth)
Health check.

**Response** (200):
```json
{ "status": "ok", "db": "connected" }
```
