package handlers

import (
	"net/http"

	"github.com/cristianbutiri/property-manager-v2/internal/database/sqlc"
)

type PropertyHandler struct {
	queries *sqlc.Queries
}

func NewPropertyHandler(q *sqlc.Queries) *PropertyHandler {
	return &PropertyHandler{queries: q}
}

func (h *PropertyHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	properties, err := h.queries.ListProperties(r.Context(), userID)
	if err != nil {
		respondInternalError(w, "failed to list properties", err)
		return
	}
	respondJSON(w, http.StatusOK, properties)
}

func (h *PropertyHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid property id")
		return
	}

	property, err := h.queries.GetProperty(r.Context(), sqlc.GetPropertyParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		respondError(w, http.StatusNotFound, "property not found")
		return
	}
	respondJSON(w, http.StatusOK, property)
}

type createPropertyRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

func (h *PropertyHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createPropertyRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}

	property, err := h.queries.CreateProperty(r.Context(), sqlc.CreatePropertyParams{
		Name:    req.Name,
		Address: req.Address,
		UserID:  userID,
	})
	if err != nil {
		respondInternalError(w, "failed to create property", err)
		return
	}
	respondJSON(w, http.StatusCreated, property)
}

type updatePropertyRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

func (h *PropertyHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid property id")
		return
	}

	var req updatePropertyRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}

	property, err := h.queries.UpdateProperty(r.Context(), sqlc.UpdatePropertyParams{
		ID:      id,
		Name:    req.Name,
		Address: req.Address,
		UserID:  userID,
	})
	if err != nil {
		respondError(w, http.StatusNotFound, "property not found")
		return
	}
	respondJSON(w, http.StatusOK, property)
}

func (h *PropertyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid property id")
		return
	}

	if err := h.queries.DeleteProperty(r.Context(), sqlc.DeletePropertyParams{
		ID:     id,
		UserID: userID,
	}); err != nil {
		respondInternalError(w, "failed to delete property", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
