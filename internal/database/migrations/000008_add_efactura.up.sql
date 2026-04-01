CREATE TABLE seller_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    legal_name TEXT NOT NULL,
    street TEXT NOT NULL DEFAULT '',
    city TEXT NOT NULL DEFAULT '',
    county TEXT NOT NULL DEFAULT '',
    postal_code TEXT NOT NULL DEFAULT '',
    country_code TEXT NOT NULL DEFAULT 'RO',
    cui TEXT NOT NULL DEFAULT '',
    cnp TEXT NOT NULL DEFAULT '',
    trade_register_nr TEXT NOT NULL DEFAULT '',
    vat_registered BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    invoice_number INT NOT NULL,
    invoice_series TEXT NOT NULL DEFAULT 'PM',
    issue_date DATE NOT NULL,
    due_date DATE NOT NULL,

    -- Seller snapshot
    seller_legal_name TEXT NOT NULL,
    seller_cui TEXT NOT NULL DEFAULT '',
    seller_cnp TEXT NOT NULL DEFAULT '',
    seller_trade_register_nr TEXT NOT NULL DEFAULT '',
    seller_street TEXT NOT NULL DEFAULT '',
    seller_city TEXT NOT NULL DEFAULT '',
    seller_county TEXT NOT NULL DEFAULT '',
    seller_postal_code TEXT NOT NULL DEFAULT '',
    seller_country_code TEXT NOT NULL DEFAULT 'RO',
    seller_vat_registered BOOLEAN NOT NULL DEFAULT false,

    -- Buyer snapshot
    buyer_legal_name TEXT NOT NULL,
    buyer_cui TEXT NOT NULL DEFAULT '',
    buyer_trade_register_nr TEXT NOT NULL DEFAULT '',
    buyer_street TEXT NOT NULL DEFAULT '',
    buyer_city TEXT NOT NULL DEFAULT '',
    buyer_county TEXT NOT NULL DEFAULT '',
    buyer_postal_code TEXT NOT NULL DEFAULT '',
    buyer_country_code TEXT NOT NULL DEFAULT 'RO',
    buyer_vat_registered BOOLEAN NOT NULL DEFAULT false,

    -- Amounts
    currency TEXT NOT NULL DEFAULT 'RON',
    net_amount DECIMAL(10,2) NOT NULL,
    tax_amount DECIMAL(10,2) NOT NULL,
    gross_amount DECIMAL(10,2) NOT NULL,
    tax_percent DECIMAL(5,2) NOT NULL,

    description TEXT NOT NULL DEFAULT '',
    nights INT NOT NULL DEFAULT 1,
    xml_content TEXT NOT NULL DEFAULT '',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (user_id, invoice_number)
);

CREATE INDEX idx_invoices_user_id ON invoices(user_id);
CREATE INDEX idx_invoices_booking_id ON invoices(booking_id);
