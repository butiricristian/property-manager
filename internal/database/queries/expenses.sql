-- name: ListExpensesByProperty :many
SELECT * FROM expenses
WHERE property_id = $1
  AND (
    (date >= @start_date::date AND date < @end_date::date)
    OR (recurring = true AND date < @end_date::date)
  )
ORDER BY date DESC;

-- name: GetExpense :one
SELECT * FROM expenses WHERE id = $1;

-- name: CreateExpense :one
INSERT INTO expenses (property_id, category, description, amount, currency, date, recurring, is_percentage, percentage)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateExpense :one
UPDATE expenses
SET category = $2, description = $3, amount = $4, currency = $5, date = $6, recurring = $7, is_percentage = $8, percentage = $9
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
  AND e.is_percentage = false
GROUP BY e.category, e.currency;

-- name: GetPercentageExpensesByProperty :many
SELECT * FROM expenses
WHERE property_id = $1
  AND (
    (date >= @start_date::date AND date < @end_date::date)
    OR (recurring = true AND date < @end_date::date)
  )
  AND is_percentage = true;

-- name: GetAllPercentageExpenses :many
SELECT * FROM expenses
WHERE (
    (date >= @start_date::date AND date < @end_date::date)
    OR (recurring = true AND date < @end_date::date)
  )
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
  AND e.is_percentage = true;
