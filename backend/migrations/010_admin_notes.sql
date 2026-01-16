-- Add admin notes field for internal admin use
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS admin_notes TEXT;

-- Add comment for documentation
COMMENT ON COLUMN tenants.admin_notes IS 'Internal notes visible only to admins, not tenant users';
