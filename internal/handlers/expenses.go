package handlers

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/cristianbutiri/property-manager-v2/internal/database/sqlc"
)

type ExpenseHandler struct {
	queries *sqlc.Queries
}

func NewExpenseHandler(q *sqlc.Queries) *ExpenseHandler {
	return &ExpenseHandler{queries: q}
}

func (h *ExpenseHandler) ListByProperty(w http.ResponseWriter, r *http.Request) {
	propertyID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid property id")
		return
	}

	month := getMonthParam(r)
	startDate, endDate, err := parseMonthDates(month)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid month format, use YYYY-MM")
		return
	}

	expenses, err := h.queries.ListExpensesByProperty(r.Context(), sqlc.ListExpensesByPropertyParams{
		PropertyID: propertyID,
		StartDate:  startDate,
		EndDate:    endDate,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list expenses")
		return
	}
	respondJSON(w, http.StatusOK, expenses)
}

type createExpenseRequest struct {
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Date        string  `json:"date"`
	Recurring   bool    `json:"recurring"`
}

func (h *ExpenseHandler) Create(w http.ResponseWriter, r *http.Request) {
	propertyID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid property id")
		return
	}

	var req createExpenseRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid date format")
		return
	}

	if req.Currency == "" {
		req.Currency = "EUR"
	}

	expense, err := h.queries.CreateExpense(r.Context(), sqlc.CreateExpenseParams{
		PropertyID:  propertyID,
		Category:    req.Category,
		Description: req.Description,
		Amount:      decimalFromFloat(req.Amount),
		Currency:    req.Currency,
		Date:        pgtype.Date{Time: date, Valid: true},
		Recurring:   req.Recurring,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create expense")
		return
	}
	respondJSON(w, http.StatusCreated, expense)
}

type updateExpenseRequest struct {
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Date        string  `json:"date"`
	Recurring   bool    `json:"recurring"`
}

func (h *ExpenseHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	var req updateExpenseRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid date format")
		return
	}

	if req.Currency == "" {
		req.Currency = "EUR"
	}

	expense, err := h.queries.UpdateExpense(r.Context(), sqlc.UpdateExpenseParams{
		ID:          id,
		Category:    req.Category,
		Description: req.Description,
		Amount:      decimalFromFloat(req.Amount),
		Currency:    req.Currency,
		Date:        pgtype.Date{Time: date, Valid: true},
		Recurring:   req.Recurring,
	})
	if err != nil {
		respondError(w, http.StatusNotFound, "expense not found")
		return
	}
	respondJSON(w, http.StatusOK, expense)
}

func (h *ExpenseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	if err := h.queries.DeleteExpense(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete expense")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
