ALTER TABLE expenses DROP CONSTRAINT expenses_category_check;
ALTER TABLE expenses ADD CONSTRAINT expenses_category_check
    CHECK (category IN ('electricity', 'heating', 'water', 'internet', 'maintenance', 'cleaning', 'other', 'partner_payout'));
