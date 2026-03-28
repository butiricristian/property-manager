package handlers

import (
	"net/http"

	"github.com/cristianbutiri/property-manager-v2/internal/services"
)

type ImportHandler struct {
	svc *services.ImportService
}

func NewImportHandler(svc *services.ImportService) *ImportHandler {
	return &ImportHandler{svc: svc}
}

func (h *ImportHandler) ImportFile(w http.ResponseWriter, r *http.Request) {
	propertyID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid property id")
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

	result, err := h.svc.ImportFile(r.Context(), propertyID, platform, header.Filename, file)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}
