# Link Manager — Overview

## Project Context

Self-hosted, multi-user link manager for research-heavy projects (electronics, guitar pedals, synthesizers, software architecture). Designed for users who accumulate large numbers of links and need lightweight structure without heavy ceremony.

**Design Philosophy**: Capture now, understand later. Links should be fast to capture with minimal friction and enriched later.

**Inspiration**: "Not quite Notion" — markdown-first workflow inspired by Obsidian, but backed by a real API and database so content is not tied to a single machine.

---

## Core Concepts

### Hierarchy

```
User → Projects → Categories → Links
         ↓
       Tags (cross-project)
```

- **Users**: All data is exclusive per user. Bob only sees Bob's data.
- **Projects**: Top-level folders (e.g., "Electronics", "Pedals", "Synths")
- **Categories**: Project-scoped folders (e.g., "Schematics", "To Build", "Shopping")
- **Links**: Always belong to exactly one project + category
- **Tags**: User-scoped, cross all projects (e.g., "delay" tag can be on links in any project)

### Default Structures

- Every user has a **Default Project** (landing page)
- Every project has a **Default Category** (e.g., "Unsorted")
- New links without explicit assignment go to default project + default category

---

## Features

### V1 Core

| Feature | Description |
|---------|-------------|
| Multi-user auth | Username/password, user data isolation |
| RBAC | Admins can create/reset users |
| Projects + Categories | Hierarchical organization |
| Tags | Cross-project classification |
| Stars (1-10) | Per-link rating for sorting |
| Click tracking | Count + last clicked timestamp |
| Favicon fetching | Auto-fetch and store icon URL |
| User notes | Manual markdown notes |
| Generated notes | LLM-generated summaries (on-demand) |
| Full-text search | PostgreSQL tsvector across title, notes |
| Cart | Boolean flag for curated working subset |
| Export | JSON + Obsidian markdown bundle |

### V2+ Future

- Semantic/vector search (pgvector)
- RAG-style queries
- OAuth/SSO
- Link sharing
- Browser extension

---

## Stack

| Layer | Technology |
|-------|------------|
| Backend | Go + pgx (no ORM) |
| Database | PostgreSQL + pgvector |
| Migrations | Goose (pure SQL) |
| Frontend | Static HTML/JS (responsive) |
| Containers | Docker Compose |

---

## Non-Goals (V1)

- Link sharing between users
- Public/anonymous access
- Mobile native apps
- Browser extension
- Real-time collaboration
