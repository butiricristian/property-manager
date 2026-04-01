package handlers

import (
	"net/http"

	"github.com/cristianbutiri/property-manager-v2/internal/database/sqlc"
	"github.com/cristianbutiri/property-manager-v2/internal/services"
)

type ImportHandler struct {
	svc     *services.ImportService
	queries *sqlc.Queries
}

func NewImportHandler(svc *services.ImportService, q *sqlc.Queries) *ImportHandler {
	return &ImportHandler{svc: svc, queries: q}
}

func (h *ImportHandler) ImportFile(w http.ResponseWriter, r *http.Request) {
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

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	platform := r.FormValue("platform")
	if platform != "booking" && platform != "airbnb" {
		respondError(w, http.StatusBadRequest, "platform must be 'booking' or 'airbnb'")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	var result *services.ImportResult

	payoutsFile, payoutsHeader, payoutsErr := r.FormFile("payouts_file")
	if payoutsErr == nil {
		defer payoutsFile.Close()
		result, err = h.svc.ImportFileWithPayouts(r.Context(), propertyID, platform, header.Filename, file, payoutsHeader.Filename, payoutsFile)
	} else {
		result, err = h.svc.ImportFile(r.Context(), propertyID, platform, header.Filename, file)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}
