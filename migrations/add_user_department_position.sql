-- Add department_id and position columns to users table
-- Migration: Add organizational information fields with proper foreign key relationship

-- Add department_id column as foreign key to departments table
ALTER TABLE users ADD COLUMN IF NOT EXISTS department_id INTEGER;

-- Add position column for job title
ALTER TABLE users ADD COLUMN IF NOT EXISTS position VARCHAR(100);

-- Create foreign key constraint
ALTER TABLE users ADD CONSTRAINT fk_users_department 
    FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE SET NULL;

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_users_department_id ON users(department_id);
CREATE INDEX IF NOT EXISTS idx_users_position ON users(position);

-- Add comments for documentation
COMMENT ON COLUMN users.department_id IS 'Foreign key to departments table - user organizational department';
COMMENT ON COLUMN users.position IS 'User job position or title (e.g., Software Engineer, Manager)';
