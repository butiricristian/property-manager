package handlers

import (
	"fmt"
	"net/http"
	"slices"
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
	Category        string   `json:"category"`
	Description     string   `json:"description"`
	Amount          float64  `json:"amount"`
	Currency        string   `json:"currency"`
	Date            string   `json:"date"`
	EndDate         string   `json:"end_date"`
	Recurring       bool     `json:"recurring"`
	IsPercentage    bool     `json:"is_percentage"`
	Percentage      float64  `json:"percentage"`
	IncomePlatforms []string `json:"income_platforms"`
}

// allIncomePlatforms is the default income base of a percentage expense.
var allIncomePlatforms = []string{"booking", "airbnb"}

// parseIncomePlatforms reads the platforms a percentage partner payout is computed
// on. Every other expense always counts all income.
func parseIncomePlatforms(selected []string, category string, isPercentage bool) ([]string, error) {
	if category != "partner_payout" || !isPercentage {
		return allIncomePlatforms, nil
	}
	for _, platform := range selected {
		if !slices.Contains(allIncomePlatforms, platform) {
			return nil, fmt.Errorf("unknown income platform %q", platform)
		}
	}
	// Rebuilt from the known list so a repeated platform can't be counted twice.
	platforms := slices.DeleteFunc(slices.Clone(allIncomePlatforms), func(platform string) bool {
		return !slices.Contains(selected, platform)
	})
	if len(platforms) == 0 {
		return nil, fmt.Errorf("select at least one income platform")
	}
	return platforms, nil
}

// parseEndDate reads the optional last-active day of a recurring expense. An
// empty value means the expense runs indefinitely.
func parseEndDate(s string, start time.Time) (pgtype.Date, error) {
	if s == "" {
		return pgtype.Date{}, nil
	}
	end, err := time.Parse("2006-01-02", s)
	if err != nil {
		return pgtype.Date{}, fmt.Errorf("invalid end date format")
	}
	if end.Before(start) {
		return pgtype.Date{}, fmt.Errorf("end date must be on or after the start date")
	}
	return pgtype.Date{Time: end, Valid: true}, nil
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

	endDate, err := parseEndDate(req.EndDate, date)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Currency == "" {
		req.Currency = "EUR"
	}

	incomePlatforms, err := parseIncomePlatforms(req.IncomePlatforms, req.Category, req.IsPercentage)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	expense, err := h.queries.CreateExpense(r.Context(), sqlc.CreateExpenseParams{
		PropertyID:      propertyID,
		Category:        req.Category,
		Description:     req.Description,
		Amount:          decimalFromFloat(req.Amount),
		Currency:        req.Currency,
		Date:            pgtype.Date{Time: date, Valid: true},
		EndDate:         endDate,
		Recurring:       req.Recurring,
		IsPercentage:    req.IsPercentage,
		Percentage:      decimalFromFloat(req.Percentage),
		IncomePlatforms: incomePlatforms,
	})
	if err != nil {
		respondInternalError(w, "failed to create expense", err)
		return
	}
	respondJSON(w, http.StatusCreated, expense)
}

type updateExpenseRequest struct {
	Category        string   `json:"category"`
	Description     string   `json:"description"`
	Amount          float64  `json:"amount"`
	Currency        string   `json:"currency"`
	Date            string   `json:"date"`
	EndDate         string   `json:"end_date"`
	Recurring       bool     `json:"recurring"`
	IsPercentage    bool     `json:"is_percentage"`
	Percentage      float64  `json:"percentage"`
	IncomePlatforms []string `json:"income_platforms"`
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

	endDate, err := parseEndDate(req.EndDate, date)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Currency == "" {
		req.Currency = "EUR"
	}

	incomePlatforms, err := parseIncomePlatforms(req.IncomePlatforms, req.Category, req.IsPercentage)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.queries.UpdateExpense(r.Context(), sqlc.UpdateExpenseParams{
		ID:              id,
		Category:        req.Category,
		Description:     req.Description,
		Amount:          decimalFromFloat(req.Amount),
		Currency:        req.Currency,
		Date:            pgtype.Date{Time: date, Valid: true},
		EndDate:         endDate,
		Recurring:       req.Recurring,
		IsPercentage:    req.IsPercentage,
		Percentage:      decimalFromFloat(req.Percentage),
		IncomePlatforms: incomePlatforms,
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
