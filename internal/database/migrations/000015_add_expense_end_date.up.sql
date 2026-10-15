ALTER TABLE expenses ADD COLUMN end_date DATE;

ALTER TABLE expenses ADD CONSTRAINT expenses_end_date_check
    CHECK (end_date IS NULL OR end_date >= date);
