-- Add role column to users table
ALTER TABLE users ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'user';

-- Add check constraint to ensure role is either 'user' or 'admin'
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('user', 'admin'));

-- Create index on role column for better query performance
CREATE INDEX idx_users_role ON users(role);
