ALTER TABLE expenses ADD COLUMN is_percentage BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE expenses ADD COLUMN percentage DECIMAL(5,2) NOT NULL DEFAULT 0;

ALTER TABLE expenses DROP CONSTRAINT expenses_category_check;
ALTER TABLE expenses ADD CONSTRAINT expenses_category_check
    CHECK (category IN ('electricity', 'heating', 'water', 'maintenance', 'cleaning', 'other', 'partner_payout'));
