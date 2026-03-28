-- name: ListExpensesByProperty :many
SELECT * FROM expenses
WHERE property_id = $1
  AND date >= @start_date::date
  AND date < @end_date::date
ORDER BY date DESC;

-- name: GetExpense :one
SELECT * FROM expenses WHERE id = $1;

-- name: CreateExpense :one
INSERT INTO expenses (property_id, category, description, amount, currency, date, recurring)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateExpense :one
UPDATE expenses
SET category = $2, description = $3, amount = $4, currency = $5, date = $6, recurring = $7
WHERE id = $1
RETURNING *;

-- name: DeleteExpense :exec
DELETE FROM expenses WHERE id = $1;

-- name: GetExpensesByCategory :many
SELECT
    category,
    COALESCE(SUM(amount), 0)::DECIMAL(10,2) AS total
FROM expenses
WHERE property_id = $1
  AND date >= @start_date::date
  AND date < @end_date::date
GROUP BY category;

-- name: GetAllExpensesByCategory :many
SELECT
    category,
    COALESCE(SUM(amount), 0)::DECIMAL(10,2) AS total
FROM expenses
WHERE date >= @start_date::date
  AND date < @end_date::date
GROUP BY category;
