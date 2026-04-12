-- name: ListTaxRulesByProperty :many
SELECT * FROM tax_rules
WHERE property_id = $1
ORDER BY created_at DESC;

-- name: ListTaxRulesForUser :many
SELECT tr.* FROM tax_rules tr
WHERE tr.property_id IN (
    SELECT prp.id FROM properties prp WHERE prp.user_id = @user_id
    UNION
    SELECT pps.property_id FROM property_shares pps WHERE pps.user_id = @user_id
);

-- name: GetTaxRule :one
SELECT * FROM tax_rules WHERE id = $1;

-- name: CreateTaxRule :one
INSERT INTO tax_rules (property_id, name, type, percentage, fixed_amount, currency)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateTaxRule :one
UPDATE tax_rules
SET name = $2, type = $3, percentage = $4, fixed_amount = $5, currency = $6
WHERE id = $1
RETURNING *;

-- name: DeleteTaxRule :exec
DELETE FROM tax_rules WHERE id = $1;
