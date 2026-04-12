package handlers

import (
	"net/http"

	"github.com/cristianbutiri/property-manager-v2/internal/database/sqlc"
)

type PropertyMappingHandler struct {
	queries *sqlc.Queries
}

func NewPropertyMappingHandler(q *sqlc.Queries) *PropertyMappingHandler {
	return &PropertyMappingHandler{queries: q}
}

type mappingResponse struct {
	ID            string `json:"id"`
	IntegrationID string `json:"integration_id"`
	PropertyID    string `json:"property_id"`
	ExternalID    string `json:"external_id"`
	Platform      string `json:"platform"`
	CreatedAt     string `json:"created_at"`
}

func (h *PropertyMappingHandler) ListByProperty(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	propertyID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid property ID")
		return
	}

	if !verifyPropertyRead(r.Context(), h.queries, propertyID, userID) {
		respondError(w, http.StatusNotFound, "Property not found")
		return
	}

	mappings, err := h.queries.ListMappingsByProperty(r.Context(), propertyID)
	if err != nil {
		respondInternalError(w, "Failed to list mappings", err)
		return
	}

	resp := make([]mappingResponse, len(mappings))
	for i, m := range mappings {
		resp[i] = mappingResponse{
			ID:            uuidBytesToString(m.ID),
			IntegrationID: uuidBytesToString(m.IntegrationID),
			PropertyID:    uuidBytesToString(m.PropertyID),
			ExternalID:    m.ExternalID,
			Platform:      m.Platform,
			CreatedAt:     m.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
	}
	respondJSON(w, http.StatusOK, resp)
}

type createMappingRequest struct {
	IntegrationID string `json:"integration_id"`
	ExternalID    string `json:"external_id"`
}

func (h *PropertyMappingHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	propertyID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid property ID")
		return
	}

	if !verifyPropertyWrite(r.Context(), h.queries, propertyID, userID) {
		respondError(w, http.StatusNotFound, "Property not found")
		return
	}

	var req createMappingRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.IntegrationID == "" || req.ExternalID == "" {
		respondError(w, http.StatusBadRequest, "integration_id and external_id are required")
		return
	}

	integrationID, err := parseUUID(req.IntegrationID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid integration ID")
		return
	}

	mapping, err := h.queries.CreateMapping(r.Context(), sqlc.CreateMappingParams{
		IntegrationID: integrationID,
		PropertyID:    propertyID,
		ExternalID:    req.ExternalID,
	})
	if err != nil {
		respondError(w, http.StatusConflict, "Mapping already exists for this integration/property or external ID")
		return
	}

	respondJSON(w, http.StatusCreated, mappingResponse{
		ID:            uuidBytesToString(mapping.ID),
		IntegrationID: uuidBytesToString(mapping.IntegrationID),
		PropertyID:    uuidBytesToString(mapping.PropertyID),
		ExternalID:    mapping.ExternalID,
		CreatedAt:     mapping.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *PropertyMappingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	propertyID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid property ID")
		return
	}

	if !verifyPropertyWrite(r.Context(), h.queries, propertyID, userID) {
		respondError(w, http.StatusNotFound, "Property not found")
		return
	}

	mappingID, err := getURLParamUUID(r, "mappingId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid mapping ID")
		return
	}

	if err := h.queries.DeleteMapping(r.Context(), mappingID); err != nil {
		respondInternalError(w, "Failed to delete mapping", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
