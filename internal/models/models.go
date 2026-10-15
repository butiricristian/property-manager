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
	Price             decimal.Decimal `json:"price"`
	PlatformFee       decimal.Decimal `json:"platform_fee"`
	NetAmount         decimal.Decimal `json:"net_amount"`
	Currency          string          `json:"currency"`
	Status            string          `json:"status"`
	CreatedAt         time.Time       `json:"created_at"`
}

type Expense struct {
	ID           uuid.UUID       `json:"id"`
	PropertyID   uuid.UUID       `json:"property_id"`
	Category     string          `json:"category"`
	Description  string          `json:"description"`
	Amount       decimal.Decimal `json:"amount"`
	Currency     string          `json:"currency"`
	Date         time.Time       `json:"date"`
	EndDate      *time.Time      `json:"end_date"`
	Recurring    bool            `json:"recurring"`
	IsPercentage bool            `json:"is_percentage"`
	Percentage   decimal.Decimal `json:"percentage"`
	CreatedAt    time.Time       `json:"created_at"`
}

type Integration struct {
	ID            uuid.UUID         `json:"id"`
	UserID        uuid.UUID         `json:"user_id"`
	Platform      string            `json:"platform"`
	Credentials   map[string]string `json:"credentials"`
	Enabled       bool              `json:"enabled"`
	LastSyncedAt  *time.Time        `json:"last_synced_at"`
	LastSyncError *string           `json:"last_sync_error"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type SyncResult struct {
	Imported int    `json:"imported"`
	Updated  int    `json:"updated"`
	Errors   int    `json:"errors"`
	Message  string `json:"message"`
}

type TaxEstimate struct {
	Name   string          `json:"name"`
	Amount decimal.Decimal `json:"amount"`
}

type DashboardSummary struct {
	Month         string          `json:"month"`
	Currency      string          `json:"currency"`
	TotalIncome   decimal.Decimal `json:"total_income"`
	PlatformFees  decimal.Decimal `json:"platform_fees"`
	NetIncome     decimal.Decimal `json:"net_income"`
	Expenses      ExpenseBreakdown `json:"expenses"`
	Profit        decimal.Decimal `json:"profit"`
	Taxes            []TaxEstimate   `json:"taxes"`
	TotalTax         decimal.Decimal `json:"total_tax"`
	ProfitAfterTax   decimal.Decimal `json:"profit_after_tax"`
	OccupancyRate          float64         `json:"occupancy_rate"`
	EstimatedOccupancyRate *float64        `json:"estimated_occupancy_rate,omitempty"`
	BookingsCount    int             `json:"bookings_count"`
	ADR              decimal.Decimal `json:"adr"`
	RevPAR           decimal.Decimal `json:"revpar"`
	AvgStayLength    float64         `json:"avg_stay_length"`
	AvgLeadTimeDays  float64         `json:"avg_lead_time_days"`
	CancellationRate float64         `json:"cancellation_rate"`
}

type ExpenseBreakdown struct {
	Electricity   decimal.Decimal `json:"electricity"`
	Heating       decimal.Decimal `json:"heating"`
	Water         decimal.Decimal `json:"water"`
	Internet      decimal.Decimal `json:"internet"`
	Maintenance   decimal.Decimal `json:"maintenance"`
	Cleaning      decimal.Decimal `json:"cleaning"`
	Other         decimal.Decimal `json:"other"`
	PartnerPayout decimal.Decimal `json:"partner_payout"`
	Total         decimal.Decimal `json:"total"`
}
