# Link Manager — Technical & Product Specification

## Purpose
Self-hosted link manager for research-heavy projects (electronics, pedals, software).

Backend-first, markdown-first, SQL-first.

Features:
- Links with user notes
- AI-generated notes
- Full-text search
- Tags + projects
- Cart subset
- Export to JSON + Obsidian Markdown

Non-goals V1:
- Auth
- Multi-user
- Sharing

---

## Stack

Backend: Go  
DB: Postgres + pgvector (pgvector/pgvector:pg17-bookworm)  
Migrations: Goose (SQL)  
Frontend: Static HTML  
Containers: Docker Compose  

---

## Data Model

### projects
id uuid PK  
name text unique  
description text  
created_at timestamptz  

### links
id uuid PK  
url text  
title text  
user_notes text  
generated_notes text  
generated_notes_size text  
project_id uuid FK  
tags text[]  
cart boolean  
created_at timestamptz  
updated_at timestamptz  
fts tsvector generated  

Indexes:
- GIN(tags)
- GIN(fts)
- project_id
- cart

---

## API

GET /healthz

Projects:
GET /api/projects
POST /api/projects

Links:
POST /api/links
GET /api/links
PATCH /api/links/{id}
DELETE /api/links/{id}

Search:
GET /api/links?q=schematic

Generate notes:
POST /api/links/{id}/generate?size=short

Export:
GET /api/export/links.json
GET /api/export/cart.json
GET /api/export/obsidian.zip

---

## Generated Notes

V1 stub.
V2 fetch page -> extract text -> LLM summarize -> store markdown.
Later embeddings via pgvector.

---

## Obsidian Export Format

Each link becomes markdown:

---
id: uuid
url: "..."
tags: []
cart: true
---

# Title

## User notes
...

## Generated notes
...

---

## Milestones

1 compose works
2 migrations
3 CRUD
4 search
5 cart
6 export
7 generate stub
8 vectors

---

Acceptance:
docker compose up --build runs clean.
