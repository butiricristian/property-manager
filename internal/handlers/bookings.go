package handlers

import (
	"math"
	"math/big"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/cristianbutiri/property-manager-v2/internal/database/sqlc"
	"github.com/cristianbutiri/property-manager-v2/internal/services"
)

type BookingHandler struct {
	queries     *sqlc.Queries
	exchangeSvc *services.ExchangeService
}

func NewBookingHandler(q *sqlc.Queries, exchangeSvc *services.ExchangeService) *BookingHandler {
	return &BookingHandler{queries: q, exchangeSvc: exchangeSvc}
}

func (h *BookingHandler) verifyPropertyRead(w http.ResponseWriter, r *http.Request, propertyID pgtype.UUID) bool {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return false
	}
	if !verifyPropertyRead(r.Context(), h.queries, propertyID, userID) {
		respondError(w, http.StatusNotFound, "property not found")
		return false
	}
	return true
}

func (h *BookingHandler) verifyPropertyWrite(w http.ResponseWriter, r *http.Request, propertyID pgtype.UUID) bool {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return false
	}
	if !verifyPropertyWrite(r.Context(), h.queries, propertyID, userID) {
		respondError(w, http.StatusNotFound, "property not found")
		return false
	}
	return true
}

func (h *BookingHandler) ListByProperty(w http.ResponseWriter, r *http.Request) {
	propertyID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid property id")
		return
	}

	if !h.verifyPropertyRead(w, r, propertyID) {
		return
	}

	startDate, endDate, err := getDateRangeParams(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid date range")
		return
	}
	month := startDate.Time.Format("2006-01")

	bookings, err := h.queries.ListBookingsByProperty(r.Context(), sqlc.ListBookingsByPropertyParams{
		PropertyID: propertyID,
		StartDate:  startDate,
		EndDate:    endDate,
	})
	if err != nil {
		respondInternalError(w, "failed to list bookings", err)
		return
	}

	targetCurrency := r.URL.Query().Get("currency")
	if targetCurrency != "" {
		for i := range bookings {
			b := &bookings[i]
			if b.Currency == targetCurrency {
				continue
			}
			gross, err := h.exchangeSvc.Convert(r.Context(), numericToDecimal(b.GrossAmount), b.Currency, targetCurrency, month)
			if err != nil {
				respondInternalError(w, "failed to convert currency", err)
				return
			}
			price, err := h.exchangeSvc.Convert(r.Context(), numericToDecimal(b.Price), b.Currency, targetCurrency, month)
			if err != nil {
				respondInternalError(w, "failed to convert currency", err)
				return
			}
			fee, err := h.exchangeSvc.Convert(r.Context(), numericToDecimal(b.PlatformFee), b.Currency, targetCurrency, month)
			if err != nil {
				respondInternalError(w, "failed to convert currency", err)
				return
			}
			net, err := h.exchangeSvc.Convert(r.Context(), numericToDecimal(b.NetAmount), b.Currency, targetCurrency, month)
			if err != nil {
				respondInternalError(w, "failed to convert currency", err)
				return
			}
			b.GrossAmount = decimalToNumeric(gross)
			b.Price = decimalToNumeric(price)
			b.PlatformFee = decimalToNumeric(fee)
			b.NetAmount = decimalToNumeric(net)
			b.Currency = targetCurrency
		}
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
	Price             float64 `json:"price"`
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

	if !h.verifyPropertyWrite(w, r, propertyID) {
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
	if req.Price == 0 {
		req.Price = req.GrossAmount
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
		Price:             decimalFromFloat(req.Price),
	})
	if err != nil {
		respondInternalError(w, "failed to create booking", err)
		return
	}
	respondJSON(w, http.StatusCreated, booking)
}

func (h *BookingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid booking id")
		return
	}

	booking, err := h.queries.GetBooking(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "booking not found")
		return
	}

	if !verifyPropertyWrite(r.Context(), h.queries, booking.PropertyID, userID) {
		respondError(w, http.StatusNotFound, "booking not found")
		return
	}

	if err := h.queries.DeleteBooking(r.Context(), id); err != nil {
		respondInternalError(w, "failed to delete booking", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
