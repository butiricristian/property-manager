DELETE FROM expenses WHERE category = 'partner_payout';

ALTER TABLE expenses DROP CONSTRAINT expenses_category_check;
ALTER TABLE expenses ADD CONSTRAINT expenses_category_check
    CHECK (category IN ('electricity', 'heating', 'water', 'maintenance', 'cleaning', 'other'));

ALTER TABLE expenses DROP COLUMN percentage;
ALTER TABLE expenses DROP COLUMN is_percentage;
