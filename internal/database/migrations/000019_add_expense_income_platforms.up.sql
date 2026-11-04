ALTER TABLE expenses
    ADD COLUMN income_platforms TEXT[] NOT NULL DEFAULT '{booking,airbnb}'
    CHECK (cardinality(income_platforms) > 0 AND income_platforms <@ ARRAY['booking', 'airbnb']);

UPDATE expenses e
SET income_platforms = '{booking}'
FROM properties p
WHERE p.id = e.property_id
  AND p.airbnb_retains_partner_payout
  AND e.category = 'partner_payout'
  AND e.is_percentage;

ALTER TABLE properties DROP COLUMN airbnb_retains_partner_payout;
