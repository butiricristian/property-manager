-- name: GetNextInvoiceNumber :one
SELECT COALESCE(MAX(invoice_number), 0) + 1 AS next_number
FROM invoices WHERE user_id = $1;

-- name: CreateInvoice :one
INSERT INTO invoices (
    user_id, booking_id, invoice_number, invoice_series,
    issue_date, due_date,
    seller_legal_name, seller_cui, seller_cnp, seller_trade_register_nr,
    seller_street, seller_city, seller_county, seller_postal_code, seller_country_code,
    seller_vat_registered,
    buyer_legal_name, buyer_cui, buyer_trade_register_nr,
    buyer_street, buyer_city, buyer_county, buyer_postal_code, buyer_country_code,
    buyer_vat_registered,
    currency, net_amount, tax_amount, gross_amount, tax_percent,
    description, nights, xml_content
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
    $17, $18, $19, $20, $21, $22, $23, $24, $25,
    $26, $27, $28, $29, $30,
    $31, $32, $33
)
RETURNING *;

-- name: GetInvoiceByBooking :one
SELECT * FROM invoices WHERE booking_id = $1;

-- name: GetInvoice :one
SELECT * FROM invoices WHERE id = $1;

-- name: ListInvoicesByUser :many
SELECT * FROM invoices WHERE user_id = $1 ORDER BY invoice_number DESC;

-- name: DeleteInvoice :exec
DELETE FROM invoices WHERE id = $1;
