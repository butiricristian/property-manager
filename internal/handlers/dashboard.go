package handlers

import (
	"net/http"

	"github.com/cristianbutiri/property-manager-v2/internal/database/sqlc"
	"github.com/cristianbutiri/property-manager-v2/internal/services"
)

type DashboardHandler struct {
	svc     *services.DashboardService
	queries *sqlc.Queries
}

func NewDashboardHandler(svc *services.DashboardService, q *sqlc.Queries) *DashboardHandler {
	return &DashboardHandler{svc: svc, queries: q}
}

func (h *DashboardHandler) PropertyDashboard(w http.ResponseWriter, r *http.Request) {
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

	owned, err := h.queries.VerifyPropertyOwnership(r.Context(), sqlc.VerifyPropertyOwnershipParams{
		ID:     propertyID,
		UserID: userID,
	})
	if err != nil || !owned {
		respondError(w, http.StatusNotFound, "property not found")
		return
	}

	month := getMonthParam(r)
	currency := r.URL.Query().Get("currency")
	summary, err := h.svc.GetPropertyDashboard(r.Context(), propertyID, month, currency)
	if err != nil {
		respondInternalError(w, "failed to get dashboard data", err)
		return
	}
	respondJSON(w, http.StatusOK, summary)
}

func (h *DashboardHandler) OverallDashboard(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	month := getMonthParam(r)
	currency := r.URL.Query().Get("currency")
	summary, err := h.svc.GetOverallDashboard(r.Context(), userID, month, currency)
	if err != nil {
		respondInternalError(w, "failed to get dashboard data", err)
		return
	}
	respondJSON(w, http.StatusOK, summary)
}
