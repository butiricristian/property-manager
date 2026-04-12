package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/cristianbutiri/property-manager-v2/internal/database/sqlc"
	"github.com/cristianbutiri/property-manager-v2/internal/services"
)

type InvoiceHandler struct {
	queries  *sqlc.Queries
	anafSvc  *services.ANAFService
}

func NewInvoiceHandler(q *sqlc.Queries, anafSvc *services.ANAFService) *InvoiceHandler {
	return &InvoiceHandler{queries: q, anafSvc: anafSvc}
}

// ANAFLookup proxies a CUI lookup to ANAF
func (h *InvoiceHandler) ANAFLookup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CUI string `json:"cui"`
	}
	if err := decodeJSON(r, &req); err != nil || req.CUI == "" {
		respondError(w, http.StatusBadRequest, "cui is required")
		return
	}

	company, err := h.anafSvc.LookupCompany(r.Context(), req.CUI)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, company)
}

type generateInvoiceRequest struct {
	BuyerLegalName       string  `json:"buyer_legal_name"`
	BuyerCUI             string  `json:"buyer_cui"`
	BuyerTradeRegisterNr string  `json:"buyer_trade_register_nr"`
	BuyerStreet          string  `json:"buyer_street"`
	BuyerCity            string  `json:"buyer_city"`
	BuyerCounty          string  `json:"buyer_county"`
	BuyerPostalCode      string  `json:"buyer_postal_code"`
	BuyerCountryCode     string  `json:"buyer_country_code"`
	BuyerVATRegistered   bool    `json:"buyer_vat_registered"`
	NetAmount            float64 `json:"net_amount"`
	Currency             string  `json:"currency"`
}

// GenerateInvoice creates an invoice for a booking
func (h *InvoiceHandler) GenerateInvoice(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	bookingID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid booking id")
		return
	}

	// Check if invoice already exists — if so, delete it and reuse its number
	var reuseInvoiceNumber int32
	existingInvoice, err := h.queries.GetInvoiceByBooking(r.Context(), bookingID)
	if err == nil {
		reuseInvoiceNumber = existingInvoice.InvoiceNumber
		if err := h.queries.DeleteInvoice(r.Context(), existingInvoice.ID); err != nil {
			respondInternalError(w, "failed to delete existing invoice", err)
			return
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		respondInternalError(w, "failed to check existing invoice", err)
		return
	}

	// Verify booking ownership via property
	booking, err := h.queries.GetBooking(r.Context(), bookingID)
	if err != nil {
		respondError(w, http.StatusNotFound, "booking not found")
		return
	}

	if !verifyPropertyWrite(r.Context(), h.queries, booking.PropertyID, userID) {
		respondError(w, http.StatusNotFound, "booking not found")
		return
	}

	// Get seller profile
	seller, err := h.queries.GetSellerProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusBadRequest, "please set up your billing profile first")
			return
		}
		respondInternalError(w, "failed to get seller profile", err)
		return
	}

	var req generateInvoiceRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.BuyerLegalName == "" {
		respondError(w, http.StatusBadRequest, "buyer_legal_name is required")
		return
	}
	if req.BuyerCountryCode == "" {
		req.BuyerCountryCode = "RO"
	}
	if req.Currency == "" {
		req.Currency = "RON"
	}

	// Calculate tax
	var taxPercent float64
	if seller.VatRegistered {
		taxPercent = 19.0
	}
	netAmount := req.NetAmount
	taxAmount := netAmount * taxPercent / 100.0
	grossAmount := netAmount + taxAmount

	// Reuse existing invoice number or get next one
	nextNum := reuseInvoiceNumber
	if nextNum == 0 {
		nextNum, err = h.queries.GetNextInvoiceNumber(r.Context(), userID)
		if err != nil {
			respondInternalError(w, "failed to get next invoice number", err)
			return
		}
	}

	now := time.Now()
	issueDate := pgtype.Date{Time: now, Valid: true}
	dueDate := pgtype.Date{Time: now.AddDate(0, 0, 15), Valid: true}

	description := fmt.Sprintf("Servicii cazare - %s, %d nopti", booking.GuestName, booking.Nights)

	// Generate XML
	xmlData, err := services.GenerateEFacturaXML(&services.InvoiceData{
		InvoiceNumber: nextNum,
		InvoiceSeries: "PM",
		IssueDate:     now,
		DueDate:       now.AddDate(0, 0, 15),

		SellerName:            seller.LegalName,
		SellerCUI:             seller.Cui,
		SellerCNP:             seller.Cnp,
		SellerTradeRegisterNr: seller.TradeRegisterNr,
		SellerStreet:          seller.Street,
		SellerCity:            seller.City,
		SellerCounty:          seller.County,
		SellerPostalCode:      seller.PostalCode,
		SellerCountryCode:     seller.CountryCode,
		SellerVATRegistered:   seller.VatRegistered,

		BuyerName:            req.BuyerLegalName,
		BuyerCUI:             req.BuyerCUI,
		BuyerTradeRegisterNr: req.BuyerTradeRegisterNr,
		BuyerStreet:          req.BuyerStreet,
		BuyerCity:            req.BuyerCity,
		BuyerCounty:          req.BuyerCounty,
		BuyerPostalCode:      req.BuyerPostalCode,
		BuyerCountryCode:     req.BuyerCountryCode,
		BuyerVATRegistered:   req.BuyerVATRegistered,

		Currency:    req.Currency,
		NetAmount:   netAmount,
		TaxAmount:   taxAmount,
		GrossAmount: grossAmount,
		TaxPercent:  taxPercent,

		Description: description,
		Nights:      booking.Nights,
	})
	if err != nil {
		respondInternalError(w, "failed to generate XML", err)
		return
	}

	invoice, err := h.queries.CreateInvoice(r.Context(), sqlc.CreateInvoiceParams{
		UserID:    userID,
		BookingID: bookingID,
		InvoiceNumber: nextNum,
		InvoiceSeries: "PM",
		IssueDate: issueDate,
		DueDate:   dueDate,

		SellerLegalName:       seller.LegalName,
		SellerCui:             seller.Cui,
		SellerCnp:             seller.Cnp,
		SellerTradeRegisterNr: seller.TradeRegisterNr,
		SellerStreet:          seller.Street,
		SellerCity:            seller.City,
		SellerCounty:          seller.County,
		SellerPostalCode:      seller.PostalCode,
		SellerCountryCode:     seller.CountryCode,
		SellerVatRegistered:   seller.VatRegistered,

		BuyerLegalName:       req.BuyerLegalName,
		BuyerCui:             req.BuyerCUI,
		BuyerTradeRegisterNr: req.BuyerTradeRegisterNr,
		BuyerStreet:          req.BuyerStreet,
		BuyerCity:            req.BuyerCity,
		BuyerCounty:          req.BuyerCounty,
		BuyerPostalCode:      req.BuyerPostalCode,
		BuyerCountryCode:     req.BuyerCountryCode,
		BuyerVatRegistered:   req.BuyerVATRegistered,

		Currency:    req.Currency,
		NetAmount:   decimalFromFloat(netAmount),
		TaxAmount:   decimalFromFloat(taxAmount),
		GrossAmount: decimalFromFloat(grossAmount),
		TaxPercent:  decimalFromFloat(taxPercent),

		Description: description,
		Nights:      booking.Nights,
		XmlContent:  string(xmlData),
	})
	if err != nil {
		respondInternalError(w, "failed to create invoice", err)
		return
	}

	respondJSON(w, http.StatusCreated, invoice)
}

