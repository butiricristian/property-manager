CREATE TABLE inventory_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    category TEXT NOT NULL CHECK (category IN ('linens', 'kitchenware', 'furniture', 'electronics', 'toiletries', 'appliances', 'other')),
    quantity INT NOT NULL DEFAULT 1,
    acquired_date DATE NOT NULL,
    lifetime_months INT NOT NULL DEFAULT 12,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inventory_items_property_id ON inventory_items(property_id);
