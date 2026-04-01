DROP INDEX IF EXISTS idx_properties_user_id;

ALTER TABLE properties DROP COLUMN IF EXISTS user_id;

DROP TABLE IF EXISTS users;
