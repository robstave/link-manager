-- +goose Up

-- This migration will be run by the API on startup to ensure admin user exists
-- For now, just create the function that the API will call

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION ensure_user_defaults(user_id uuid) RETURNS void AS $$
DECLARE
  default_project_id uuid;
  default_category_id uuid;
BEGIN
  -- Create default project if it doesn't exist
  INSERT INTO projects (owner_id, name, description, is_default, display_order)
  VALUES (user_id, 'Default', 'Default project', true, 0)
  ON CONFLICT (owner_id, name) DO NOTHING
  RETURNING id INTO default_project_id;

  -- Get the project id if it already existed
  IF default_project_id IS NULL THEN
    SELECT id INTO default_project_id 
    FROM projects 
    WHERE owner_id = user_id AND is_default = true;
  END IF;

  -- Create default category if it doesn't exist
  INSERT INTO categories (project_id, name, is_default, display_order)
  VALUES (default_project_id, 'Unsorted', true, 0)
  ON CONFLICT (project_id, name) DO NOTHING
  RETURNING id INTO default_category_id;

  -- Get the category id if it already existed
  IF default_category_id IS NULL THEN
    SELECT id INTO default_category_id 
    FROM categories 
    WHERE project_id = default_project_id AND is_default = true;
  END IF;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down

DROP FUNCTION IF EXISTS ensure_user_defaults(uuid);
