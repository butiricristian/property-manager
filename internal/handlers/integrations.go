package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/cristianbutiri/property-manager-v2/internal/database/sqlc"
	"github.com/cristianbutiri/property-manager-v2/internal/services"
)

type IntegrationHandler struct {
	queries *sqlc.Queries
	syncSvc *services.SyncService
}

func NewIntegrationHandler(q *sqlc.Queries, syncSvc *services.SyncService) *IntegrationHandler {
	return &IntegrationHandler{queries: q, syncSvc: syncSvc}
}

type integrationResponse struct {
	ID            string            `json:"id"`
	UserID        string            `json:"user_id"`
	Platform      string            `json:"platform"`
	Credentials   map[string]string `json:"credentials"`
	Enabled       bool              `json:"enabled"`
	LastSyncedAt  *time.Time        `json:"last_synced_at"`
	LastSyncError *string           `json:"last_sync_error"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

func toIntegrationResponse(i sqlc.Integration, mask bool) integrationResponse {
	creds := parseCredentials(i.Credentials)
	if mask {
		creds = maskCredentials(creds)
	}

	resp := integrationResponse{
		ID:          uuidBytesToString(i.ID),
		UserID:      uuidBytesToString(i.UserID),
		Platform:    i.Platform,
		Credentials: creds,
		Enabled:     i.Enabled,
		CreatedAt:   i.CreatedAt.Time,
		UpdatedAt:   i.UpdatedAt.Time,
	}
	if i.LastSyncedAt.Valid {
		t := i.LastSyncedAt.Time
		resp.LastSyncedAt = &t
	}
	if i.LastSyncError.Valid {
		s := i.LastSyncError.String
		resp.LastSyncError = &s
	}
	return resp
}

func parseCredentials(data []byte) map[string]string {
	var creds map[string]string
	if err := json.Unmarshal(data, &creds); err != nil {
		return map[string]string{}
	}
	return creds
}

func maskCredentials(creds map[string]string) map[string]string {
	masked := make(map[string]string, len(creds))
	for k, v := range creds {
		if len(v) <= 4 {
			masked[k] = "****"
		} else {
			masked[k] = "****" + v[len(v)-4:]
		}
	}
	return masked
}

func (h *IntegrationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	integrations, err := h.queries.ListIntegrationsByUser(r.Context(), userID)
	if err != nil {
		respondInternalError(w, "failed to list integrations", err)
		return
	}

	result := make([]integrationResponse, len(integrations))
	for i, integ := range integrations {
		result[i] = toIntegrationResponse(integ, true)
	}
	respondJSON(w, http.StatusOK, result)
}

type createIntegrationRequest struct {
	Platform    string            `json:"platform"`
	Credentials map[string]string `json:"credentials"`
}

func (h *IntegrationHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createIntegrationRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Platform != "booking" && req.Platform != "airbnb" {
		respondError(w, http.StatusBadRequest, "platform must be 'booking' or 'airbnb'")
		return
	}

	credsJSON, err := json.Marshal(req.Credentials)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid credentials")
		return
	}

	integration, err := h.queries.CreateIntegration(r.Context(), sqlc.CreateIntegrationParams{
		UserID:      userID,
		Platform:    req.Platform,
		Credentials: credsJSON,
		Enabled:     true,
	})
	if err != nil {
		respondInternalError(w, "failed to create integration", err)
		return
	}

	respondJSON(w, http.StatusCreated, toIntegrationResponse(integration, true))
}

type updateIntegrationRequest struct {
	Credentials map[string]string `json:"credentials,omitempty"`
	Enabled     *bool             `json:"enabled,omitempty"`
}

func (h *IntegrationHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid integration id")
		return
	}

	existing, err := h.queries.GetIntegration(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "integration not found")
		return
	}
	if existing.UserID != userID {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req updateIntegrationRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	creds := parseCredentials(existing.Credentials)
	if req.Credentials != nil {
		creds = req.Credentials
	}
	credsJSON, _ := json.Marshal(creds)

	enabled := existing.Enabled
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	integration, err := h.queries.UpdateIntegration(r.Context(), sqlc.UpdateIntegrationParams{
		ID:          id,
		Credentials: credsJSON,
		Enabled:     enabled,
	})
	if err != nil {
		respondInternalError(w, "failed to update integration", err)
		return
	}

	respondJSON(w, http.StatusOK, toIntegrationResponse(integration, true))
}

func (h *IntegrationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid integration id")
		return
	}

	existing, err := h.queries.GetIntegration(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "integration not found")
		return
	}
	if existing.UserID != userID {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := h.queries.DeleteIntegration(r.Context(), id); err != nil {
		respondInternalError(w, "failed to delete integration", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *IntegrationHandler) Sync(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid integration id")
		return
	}

	integration, err := h.queries.GetIntegration(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "integration not found")
		return
	}
	if integration.UserID != userID {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	result := h.syncSvc.SyncIntegration(r.Context(), integration)

	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	var syncError pgtype.Text
	if result.Errors > 0 {
		syncError = pgtype.Text{String: result.Message, Valid: true}
	}
	_ = h.queries.UpdateIntegrationSyncStatus(r.Context(), sqlc.UpdateIntegrationSyncStatusParams{
		ID:            id,
		LastSyncedAt:  now,
		LastSyncError: syncError,
	})

	respondJSON(w, http.StatusOK, result)
}
