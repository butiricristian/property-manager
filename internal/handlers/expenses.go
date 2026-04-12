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

func (h *ExpenseHandler) verifyPropertyRead(w http.ResponseWriter, r *http.Request, propertyID pgtype.UUID) bool {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return false
	}
	if !verifyPropertyRead(r.Context(), h.queries, propertyID, userID) {
		respondError(w, http.StatusNotFound, "property not found")
		return false
	}
	return true
}

func (h *ExpenseHandler) verifyPropertyWrite(w http.ResponseWriter, r *http.Request, propertyID pgtype.UUID) bool {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return false
	}
	if !verifyPropertyWrite(r.Context(), h.queries, propertyID, userID) {
		respondError(w, http.StatusNotFound, "property not found")
		return false
	}
	return true
}

func (h *ExpenseHandler) ListByProperty(w http.ResponseWriter, r *http.Request) {
	propertyID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid property id")
		return
	}

	if !h.verifyPropertyRead(w, r, propertyID) {
		return
	}

	startDate, endDate, err := getDateRangeParams(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid date range")
		return
	}

	expenses, err := h.queries.ListExpensesByProperty(r.Context(), sqlc.ListExpensesByPropertyParams{
		PropertyID: propertyID,
		StartDate:  startDate,
		EndDate:    endDate,
	})
	if err != nil {
		respondInternalError(w, "failed to list expenses", err)
		return
	}
	respondJSON(w, http.StatusOK, expenses)
}

type createExpenseRequest struct {
	Category     string  `json:"category"`
	Description  string  `json:"description"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	Date         string  `json:"date"`
	Recurring    bool    `json:"recurring"`
	IsPercentage bool    `json:"is_percentage"`
	Percentage   float64 `json:"percentage"`
}

func (h *ExpenseHandler) Create(w http.ResponseWriter, r *http.Request) {
	propertyID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid property id")
		return
	}

	if !h.verifyPropertyWrite(w, r, propertyID) {
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
		PropertyID:   propertyID,
		Category:     req.Category,
		Description:  req.Description,
		Amount:       decimalFromFloat(req.Amount),
		Currency:     req.Currency,
		Date:         pgtype.Date{Time: date, Valid: true},
		Recurring:    req.Recurring,
		IsPercentage: req.IsPercentage,
		Percentage:   decimalFromFloat(req.Percentage),
	})
	if err != nil {
		respondInternalError(w, "failed to create expense", err)
		return
	}
	respondJSON(w, http.StatusCreated, expense)
}

type updateExpenseRequest struct {
	Category     string  `json:"category"`
	Description  string  `json:"description"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	Date         string  `json:"date"`
	Recurring    bool    `json:"recurring"`
	IsPercentage bool    `json:"is_percentage"`
	Percentage   float64 `json:"percentage"`
}

func (h *ExpenseHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	expense, err := h.queries.GetExpense(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "expense not found")
		return
	}

	if !verifyPropertyWrite(r.Context(), h.queries, expense.PropertyID, userID) {
		respondError(w, http.StatusNotFound, "expense not found")
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

	updated, err := h.queries.UpdateExpense(r.Context(), sqlc.UpdateExpenseParams{
		ID:           id,
		Category:     req.Category,
		Description:  req.Description,
		Amount:       decimalFromFloat(req.Amount),
		Currency:     req.Currency,
		Date:         pgtype.Date{Time: date, Valid: true},
		Recurring:    req.Recurring,
		IsPercentage: req.IsPercentage,
		Percentage:   decimalFromFloat(req.Percentage),
	})
	if err != nil {
		respondError(w, http.StatusNotFound, "expense not found")
		return
	}
	respondJSON(w, http.StatusOK, updated)
}

func (h *ExpenseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	expense, err := h.queries.GetExpense(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "expense not found")
		return
	}

	if !verifyPropertyWrite(r.Context(), h.queries, expense.PropertyID, userID) {
		respondError(w, http.StatusNotFound, "expense not found")
		return
	}

	if err := h.queries.DeleteExpense(r.Context(), id); err != nil {
		respondInternalError(w, "failed to delete expense", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
