package handlers

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/cristianbutiri/property-manager-v2/internal/database/sqlc"
)

type InventoryHandler struct {
	queries *sqlc.Queries
}

func NewInventoryHandler(q *sqlc.Queries) *InventoryHandler {
	return &InventoryHandler{queries: q}
}

func (h *InventoryHandler) verifyPropertyRead(w http.ResponseWriter, r *http.Request, propertyID pgtype.UUID) bool {
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

func (h *InventoryHandler) verifyPropertyWrite(w http.ResponseWriter, r *http.Request, propertyID pgtype.UUID) bool {
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

func (h *InventoryHandler) ListByProperty(w http.ResponseWriter, r *http.Request) {
	propertyID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid property id")
		return
	}

	if !h.verifyPropertyRead(w, r, propertyID) {
		return
	}

	items, err := h.queries.ListInventoryByProperty(r.Context(), propertyID)
	if err != nil {
		respondInternalError(w, "failed to list inventory", err)
		return
	}
	respondJSON(w, http.StatusOK, items)
}

type createInventoryRequest struct {
	Name           string `json:"name"`
	Category       string `json:"category"`
	Quantity       int32  `json:"quantity"`
	AcquiredDate   string `json:"acquired_date"`
	LifetimeMonths int32  `json:"lifetime_months"`
	Notes          string `json:"notes"`
}

func (h *InventoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	propertyID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid property id")
		return
	}

	if !h.verifyPropertyWrite(w, r, propertyID) {
		return
	}

	var req createInventoryRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	date, err := time.Parse("2006-01-02", req.AcquiredDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid date format")
		return
	}

	item, err := h.queries.CreateInventoryItem(r.Context(), sqlc.CreateInventoryItemParams{
		PropertyID:     propertyID,
		Name:           req.Name,
		Category:       req.Category,
		Quantity:       req.Quantity,
		AcquiredDate:   pgtype.Date{Time: date, Valid: true},
		LifetimeMonths: req.LifetimeMonths,
		Notes:          req.Notes,
	})
	if err != nil {
		respondInternalError(w, "failed to create inventory item", err)
		return
	}
	respondJSON(w, http.StatusCreated, item)
}

type updateInventoryRequest struct {
	Name           string `json:"name"`
	Category       string `json:"category"`
	Quantity       int32  `json:"quantity"`
	AcquiredDate   string `json:"acquired_date"`
	LifetimeMonths int32  `json:"lifetime_months"`
	Notes          string `json:"notes"`
}

func (h *InventoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid inventory item id")
		return
	}

	item, err := h.queries.GetInventoryItem(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "inventory item not found")
		return
	}

	if !verifyPropertyWrite(r.Context(), h.queries, item.PropertyID, userID) {
		respondError(w, http.StatusNotFound, "inventory item not found")
		return
	}

	var req updateInventoryRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	date, err := time.Parse("2006-01-02", req.AcquiredDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid date format")
		return
	}

	updated, err := h.queries.UpdateInventoryItem(r.Context(), sqlc.UpdateInventoryItemParams{
		ID:             id,
		Name:           req.Name,
		Category:       req.Category,
		Quantity:       req.Quantity,
		AcquiredDate:   pgtype.Date{Time: date, Valid: true},
		LifetimeMonths: req.LifetimeMonths,
		Notes:          req.Notes,
	})
	if err != nil {
		respondError(w, http.StatusNotFound, "inventory item not found")
		return
	}
	respondJSON(w, http.StatusOK, updated)
}

func (h *InventoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid inventory item id")
		return
	}

	item, err := h.queries.GetInventoryItem(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "inventory item not found")
		return
	}

	if !verifyPropertyWrite(r.Context(), h.queries, item.PropertyID, userID) {
		respondError(w, http.StatusNotFound, "inventory item not found")
		return
	}

	if err := h.queries.DeleteInventoryItem(r.Context(), id); err != nil {
		respondInternalError(w, "failed to delete inventory item", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
