-- Multi-tenant support for Karen
-- Adds tenants, users, and tenant phone numbers

-- Tenants (customers/organizations)
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,

    -- AI Configuration
    system_prompt TEXT NOT NULL,           -- Custom agent prompt
    greeting_text TEXT,                     -- Custom greeting
    voice_id TEXT,                          -- ElevenLabs voice ID
    language TEXT DEFAULT 'cs',             -- STT/TTS language

    -- Behavior settings
    vip_names TEXT[] DEFAULT '{}',          -- Names to forward immediately
    marketing_email TEXT,                   -- Email for marketing redirects
    forward_number TEXT,                    -- Number to forward urgent calls

    -- Subscription
    plan TEXT DEFAULT 'trial',              -- trial, basic, pro
    status TEXT DEFAULT 'active',           -- active, suspended, cancelled

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Users (authenticated via phone)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    phone TEXT UNIQUE NOT NULL,            -- E.164 format: +420777123456
    phone_verified BOOLEAN DEFAULT false,
    name TEXT,
    role TEXT DEFAULT 'owner',             -- owner, admin, member
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE INDEX idx_users_phone ON users(phone);

-- Phone numbers assigned to tenants (for incoming calls)
CREATE TABLE tenant_phone_numbers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    twilio_number TEXT UNIQUE NOT NULL,    -- E.164 format: +1234567890
    twilio_sid TEXT,                        -- Twilio Phone Number SID

    -- For forwarding detection fallback
    forwarding_source TEXT,                 -- User's original number

    is_primary BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_tenant_phone_numbers_tenant ON tenant_phone_numbers(tenant_id);
CREATE INDEX idx_tenant_phone_numbers_twilio ON tenant_phone_numbers(twilio_number);
CREATE INDEX idx_tenant_phone_numbers_forwarding ON tenant_phone_numbers(forwarding_source);

-- Add tenant_id to calls table
ALTER TABLE calls ADD COLUMN tenant_id UUID REFERENCES tenants(id);
CREATE INDEX idx_calls_tenant ON calls(tenant_id);

-- User sessions for JWT invalidation (optional but useful for logout)
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,             -- SHA256 of JWT
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_user_sessions_user ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token ON user_sessions(token_hash);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