// GetBookingInvoice returns the invoice for a specific booking
func (h *InvoiceHandler) GetBookingInvoice(w http.ResponseWriter, r *http.Request) {
	bookingID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid booking id")
		return
	}

	invoice, err := h.queries.GetInvoiceByBooking(r.Context(), bookingID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound, "no invoice for this booking")
			return
		}
		respondInternalError(w, "failed to get invoice", err)
		return
	}

	respondJSON(w, http.StatusOK, invoice)
}

// DownloadXML serves the XML content of an invoice
func (h *InvoiceHandler) DownloadXML(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	invoiceID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	invoice, err := h.queries.GetInvoice(r.Context(), invoiceID)
	if err != nil {
		respondError(w, http.StatusNotFound, "invoice not found")
		return
	}

	if uuidBytesToString(invoice.UserID) != uuidBytesToString(userID) {
		respondError(w, http.StatusNotFound, "invoice not found")
		return
	}

	filename := fmt.Sprintf("%s%d.xml", invoice.InvoiceSeries, invoice.InvoiceNumber)
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(invoice.XmlContent))
}

// List returns all invoices for the current user
func (h *InvoiceHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	invoices, err := h.queries.ListInvoicesByUser(r.Context(), userID)
	if err != nil {
		respondInternalError(w, "failed to list invoices", err)
		return
	}

	respondJSON(w, http.StatusOK, invoices)
}

// Delete removes an invoice
func (h *InvoiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	invoiceID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	invoice, err := h.queries.GetInvoice(r.Context(), invoiceID)
	if err != nil {
		respondError(w, http.StatusNotFound, "invoice not found")
		return
	}

	if uuidBytesToString(invoice.UserID) != uuidBytesToString(userID) {
		respondError(w, http.StatusNotFound, "invoice not found")
		return
	}

	if err := h.queries.DeleteInvoice(r.Context(), invoiceID); err != nil {
		respondInternalError(w, "failed to delete invoice", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
