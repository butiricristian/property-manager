package services

import (
	"context"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/cristianbutiri/property-manager-v2/internal/database/sqlc"
	"github.com/cristianbutiri/property-manager-v2/internal/models"
)

type DashboardService struct {
	queries *sqlc.Queries
}

func NewDashboardService(q *sqlc.Queries) *DashboardService {
	return &DashboardService{queries: q}
}

func (s *DashboardService) GetPropertyDashboard(ctx context.Context, propertyID pgtype.UUID, month string) (*models.DashboardSummary, error) {
	startDate, endDate, err := parseMonthRange(month)
	if err != nil {
		return nil, err
	}

	bookingSummary, err := s.queries.GetBookingsSummary(ctx, sqlc.GetBookingsSummaryParams{
		PropertyID: propertyID,
		StartDate:  startDate,
		EndDate:    endDate,
	})
	if err != nil {
		return nil, err
	}

	expenseRows, err := s.queries.GetExpensesByCategory(ctx, sqlc.GetExpensesByCategoryParams{
		PropertyID: propertyID,
		StartDate:  startDate,
		EndDate:    endDate,
	})
	if err != nil {
		return nil, err
	}

	return buildSummary(month, bookingSummary.TotalIncome, bookingSummary.PlatformFees,
		bookingSummary.NetIncome, int(bookingSummary.BookingsCount),
		int(bookingSummary.TotalNights), startDate, expenseRows), nil
}

func (s *DashboardService) GetOverallDashboard(ctx context.Context, month string) (*models.DashboardSummary, error) {
	startDate, endDate, err := parseMonthRange(month)
	if err != nil {
		return nil, err
	}

	bookingSummary, err := s.queries.GetAllBookingsSummary(ctx, sqlc.GetAllBookingsSummaryParams{
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return nil, err
	}

	expenseRows, err := s.queries.GetAllExpensesByCategory(ctx, sqlc.GetAllExpensesByCategoryParams{
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return nil, err
	}

	return buildSummary(month, bookingSummary.TotalIncome, bookingSummary.PlatformFees,
		bookingSummary.NetIncome, int(bookingSummary.BookingsCount),
		int(bookingSummary.TotalNights), startDate, expenseRows), nil
}

func parseMonthRange(month string) (pgtype.Date, pgtype.Date, error) {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return pgtype.Date{}, pgtype.Date{}, err
	}
	startDate := pgtype.Date{Time: t, Valid: true}
	endDate := pgtype.Date{Time: t.AddDate(0, 1, 0), Valid: true}
	return startDate, endDate, nil
}

type expenseCategoryRow interface {
	GetCategory() string
	GetTotal() pgtype.Numeric
}

func buildSummary(month string, totalIncome, platformFees, netIncome pgtype.Numeric,
	bookingsCount, totalNights int, startDate pgtype.Date, expenseRows interface{}) *models.DashboardSummary {

	expenses := models.ExpenseBreakdown{}
	totalExpenses := 0.0

	switch rows := expenseRows.(type) {
	case []sqlc.GetExpensesByCategoryRow:
		for _, row := range rows {
			val := numericToFloat(row.Total)
			totalExpenses += val
			setCategoryValue(&expenses, row.Category, val)
		}
	case []sqlc.GetAllExpensesByCategoryRow:
		for _, row := range rows {
			val := numericToFloat(row.Total)
			totalExpenses += val
			setCategoryValue(&expenses, row.Category, val)
		}
	}

	expenses.Total = floatToDecimal(totalExpenses)

	netIncomeFloat := numericToFloat(netIncome)
	profit := netIncomeFloat - totalExpenses

	daysInMonth := startDate.Time.AddDate(0, 1, 0).Sub(startDate.Time).Hours() / 24
	occupancyRate := 0.0
	if daysInMonth > 0 {
		occupancyRate = math.Min(float64(totalNights)/daysInMonth, 1.0)
	}

	return &models.DashboardSummary{
		Month:         month,
		TotalIncome:   numericToDecimal(totalIncome),
		PlatformFees:  numericToDecimal(platformFees),
		NetIncome:     numericToDecimal(netIncome),
		Expenses:      expenses,
		Profit:        floatToDecimal(profit),
		OccupancyRate: math.Round(occupancyRate*100) / 100,
		BookingsCount: bookingsCount,
	}
}

func setCategoryValue(expenses *models.ExpenseBreakdown, category string, value float64) {
	d := floatToDecimal(value)
	switch category {
	case "electricity":
		expenses.Electricity = d
	case "heating":
		expenses.Heating = d
	case "water":
		expenses.Water = d
	case "maintenance":
		expenses.Maintenance = d
	case "cleaning":
		expenses.Cleaning = d
	case "other":
		expenses.Other = d
	}
}

func numericToFloat(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, _ := n.Float64Value()
	return f.Float64
}

func numericToDecimal(n pgtype.Numeric) decimal.Decimal {
	return decimal.NewFromFloat(numericToFloat(n))
}

func floatToDecimal(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}
