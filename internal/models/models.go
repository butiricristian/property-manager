package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Property struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Booking struct {
	ID                uuid.UUID       `json:"id"`
	PropertyID        uuid.UUID       `json:"property_id"`
	Platform          string          `json:"platform"`
	PlatformReference string          `json:"platform_reference"`
	GuestName         string          `json:"guest_name"`
	CheckIn           time.Time       `json:"check_in"`
	CheckOut          time.Time       `json:"check_out"`
	Nights            int32           `json:"nights"`
	GrossAmount       decimal.Decimal `json:"gross_amount"`
	PlatformFee       decimal.Decimal `json:"platform_fee"`
	NetAmount         decimal.Decimal `json:"net_amount"`
	Currency          string          `json:"currency"`
	Status            string          `json:"status"`
	CreatedAt         time.Time       `json:"created_at"`
}

type Expense struct {
	ID         uuid.UUID       `json:"id"`
	PropertyID uuid.UUID       `json:"property_id"`
	Category   string          `json:"category"`
	Description string         `json:"description"`
	Amount     decimal.Decimal `json:"amount"`
	Currency   string          `json:"currency"`
	Date       time.Time       `json:"date"`
	Recurring  bool            `json:"recurring"`
	CreatedAt  time.Time       `json:"created_at"`
}

type DashboardSummary struct {
	Month         string          `json:"month"`
	TotalIncome   decimal.Decimal `json:"total_income"`
	PlatformFees  decimal.Decimal `json:"platform_fees"`
	NetIncome     decimal.Decimal `json:"net_income"`
	Expenses      ExpenseBreakdown `json:"expenses"`
	Profit        decimal.Decimal `json:"profit"`
	OccupancyRate float64         `json:"occupancy_rate"`
	BookingsCount int             `json:"bookings_count"`
}

type ExpenseBreakdown struct {
	Electricity decimal.Decimal `json:"electricity"`
	Heating     decimal.Decimal `json:"heating"`
	Water       decimal.Decimal `json:"water"`
	Maintenance decimal.Decimal `json:"maintenance"`
	Cleaning    decimal.Decimal `json:"cleaning"`
	Other       decimal.Decimal `json:"other"`
	Total       decimal.Decimal `json:"total"`
}
