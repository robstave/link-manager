-- +goose Up

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS vector;

-- Users table
CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  username text UNIQUE NOT NULL,
  password_hash text NOT NULL,
  email text UNIQUE,
  role text NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- Projects table
CREATE TABLE projects (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name text NOT NULL,
  description text DEFAULT '',
  is_default boolean NOT NULL DEFAULT false,
  display_order int NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(owner_id, name)
);

-- Ensure only one default project per user
CREATE UNIQUE INDEX idx_one_default_project_per_user 
  ON projects(owner_id) WHERE is_default = true;

-- Categories table
CREATE TABLE categories (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name text NOT NULL,
  is_default boolean NOT NULL DEFAULT false,
  display_order int NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(project_id, name)
);

-- Ensure only one default category per project
CREATE UNIQUE INDEX idx_one_default_category_per_project 
  ON categories(project_id) WHERE is_default = true;

-- Links table
CREATE TABLE links (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  category_id uuid NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  url text NOT NULL,
  title text DEFAULT '',
  description text DEFAULT '',
  icon_url text DEFAULT '',
  user_notes text DEFAULT '',
  generated_notes text DEFAULT '',
  generated_notes_size text DEFAULT 'short' CHECK (generated_notes_size IN ('tiny', 'short', 'medium', 'long')),
  stars int DEFAULT 0 CHECK (stars >= 0 AND stars <= 10),
  click_count int NOT NULL DEFAULT 0,
  last_clicked_at timestamptz,
  cart boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- Indexes for links
CREATE INDEX idx_links_owner ON links(owner_id);
CREATE INDEX idx_links_project ON links(project_id);
CREATE INDEX idx_links_category ON links(category_id);
CREATE INDEX idx_links_cart ON links(owner_id, cart) WHERE cart = true;
CREATE INDEX idx_links_stars ON links(owner_id, stars DESC);
CREATE INDEX idx_links_clicks ON links(owner_id, click_count DESC);

-- Full-text search
ALTER TABLE links ADD COLUMN fts tsvector GENERATED ALWAYS AS (
  setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
  setweight(to_tsvector('english', coalesce(description, '')), 'B') ||
  setweight(to_tsvector('english', coalesce(user_notes, '')), 'C') ||
  setweight(to_tsvector('english', coalesce(generated_notes, '')), 'D')
) STORED;

CREATE INDEX idx_links_fts ON links USING GIN(fts);

-- Tags table
CREATE TABLE tags (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name text NOT NULL,
  color text,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(owner_id, name)
);

-- Link-Tags junction table
CREATE TABLE link_tags (
  link_id uuid NOT NULL REFERENCES links(id) ON DELETE CASCADE,
  tag_id uuid NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY (link_id, tag_id)
);

CREATE INDEX idx_link_tags_tag ON link_tags(tag_id);

-- +goose Down

DROP TABLE link_tags;
DROP TABLE tags;
DROP TABLE links;
DROP TABLE categories;
DROP TABLE projects;
DROP TABLE users;
