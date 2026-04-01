package handlers

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/cristianbutiri/property-manager-v2/internal/database/sqlc"
)

type SellerProfileHandler struct {
	queries *sqlc.Queries
}

func NewSellerProfileHandler(q *sqlc.Queries) *SellerProfileHandler {
	return &SellerProfileHandler{queries: q}
}

func (h *SellerProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	profile, err := h.queries.GetSellerProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondJSON(w, http.StatusOK, nil)
			return
		}
		respondInternalError(w, "failed to get seller profile", err)
		return
	}

	respondJSON(w, http.StatusOK, profile)
}

type upsertSellerProfileRequest struct {
	LegalName       string `json:"legal_name"`
	Street          string `json:"street"`
	City            string `json:"city"`
	County          string `json:"county"`
	PostalCode      string `json:"postal_code"`
	CountryCode     string `json:"country_code"`
	CUI             string `json:"cui"`
	CNP             string `json:"cnp"`
	TradeRegisterNr string `json:"trade_register_nr"`
	VATRegistered   bool   `json:"vat_registered"`
}

func (h *SellerProfileHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req upsertSellerProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.LegalName == "" {
		respondError(w, http.StatusBadRequest, "legal_name is required")
		return
	}
	if req.CountryCode == "" {
		req.CountryCode = "RO"
	}

	profile, err := h.queries.UpsertSellerProfile(r.Context(), sqlc.UpsertSellerProfileParams{
		UserID:          userID,
		LegalName:       req.LegalName,
		Street:          req.Street,
		City:            req.City,
		County:          req.County,
		PostalCode:      req.PostalCode,
		CountryCode:     req.CountryCode,
		Cui:             req.CUI,
		Cnp:             req.CNP,
		TradeRegisterNr: req.TradeRegisterNr,
		VatRegistered:   req.VATRegistered,
	})
	if err != nil {
		respondInternalError(w, "failed to upsert seller profile", err)
		return
	}

	respondJSON(w, http.StatusOK, profile)
}
