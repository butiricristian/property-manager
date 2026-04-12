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

	if !verifyPropertyRead(r.Context(), h.queries, propertyID, userID) {
		respondError(w, http.StatusNotFound, "property not found")
		return
	}

	startDate, endDate, err := getDateRangeParams(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid date range")
		return
	}
	currency := r.URL.Query().Get("currency")
	summary, err := h.svc.GetPropertyDashboard(r.Context(), propertyID, startDate, endDate, currency)
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

	startDate, endDate, err := getDateRangeParams(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid date range")
		return
	}
	currency := r.URL.Query().Get("currency")
	summary, err := h.svc.GetOverallDashboard(r.Context(), userID, startDate, endDate, currency)
	if err != nil {
		respondInternalError(w, "failed to get dashboard data", err)
		return
	}
	respondJSON(w, http.StatusOK, summary)
}
