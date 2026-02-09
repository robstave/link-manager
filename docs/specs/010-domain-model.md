# Link Manager — Domain Model

## Entity Relationship

```
┌─────────┐       ┌──────────┐       ┌────────────┐       ┌─────────┐
│  users  │───1:N─│ projects │───1:N─│ categories │───1:N─│  links  │
└─────────┘       └──────────┘       └────────────┘       └─────────┘
     │                                                         │
     └────────────────────1:N──────────────────────────────────┤
                                                               │
                           ┌──────┐                            │
                           │ tags │◄───────────N:M─────────────┘
                           └──────┘
```

---

## Tables

### users

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | uuid | PK | |
| username | text | UNIQUE NOT NULL | Login identifier |
| password_hash | text | NOT NULL | bcrypt hash |
| email | text | UNIQUE | Optional |
| role | text | NOT NULL DEFAULT 'user' | 'user' or 'admin' |
| created_at | timestamptz | NOT NULL DEFAULT now() | |
| updated_at | timestamptz | NOT NULL DEFAULT now() | |

---

### projects

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | uuid | PK | |
| owner_id | uuid | FK users(id) NOT NULL | Data isolation |
| name | text | NOT NULL | |
| description | text | | |
| is_default | boolean | NOT NULL DEFAULT false | One per user |
| display_order | int | NOT NULL DEFAULT 0 | Sidebar ordering |
| created_at | timestamptz | NOT NULL DEFAULT now() | |
| updated_at | timestamptz | NOT NULL DEFAULT now() | |

**Constraints**:
- UNIQUE(owner_id, name)
- Only one project with is_default=true per owner_id

---

### categories

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | uuid | PK | |
| project_id | uuid | FK projects(id) NOT NULL | |
| name | text | NOT NULL | |
| is_default | boolean | NOT NULL DEFAULT false | One per project |
| display_order | int | NOT NULL DEFAULT 0 | |
| created_at | timestamptz | NOT NULL DEFAULT now() | |
| updated_at | timestamptz | NOT NULL DEFAULT now() | |

**Constraints**:
- UNIQUE(project_id, name)
- Only one category with is_default=true per project_id
- ON DELETE CASCADE from projects

---

### links

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | uuid | PK | |
| owner_id | uuid | FK users(id) NOT NULL | Denormalized for fast queries |
| project_id | uuid | FK projects(id) NOT NULL | |
| category_id | uuid | FK categories(id) NOT NULL | |
| url | text | NOT NULL | |
| title | text | | Display name |
| description | text | | Short description |
| icon_url | text | | Fetched favicon URL |
| user_notes | text | | Markdown |
| generated_notes | text | | LLM-generated markdown |
| generated_notes_size | text | | tiny\|short\|medium\|long |
| stars | int | CHECK (stars >= 0 AND stars <= 10) | 0 = unrated |
| click_count | int | NOT NULL DEFAULT 0 | |
| last_clicked_at | timestamptz | | |
| cart | boolean | NOT NULL DEFAULT false | |
| created_at | timestamptz | NOT NULL DEFAULT now() | |
| updated_at | timestamptz | NOT NULL DEFAULT now() | |
| fts | tsvector | GENERATED | For full-text search |

**Indexes**:
- owner_id (all queries filter by owner)
- project_id
- category_id
- GIN(fts)
- (owner_id, stars DESC) for sorted queries
- (owner_id, click_count DESC) for popular sorting

---

### tags

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | uuid | PK | |
| owner_id | uuid | FK users(id) NOT NULL | User-scoped |
| name | text | NOT NULL | |
| color | text | | Optional hex color |
| created_at | timestamptz | NOT NULL DEFAULT now() | |

**Constraints**:
- UNIQUE(owner_id, name)

---

### link_tags (junction table)

| Column | Type | Constraints |
|--------|------|-------------|
| link_id | uuid | FK links(id) ON DELETE CASCADE |
| tag_id | uuid | FK tags(id) ON DELETE CASCADE |

**Constraints**:
- PRIMARY KEY (link_id, tag_id)

---

## Full-Text Search

The `fts` column is a generated tsvector:

```sql
fts tsvector GENERATED ALWAYS AS (
  setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
  setweight(to_tsvector('english', coalesce(description, '')), 'B') ||
  setweight(to_tsvector('english', coalesce(user_notes, '')), 'C') ||
  setweight(to_tsvector('english', coalesce(generated_notes, '')), 'D')
) STORED
```

---

## Seed Data

On user creation, automatically create:

1. **Default Project**: name="Default", is_default=true
2. **Default Category**: name="Unsorted", is_default=true

For system initialization (first admin user), seed with sample links:
- www.google.com
- www.github.com
- www.msn.com
