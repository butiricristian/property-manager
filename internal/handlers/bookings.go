package handlers

import (
	"math"
	"math/big"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/cristianbutiri/property-manager-v2/internal/database/sqlc"
)

type BookingHandler struct {
	queries *sqlc.Queries
}

func NewBookingHandler(q *sqlc.Queries) *BookingHandler {
	return &BookingHandler{queries: q}
}

func (h *BookingHandler) ListByProperty(w http.ResponseWriter, r *http.Request) {
	propertyID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid property id")
		return
	}

	month := getMonthParam(r)
	startDate, endDate, err := parseMonthDates(month)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid month format, use YYYY-MM")
		return
	}

	bookings, err := h.queries.ListBookingsByProperty(r.Context(), sqlc.ListBookingsByPropertyParams{
		PropertyID: propertyID,
		StartDate:  startDate,
		EndDate:    endDate,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list bookings")
		return
	}
	respondJSON(w, http.StatusOK, bookings)
}

type createBookingRequest struct {
	Platform          string  `json:"platform"`
	PlatformReference string  `json:"platform_reference"`
	GuestName         string  `json:"guest_name"`
	CheckIn           string  `json:"check_in"`
	CheckOut          string  `json:"check_out"`
	Nights            int32   `json:"nights"`
	GrossAmount       float64 `json:"gross_amount"`
	PlatformFee       float64 `json:"platform_fee"`
	NetAmount         float64 `json:"net_amount"`
	Currency          string  `json:"currency"`
	Status            string  `json:"status"`
}

func decimalFromFloat(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	n.Int = big.NewInt(int64(math.Round(f * 100)))
	n.Exp = -2
	n.Valid = true
	return n
}

func (h *BookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	propertyID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid property id")
		return
	}

	var req createBookingRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid check_in date")
		return
	}
	checkOut, err := time.Parse("2006-01-02", req.CheckOut)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid check_out date")
		return
	}

	if req.Currency == "" {
		req.Currency = "EUR"
	}
	if req.Status == "" {
		req.Status = "confirmed"
	}

	booking, err := h.queries.CreateBooking(r.Context(), sqlc.CreateBookingParams{
		PropertyID:        propertyID,
		Platform:          req.Platform,
		PlatformReference: req.PlatformReference,
		GuestName:         req.GuestName,
		CheckIn:           pgtype.Date{Time: checkIn, Valid: true},
		CheckOut:          pgtype.Date{Time: checkOut, Valid: true},
		Nights:            req.Nights,
		GrossAmount:       decimalFromFloat(req.GrossAmount),
		PlatformFee:       decimalFromFloat(req.PlatformFee),
		NetAmount:         decimalFromFloat(req.NetAmount),
		Currency:          req.Currency,
		Status:            req.Status,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create booking")
		return
	}
	respondJSON(w, http.StatusCreated, booking)
}

func (h *BookingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid booking id")
		return
	}

	if err := h.queries.DeleteBooking(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete booking")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
