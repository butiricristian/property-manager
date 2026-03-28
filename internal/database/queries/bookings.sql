-- name: ListBookingsByProperty :many
SELECT * FROM bookings
WHERE property_id = $1
  AND check_in >= @start_date::date
  AND check_in < @end_date::date
ORDER BY check_in DESC;

-- name: GetBooking :one
SELECT * FROM bookings WHERE id = $1;

-- name: CreateBooking :one
INSERT INTO bookings (
    property_id, platform, platform_reference, guest_name,
    check_in, check_out, nights, gross_amount, platform_fee,
    net_amount, currency, status
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: UpsertBooking :one
INSERT INTO bookings (
    property_id, platform, platform_reference, guest_name,
    check_in, check_out, nights, gross_amount, platform_fee,
    net_amount, currency, status
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (platform, platform_reference) DO UPDATE SET
    guest_name = EXCLUDED.guest_name,
    check_in = EXCLUDED.check_in,
    check_out = EXCLUDED.check_out,
    nights = EXCLUDED.nights,
    gross_amount = EXCLUDED.gross_amount,
    platform_fee = EXCLUDED.platform_fee,
    net_amount = EXCLUDED.net_amount,
    currency = EXCLUDED.currency,
    status = EXCLUDED.status
RETURNING *;

-- name: DeleteBooking :exec
DELETE FROM bookings WHERE id = $1;

-- name: GetBookingsSummary :one
SELECT
    COALESCE(SUM(gross_amount), 0)::DECIMAL(10,2) AS total_income,
    COALESCE(SUM(platform_fee), 0)::DECIMAL(10,2) AS platform_fees,
    COALESCE(SUM(net_amount), 0)::DECIMAL(10,2) AS net_income,
    COUNT(*)::INT AS bookings_count,
    COALESCE(SUM(nights), 0)::INT AS total_nights
FROM bookings
WHERE property_id = $1
  AND status != 'cancelled'
  AND check_in >= @start_date::date
  AND check_in < @end_date::date;

-- name: GetAllBookingsSummary :one
SELECT
    COALESCE(SUM(gross_amount), 0)::DECIMAL(10,2) AS total_income,
    COALESCE(SUM(platform_fee), 0)::DECIMAL(10,2) AS platform_fees,
    COALESCE(SUM(net_amount), 0)::DECIMAL(10,2) AS net_income,
    COUNT(*)::INT AS bookings_count,
    COALESCE(SUM(nights), 0)::INT AS total_nights
FROM bookings
WHERE status != 'cancelled'
  AND check_in >= @start_date::date
  AND check_in < @end_date::date;
