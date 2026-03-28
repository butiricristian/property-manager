package handlers

import (
	"net/http"

	"github.com/cristianbutiri/property-manager-v2/internal/services"
)

type DashboardHandler struct {
	svc *services.DashboardService
}

func NewDashboardHandler(svc *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) PropertyDashboard(w http.ResponseWriter, r *http.Request) {
	propertyID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid property id")
		return
	}

	month := getMonthParam(r)
	summary, err := h.svc.GetPropertyDashboard(r.Context(), propertyID, month)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get dashboard data")
		return
	}
	respondJSON(w, http.StatusOK, summary)
}

func (h *DashboardHandler) OverallDashboard(w http.ResponseWriter, r *http.Request) {
	month := getMonthParam(r)
	summary, err := h.svc.GetOverallDashboard(r.Context(), month)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get dashboard data")
		return
	}
	respondJSON(w, http.StatusOK, summary)
}
