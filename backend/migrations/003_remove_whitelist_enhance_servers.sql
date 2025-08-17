-- Drop IP whitelist table as we no longer need it
DROP TABLE IF EXISTS ip_whitelist CASCADE;

-- Enhance servers table with better structure
ALTER TABLE servers 
ADD COLUMN IF NOT EXISTS api_key VARCHAR(255) UNIQUE,
ADD COLUMN IF NOT EXISTS description TEXT,
ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true,
ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id),
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW();

-- Add index for API key lookups
CREATE INDEX IF NOT EXISTS idx_servers_api_key ON servers(api_key) WHERE is_active = true;

-- Add index for active servers
CREATE INDEX IF NOT EXISTS idx_servers_active ON servers(is_active);

-- Create a function to generate unique API keys
CREATE OR REPLACE FUNCTION generate_api_key() RETURNS TEXT AS $$
BEGIN
    RETURN 'srv_' || encode(gen_random_bytes(32), 'hex');
END;
$$ LANGUAGE plpgsql;

-- Update existing servers with API keys if they don't have them
UPDATE servers 
SET api_key = generate_api_key(),
    updated_at = NOW()
WHERE api_key IS NULL;