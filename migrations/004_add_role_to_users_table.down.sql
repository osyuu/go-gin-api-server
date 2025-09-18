-- Remove role column from users table
-- First drop the index
DROP INDEX IF EXISTS idx_users_role;

-- Then drop the constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;

-- Finally drop the column
ALTER TABLE users DROP COLUMN IF EXISTS role;
