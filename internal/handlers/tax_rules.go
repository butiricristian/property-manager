package handlers

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/cristianbutiri/property-manager-v2/internal/database/sqlc"
)

type TaxRuleHandler struct {
	queries *sqlc.Queries
}

func NewTaxRuleHandler(q *sqlc.Queries) *TaxRuleHandler {
	return &TaxRuleHandler{queries: q}
}

func (h *TaxRuleHandler) verifyPropertyOwnership(w http.ResponseWriter, r *http.Request, propertyID pgtype.UUID) bool {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return false
	}
	owned, err := h.queries.VerifyPropertyOwnership(r.Context(), sqlc.VerifyPropertyOwnershipParams{
		ID:     propertyID,
		UserID: userID,
	})
	if err != nil || !owned {
		respondError(w, http.StatusNotFound, "property not found")
		return false
	}
	return true
}

func (h *TaxRuleHandler) ListByProperty(w http.ResponseWriter, r *http.Request) {
	propertyID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid property id")
		return
	}

	if !h.verifyPropertyOwnership(w, r, propertyID) {
		return
	}

	rules, err := h.queries.ListTaxRulesByProperty(r.Context(), propertyID)
	if err != nil {
		respondInternalError(w, "failed to list tax rules", err)
		return
	}
	respondJSON(w, http.StatusOK, rules)
}

type createTaxRuleRequest struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Percentage  float64 `json:"percentage"`
	FixedAmount float64 `json:"fixed_amount"`
	Currency    string  `json:"currency"`
}

func (h *TaxRuleHandler) Create(w http.ResponseWriter, r *http.Request) {
	propertyID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid property id")
		return
	}

	if !h.verifyPropertyOwnership(w, r, propertyID) {
		return
	}

	var req createTaxRuleRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Type != "percentage" && req.Type != "fixed" {
		respondError(w, http.StatusBadRequest, "type must be 'percentage' or 'fixed'")
		return
	}
	if req.Currency == "" {
		req.Currency = "EUR"
	}

	rule, err := h.queries.CreateTaxRule(r.Context(), sqlc.CreateTaxRuleParams{
		PropertyID:  propertyID,
		Name:        req.Name,
		Type:        req.Type,
		Percentage:  decimalFromFloat(req.Percentage),
		FixedAmount: decimalFromFloat(req.FixedAmount),
		Currency:    req.Currency,
	})
	if err != nil {
		respondInternalError(w, "failed to create tax rule", err)
		return
	}
	respondJSON(w, http.StatusCreated, rule)
}

type updateTaxRuleRequest struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Percentage  float64 `json:"percentage"`
	FixedAmount float64 `json:"fixed_amount"`
	Currency    string  `json:"currency"`
}

func (h *TaxRuleHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid tax rule id")
		return
	}

	rule, err := h.queries.GetTaxRule(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "tax rule not found")
		return
	}

	owned, err := h.queries.VerifyPropertyOwnership(r.Context(), sqlc.VerifyPropertyOwnershipParams{
		ID:     rule.PropertyID,
		UserID: userID,
	})
	if err != nil || !owned {
		respondError(w, http.StatusNotFound, "tax rule not found")
		return
	}

	var req updateTaxRuleRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Type != "percentage" && req.Type != "fixed" {
		respondError(w, http.StatusBadRequest, "type must be 'percentage' or 'fixed'")
		return
	}
	if req.Currency == "" {
		req.Currency = "EUR"
	}

	updated, err := h.queries.UpdateTaxRule(r.Context(), sqlc.UpdateTaxRuleParams{
		ID:          id,
		Name:        req.Name,
		Type:        req.Type,
		Percentage:  decimalFromFloat(req.Percentage),
		FixedAmount: decimalFromFloat(req.FixedAmount),
		Currency:    req.Currency,
	})
	if err != nil {
		respondError(w, http.StatusNotFound, "tax rule not found")
		return
	}
	respondJSON(w, http.StatusOK, updated)
}

func (h *TaxRuleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid tax rule id")
		return
	}

	rule, err := h.queries.GetTaxRule(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "tax rule not found")
		return
	}

	owned, err := h.queries.VerifyPropertyOwnership(r.Context(), sqlc.VerifyPropertyOwnershipParams{
		ID:     rule.PropertyID,
		UserID: userID,
	})
	if err != nil || !owned {
		respondError(w, http.StatusNotFound, "tax rule not found")
		return
	}

	if err := h.queries.DeleteTaxRule(r.Context(), id); err != nil {
		respondInternalError(w, "failed to delete tax rule", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
