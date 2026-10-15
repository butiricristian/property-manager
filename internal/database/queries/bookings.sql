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
    net_amount, currency, status, price, estimated_gross_amount
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING *;

-- name: UpsertBooking :one
INSERT INTO bookings (
    property_id, platform, platform_reference, guest_name,
    check_in, check_out, nights, gross_amount, platform_fee,
    net_amount, currency, status, price, estimated_gross_amount
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
ON CONFLICT (platform, platform_reference) DO UPDATE SET
    guest_name = EXCLUDED.guest_name,
    check_in = EXCLUDED.check_in,
    check_out = EXCLUDED.check_out,
    nights = EXCLUDED.nights,
    gross_amount = EXCLUDED.gross_amount,
    platform_fee = EXCLUDED.platform_fee,
    net_amount = EXCLUDED.net_amount,
    currency = EXCLUDED.currency,
    status = EXCLUDED.status,
    price = EXCLUDED.price,
    estimated_gross_amount = COALESCE(EXCLUDED.estimated_gross_amount, bookings.estimated_gross_amount)
RETURNING *;

-- name: DeleteBooking :exec
DELETE FROM bookings WHERE id = $1;

-- name: GetBookingsSummaryByCurrency :many
SELECT
    currency,
    COALESCE(SUM(gross_amount), 0)::DECIMAL(10,2) AS total_income,
    COALESCE(SUM(price), 0)::DECIMAL(10,2) AS total_price,
    COALESCE(SUM(platform_fee), 0)::DECIMAL(10,2) AS platform_fees,
    COALESCE(SUM(net_amount), 0)::DECIMAL(10,2) AS net_income,
    COUNT(*)::INT AS bookings_count,
    COALESCE(SUM(nights), 0)::INT AS total_nights
FROM bookings
WHERE property_id = $1
  AND status != 'cancelled'
  AND check_in >= @start_date::date
  AND check_in < @end_date::date
GROUP BY currency;

-- name: GetAllBookingsSummaryByCurrency :many
SELECT
    currency,
    COALESCE(SUM(gross_amount), 0)::DECIMAL(10,2) AS total_income,
    COALESCE(SUM(price), 0)::DECIMAL(10,2) AS total_price,
    COALESCE(SUM(platform_fee), 0)::DECIMAL(10,2) AS platform_fees,
    COALESCE(SUM(net_amount), 0)::DECIMAL(10,2) AS net_income,
    COUNT(*)::INT AS bookings_count,
    COALESCE(SUM(nights), 0)::INT AS total_nights
FROM bookings
WHERE status != 'cancelled'
  AND check_in >= @start_date::date
  AND check_in < @end_date::date
GROUP BY currency;

-- name: GetAllBookingsSummaryByCurrencyForUser :many
SELECT
    b.property_id,
    b.currency,
    COALESCE(SUM(b.gross_amount), 0)::DECIMAL(10,2) AS total_income,
    COALESCE(SUM(b.price), 0)::DECIMAL(10,2) AS total_price,
    COALESCE(SUM(b.platform_fee), 0)::DECIMAL(10,2) AS platform_fees,
    COALESCE(SUM(b.net_amount), 0)::DECIMAL(10,2) AS net_income,
    COUNT(*)::INT AS bookings_count,
    COALESCE(SUM(b.nights), 0)::INT AS total_nights
FROM bookings b
WHERE b.property_id IN (
    SELECT prp.id FROM properties prp WHERE prp.user_id = @user_id
    UNION
    SELECT pps.property_id FROM property_shares pps WHERE pps.user_id = @user_id
  )
  AND b.status != 'cancelled'
  AND b.check_in >= @start_date::date
  AND b.check_in < @end_date::date
GROUP BY b.property_id, b.currency;

-- name: GetBookingAnalyticsByProperty :one
SELECT
    COALESCE(AVG(b.check_in - b.created_at::date), 0)::FLOAT8 AS avg_lead_time_days,
    COUNT(*) FILTER (WHERE b.status = 'cancelled')::INT AS cancelled_count,
    COUNT(*)::INT AS total_count
FROM bookings b
WHERE b.property_id = $1
  AND b.check_in >= @start_date::date
  AND b.check_in < @end_date::date;

-- name: GetBookingAnalyticsForUser :one
SELECT
    COALESCE(AVG(b.check_in - b.created_at::date), 0)::FLOAT8 AS avg_lead_time_days,
    COUNT(*) FILTER (WHERE b.status = 'cancelled')::INT AS cancelled_count,
    COUNT(*)::INT AS total_count
FROM bookings b
WHERE b.property_id IN (
    SELECT prp.id FROM properties prp WHERE prp.user_id = @user_id
    UNION
    SELECT pps.property_id FROM property_shares pps WHERE pps.user_id = @user_id
  )
  AND b.check_in >= @start_date::date
  AND b.check_in < @end_date::date;
