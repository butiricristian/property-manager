# Percentage-Based Expenses

## Context
The user splits net profit with a partner. They want to define expenses as a percentage of net income (booking revenue minus platform fees), rather than a fixed amount. For example: "Partner payout: 50% of net income."

## Approach

Add `is_percentage` boolean and `percentage` decimal fields to the expenses table. When `is_percentage=true`, the expense amount is computed dynamically as `percentage/100 × net_income`. Fixed expenses keep using the `amount` field as before.

Also add `"partner_payout"` as a new expense category since that's the primary use case.

### Dashboard calculation order
1. Compute net income from bookings (unchanged)
2. Sum fixed expenses (WHERE `is_percentage = false`) — existing query with filter added
3. Fetch percentage expenses separately, compute each as `percentage/100 × netIncome`
4. Combine fixed + percentage into ExpenseBreakdown
5. Profit = netIncome − totalExpenses

### Expense list behavior
- Fixed expenses: show amount as before
- Percentage expenses: `amount` column shows the percentage value; frontend displays it with a "%" suffix. The `is_percentage` flag tells the UI how to render it.

## Changes

### 1. New migration: `000002_percentage_expenses.up.sql` / `.down.sql`
- `ALTER TABLE expenses ADD COLUMN is_percentage BOOLEAN NOT NULL DEFAULT false;`
- `ALTER TABLE expenses ADD COLUMN percentage DECIMAL(5,2) NOT NULL DEFAULT 0;`
- Update category CHECK constraint to include `'partner_payout'`

### 2. `internal/database/queries/expenses.sql`
- Update `CreateExpense` and `UpdateExpense` to include `is_percentage` and `percentage`
- Modify `GetExpensesByCategoryCurrency` to filter `WHERE is_percentage = false`
- Add `GetPercentageExpensesByProperty` — returns percentage expenses for a property+date range
- Add `GetAllPercentageExpensesByCategoryCurrency` — same but across all properties
- Update `ListExpensesByProperty` to include new columns (already covered by `SELECT *`)

### 3. Regenerate sqlc → `internal/database/sqlc/`
- Run `make sqlc`

### 4. `internal/models/models.go`
- Add `IsPercentage bool` and `Percentage decimal.Decimal` to `Expense` struct
- Add `PartnerPayout decimal.Decimal` to `ExpenseBreakdown`

### 5. `internal/services/dashboard_service.go`
- Update `aggregateExpenses` / `aggregateAllExpenses` to accept `netIncome decimal.Decimal`
- After aggregating fixed expenses, fetch percentage expenses
- For each percentage expense: `computed = netIncome × percentage / 100`
- Add computed amounts into the category breakdown
- Update `setCategoryDecimal` to handle `"partner_payout"`

### 6. `internal/handlers/expenses.go`
- Update `createExpenseRequest` / `updateExpenseRequest`: add `IsPercentage bool` and `Percentage float64`
- Pass new fields through to sqlc `CreateExpense` / `UpdateExpense`
- Validation: if `is_percentage=true`, require `percentage > 0`; if false, require `amount > 0`

### 7. `internal/handlers/helpers.go`
- Ensure `decimalFromFloat` is available (already exists)

### 8. `web/src/types/index.ts`
- Add `is_percentage` and `percentage` to `Expense` interface
- Add `partner_payout` to `ExpenseCategory` union
- Add `partner_payout` to `ExpenseBreakdown`
- Add `is_percentage` and `percentage` to `ExtractedExpense`

### 9. `web/src/pages/ExpensesPage.tsx`
- Add `"partner_payout"` to `CATEGORIES` array
- Add toggle in expense form: "Percentage of net income" checkbox
- When checked: show percentage input instead of amount+currency inputs
- Expense table: render amount column as `"50%"` for percentage expenses, normal amount for fixed
- Update `emptyForm` type to include `is_percentage` and `percentage`
- Update `handleSubmit` and `handleEdit` to pass/read new fields

### 10. `internal/services/bill_extract_service.go`
- Add `partner_payout` to `validCategories` map

## Files
| File | Change |
|------|--------|
| `internal/database/migrations/000002_percentage_expenses.up.sql` | **New** — add columns + update constraint |
| `internal/database/migrations/000002_percentage_expenses.down.sql` | **New** — rollback |
| `internal/database/queries/expenses.sql` | Update queries for new fields + add percentage queries |
| `internal/database/sqlc/` | Regenerated |
| `internal/models/models.go` | Add fields to Expense + ExpenseBreakdown |
| `internal/services/dashboard_service.go` | Percentage expense computation in dashboard |
| `internal/services/bill_extract_service.go` | Add partner_payout category |
| `internal/handlers/expenses.go` | Update request structs + pass new fields |
| `web/src/types/index.ts` | Add new fields + category |
| `web/src/pages/ExpensesPage.tsx` | Form toggle, table display |

## Verification
1. `make migrate-up` — migration runs
2. `make sqlc` — regenerates without errors
3. `go build ./...` — compiles
4. `cd web && npx tsc --noEmit` — no type errors
5. Create a fixed expense → works as before
6. Create a percentage expense (e.g. partner_payout 50%) → shows "50%" in list
7. Dashboard shows computed amount for percentage expense (50% of net income)
8. Edit percentage expense → form pre-fills correctly with percentage toggle on
