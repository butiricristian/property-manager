-- name: GetSellerProfile :one
SELECT * FROM seller_profiles WHERE user_id = $1;

-- name: UpsertSellerProfile :one
INSERT INTO seller_profiles (
    user_id, legal_name, street, city, county, postal_code, country_code,
    cui, cnp, trade_register_nr, vat_registered
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (user_id) DO UPDATE SET
    legal_name = EXCLUDED.legal_name,
    street = EXCLUDED.street,
    city = EXCLUDED.city,
    county = EXCLUDED.county,
    postal_code = EXCLUDED.postal_code,
    country_code = EXCLUDED.country_code,
    cui = EXCLUDED.cui,
    cnp = EXCLUDED.cnp,
    trade_register_nr = EXCLUDED.trade_register_nr,
    vat_registered = EXCLUDED.vat_registered,
    updated_at = NOW()
RETURNING *;
