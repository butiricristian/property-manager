ALTER TABLE properties ADD COLUMN airbnb_retains_partner_payout BOOLEAN NOT NULL DEFAULT false;

UPDATE properties p
SET airbnb_retains_partner_payout = true
WHERE EXISTS (
    SELECT 1 FROM expenses e
    WHERE e.property_id = p.id
      AND e.category = 'partner_payout'
      AND e.is_percentage
      AND NOT 'airbnb' = ANY(e.income_platforms)
);

ALTER TABLE expenses DROP COLUMN income_platforms;
