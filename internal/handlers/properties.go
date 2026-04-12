package handlers

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

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

	properties, err := h.queries.ListAccessibleProperties(r.Context(), userID)
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

	// Enforce guest property limit
	userType, err := h.queries.GetUserType(r.Context(), userID)
	if err != nil {
		respondInternalError(w, "failed to check user type", err)
		return
	}
	if userType == "guest" {
		count, err := h.queries.CountProperties(r.Context(), userID)
		if err != nil {
			respondInternalError(w, "failed to count properties", err)
			return
		}
		if count >= 1 {
			respondError(w, http.StatusForbidden, "property_limit")
			return
		}
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

	if !verifyPropertyWrite(r.Context(), h.queries, id, userID) {
		respondError(w, http.StatusNotFound, "property not found")
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

// --- Share management ---

type propertyShareResponse struct {
	UserID     string `json:"user_id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Permission string `json:"permission"`
}

func (h *PropertyHandler) ListShares(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	propertyID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid property id")
		return
	}

	if !verifyPropertyWrite(r.Context(), h.queries, propertyID, userID) {
		respondError(w, http.StatusNotFound, "property not found")
		return
	}

	shares, err := h.queries.ListPropertySharesWithUsers(r.Context(), sqlc.ListPropertySharesWithUsersParams{
		PropertyID: propertyID,
		UserID:     userID,
	})
	if err != nil {
		respondInternalError(w, "failed to list shares", err)
		return
	}

	resp := make([]propertyShareResponse, len(shares))
	for i, s := range shares {
		resp[i] = propertyShareResponse{
			UserID:     uuidBytesToString(s.UserID),
			Email:      s.Email,
			Name:       s.Name,
			Permission: s.Permission,
		}
	}
	respondJSON(w, http.StatusOK, resp)
}

type createShareRequest struct {
	Email string `json:"email"`
}

func (h *PropertyHandler) CreateShare(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Guests cannot share properties
	userType, err := h.queries.GetUserType(r.Context(), userID)
	if err != nil {
		respondInternalError(w, "failed to check user type", err)
		return
	}
	if userType == "guest" {
		respondError(w, http.StatusForbidden, "guests cannot share properties")
		return
	}

	propertyID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid property id")
		return
	}

	if !verifyPropertyWrite(r.Context(), h.queries, propertyID, userID) {
		respondError(w, http.StatusNotFound, "property not found")
		return
	}

	var req createShareRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" {
		respondError(w, http.StatusBadRequest, "email is required")
		return
	}

	invitee, err := h.queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusBadRequest, "no user with that email")
			return
		}
		respondInternalError(w, "failed to look up user", err)
		return
	}

	if uuidBytesToString(invitee.ID) == uuidBytesToString(userID) {
		respondError(w, http.StatusBadRequest, "cannot share with yourself")
		return
	}

	// Check if already shared
	existing, err := h.queries.GetPropertyShare(r.Context(), sqlc.GetPropertyShareParams{
		PropertyID: propertyID,
		UserID:     invitee.ID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		respondInternalError(w, "failed to check existing share", err)
		return
	}
	if err == nil {
		_ = existing
		respondError(w, http.StatusConflict, "already shared with this user")
		return
	}

	_, err = h.queries.CreatePropertyShare(r.Context(), sqlc.CreatePropertyShareParams{
		PropertyID: propertyID,
		UserID:     invitee.ID,
		Permission: "read",
	})
	if err != nil {
		respondInternalError(w, "failed to create share", err)
		return
	}

	respondJSON(w, http.StatusCreated, propertyShareResponse{
		UserID:     uuidBytesToString(invitee.ID),
		Email:      invitee.Email,
		Name:       invitee.Name,
		Permission: "read",
	})
}

func (h *PropertyHandler) DeleteShare(w http.ResponseWriter, r *http.Request) {
	ownerID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	propertyID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid property id")
		return
	}

	shareUserID, err := getURLParamUUID(r, "userId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if !verifyPropertyWrite(r.Context(), h.queries, propertyID, ownerID) {
		respondError(w, http.StatusNotFound, "property not found")
		return
	}

	if err := h.queries.DeletePropertyShare(r.Context(), sqlc.DeletePropertyShareParams{
		PropertyID: propertyID,
		UserID:     shareUserID,
		UserID_2:   ownerID,
	}); err != nil {
		respondInternalError(w, "failed to delete share", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

