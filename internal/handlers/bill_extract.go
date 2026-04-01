package handlers

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/cristianbutiri/property-manager-v2/internal/database/sqlc"
	"github.com/cristianbutiri/property-manager-v2/internal/services"
)

type BillExtractHandler struct {
	svc     *services.BillExtractService
	queries *sqlc.Queries
}

func NewBillExtractHandler(svc *services.BillExtractService, q *sqlc.Queries) *BillExtractHandler {
	return &BillExtractHandler{svc: svc, queries: q}
}

func (h *BillExtractHandler) Extract(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		respondError(w, http.StatusServiceUnavailable, "bill extraction is not configured (missing ANTHROPIC_API_KEY)")
		return
	}

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

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".pdf" {
		respondError(w, http.StatusBadRequest, "only PDF files are supported")
		return
	}

	pdfBytes, err := io.ReadAll(file)
	if err != nil {
		respondInternalError(w, "failed to read file", err)
		return
	}

	result, err := h.svc.ExtractFromPDF(r.Context(), pdfBytes)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}
