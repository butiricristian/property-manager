-- name: ListExpensesByProperty :many
SELECT * FROM expenses
WHERE property_id = $1
  AND (
    (date >= @start_date::date AND date < @end_date::date)
    OR (recurring = true AND date < @end_date::date)
  )
  AND (expenses.end_date IS NULL OR expenses.end_date >= @start_date::date)
ORDER BY date DESC;

-- name: GetExpense :one
SELECT * FROM expenses WHERE id = $1;

-- name: CreateExpense :one
INSERT INTO expenses (property_id, category, description, amount, currency, date, recurring, is_percentage, percentage, end_date, income_platforms)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: UpdateExpense :one
UPDATE expenses
SET category = $2, description = $3, amount = $4, currency = $5, date = $6, recurring = $7, is_percentage = $8, percentage = $9, end_date = $10, income_platforms = $11
WHERE id = $1
RETURNING *;

-- name: DeleteExpense :exec
DELETE FROM expenses WHERE id = $1;

-- name: GetExpensesByCategoryCurrency :many
SELECT
    category,
    currency,
    COALESCE(SUM(amount), 0)::DECIMAL(10,2) AS total
FROM expenses
WHERE property_id = $1
  AND (
    (date >= @start_date::date AND date < @end_date::date)
    OR (recurring = true AND date < @end_date::date)
  )
  AND (expenses.end_date IS NULL OR expenses.end_date >= @start_date::date)
  AND is_percentage = false
GROUP BY category, currency;

-- name: GetAllExpensesByCategoryCurrency :many
SELECT
    category,
    currency,
    COALESCE(SUM(amount), 0)::DECIMAL(10,2) AS total
FROM expenses
WHERE (
    (date >= @start_date::date AND date < @end_date::date)
    OR (recurring = true AND date < @end_date::date)
  )
  AND (expenses.end_date IS NULL OR expenses.end_date >= @start_date::date)
  AND is_percentage = false
GROUP BY category, currency;

-- name: GetAllExpensesByCategoryCurrencyForUser :many
SELECT
    e.category,
    e.currency,
    COALESCE(SUM(e.amount), 0)::DECIMAL(10,2) AS total
FROM expenses e
WHERE e.property_id IN (
    SELECT prp.id FROM properties prp WHERE prp.user_id = @user_id
    UNION
    SELECT pps.property_id FROM property_shares pps WHERE pps.user_id = @user_id
  )
  AND (
    (e.date >= @start_date::date AND e.date < @end_date::date)
    OR (e.recurring = true AND e.date < @end_date::date)
  )
  AND (e.end_date IS NULL OR e.end_date >= @start_date::date)
  AND e.is_percentage = false
GROUP BY e.category, e.currency;

-- name: GetPercentageExpensesByProperty :many
SELECT * FROM expenses
WHERE property_id = $1
  AND (
    (date >= @start_date::date AND date < @end_date::date)
    OR (recurring = true AND date < @end_date::date)
  )
  AND (expenses.end_date IS NULL OR expenses.end_date >= @start_date::date)
  AND is_percentage = true;

-- name: GetAllPercentageExpenses :many
SELECT * FROM expenses
WHERE (
    (date >= @start_date::date AND date < @end_date::date)
    OR (recurring = true AND date < @end_date::date)
  )
  AND (expenses.end_date IS NULL OR expenses.end_date >= @start_date::date)
  AND is_percentage = true;

-- name: GetAllPercentageExpensesForUser :many
SELECT e.* FROM expenses e
WHERE e.property_id IN (
    SELECT prp.id FROM properties prp WHERE prp.user_id = @user_id
    UNION
    SELECT pps.property_id FROM property_shares pps WHERE pps.user_id = @user_id
  )
  AND (
    (e.date >= @start_date::date AND e.date < @end_date::date)
    OR (e.recurring = true AND e.date < @end_date::date)
  )
  AND (e.end_date IS NULL OR e.end_date >= @start_date::date)
  AND e.is_percentage = true;

-- A percentage expense only counts bookings checking out while it was active: a
-- recurring one from its start date through its end date, a one-off one during
-- its calendar month.
-- name: GetPercentageExpenseNetIncome :many
SELECT
    e.id AS expense_id,
    b.currency,
    COALESCE(SUM(b.net_amount), 0)::DECIMAL(10,2) AS net_income
FROM expenses e
JOIN bookings b ON b.property_id = e.property_id
    AND b.platform = ANY(e.income_platforms)
    AND b.status != 'cancelled'
    AND b.check_out >= GREATEST(
        @start_date::date,
        CASE WHEN e.recurring THEN e.date ELSE date_trunc('month', e.date)::date END
    )
    AND b.check_out < LEAST(
        @end_date::date,
        CASE WHEN e.recurring THEN COALESCE(e.end_date + 1, @end_date::date)
             ELSE (date_trunc('month', e.date) + INTERVAL '1 month')::date END
    )
WHERE e.id = ANY(@expense_ids::uuid[])
GROUP BY e.id, b.currency;

-- Net income covered by at least one of the given expenses, so income under two
-- overlapping expenses is only counted once.
-- name: GetNetIncomeCoveredByExpenses :many
SELECT
    b.currency,
    COALESCE(SUM(b.net_amount), 0)::DECIMAL(10,2) AS net_income
FROM bookings b
WHERE EXISTS (
    SELECT 1 FROM expenses e
    WHERE e.id = ANY(@expense_ids::uuid[])
      AND b.property_id = e.property_id
        AND b.platform = ANY(e.income_platforms)
        AND b.status != 'cancelled'
        AND b.check_out >= GREATEST(
            @start_date::date,
            CASE WHEN e.recurring THEN e.date ELSE date_trunc('month', e.date)::date END
        )
        AND b.check_out < LEAST(
            @end_date::date,
            CASE WHEN e.recurring THEN COALESCE(e.end_date + 1, @end_date::date)
                 ELSE (date_trunc('month', e.date) + INTERVAL '1 month')::date END
        )
)
GROUP BY b.currency;
