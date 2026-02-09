-- +goose Up

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE projects (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text UNIQUE NOT NULL,
  description text DEFAULT '',
  created_at timestamptz DEFAULT now()
);

CREATE TABLE links (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  url text NOT NULL,
  title text DEFAULT '',
  user_notes text DEFAULT '',
  generated_notes text DEFAULT '',
  generated_notes_size text DEFAULT 'short',
  project_id uuid REFERENCES projects(id),
  tags text[] DEFAULT '{}',
  cart boolean DEFAULT false,
  created_at timestamptz DEFAULT now(),
  updated_at timestamptz DEFAULT now()
);

CREATE INDEX ON links(cart);
CREATE INDEX ON links USING GIN(tags);

ALTER TABLE links ADD COLUMN fts tsvector GENERATED ALWAYS AS (
  setweight(to_tsvector('english', coalesce(title,'')), 'A') ||
  setweight(to_tsvector('english', coalesce(user_notes,'')), 'B') ||
  setweight(to_tsvector('english', coalesce(generated_notes,'')), 'C')
) STORED;

CREATE INDEX ON links USING GIN(fts);

INSERT INTO projects(name) VALUES ('Pedals') ON CONFLICT DO NOTHING;

-- +goose Down

DROP TABLE links;
DROP TABLE projects;
