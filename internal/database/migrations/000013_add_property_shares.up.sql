CREATE TABLE property_shares (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission TEXT NOT NULL DEFAULT 'read',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (property_id, user_id)
);

CREATE INDEX idx_property_shares_user_id ON property_shares(user_id);
CREATE INDEX idx_property_shares_property_id ON property_shares(property_id);
