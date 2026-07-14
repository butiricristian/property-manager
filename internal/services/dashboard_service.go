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
	queries    *sqlc.Queries
	exchangeSvc *ExchangeService
}

func NewDashboardService(q *sqlc.Queries, exchangeSvc *ExchangeService) *DashboardService {
	return &DashboardService{queries: q, exchangeSvc: exchangeSvc}
}

func (s *DashboardService) GetPropertyDashboard(ctx context.Context, propertyID pgtype.UUID, startDate, endDate pgtype.Date, currency string) (*models.DashboardSummary, error) {
	month := startDate.Time.Format("2006-01")

	baseCurrency := s.exchangeSvc.BaseCurrency()
	if currency != "" {
		baseCurrency = currency
	}

	bookingRows, err := s.queries.GetBookingsSummaryByCurrency(ctx, sqlc.GetBookingsSummaryByCurrencyParams{
		PropertyID: propertyID,
		StartDate:  startDate,
		EndDate:    endDate,
	})
	if err != nil {
		return nil, err
	}

	totalIncome := decimal.Zero
	platformFees := decimal.Zero
	netIncome := decimal.Zero
	bookingsCount := 0
	totalNights := 0

	for _, row := range bookingRows {
		currency := row.Currency
		income, err := s.exchangeSvc.Convert(ctx, numericToDecimal(row.TotalIncome), currency, baseCurrency, month)
		if err != nil {
			return nil, err
		}
		fees, err := s.exchangeSvc.Convert(ctx, numericToDecimal(row.PlatformFees), currency, baseCurrency, month)
		if err != nil {
			return nil, err
		}
		net, err := s.exchangeSvc.Convert(ctx, numericToDecimal(row.NetIncome), currency, baseCurrency, month)
		if err != nil {
			return nil, err
		}
		totalIncome = totalIncome.Add(income)
		platformFees = platformFees.Add(fees)
		netIncome = netIncome.Add(net)
		bookingsCount += int(row.BookingsCount)
		totalNights += int(row.TotalNights)
	}

	expenseRows, err := s.queries.GetExpensesByCategoryCurrency(ctx, sqlc.GetExpensesByCategoryCurrencyParams{
		PropertyID: propertyID,
		StartDate:  startDate,
		EndDate:    endDate,
	})
	if err != nil {
		return nil, err
	}

	expenses, err := s.aggregateExpenses(ctx, expenseRows, baseCurrency, month)
	if err != nil {
		return nil, err
	}

	pctRows, err := s.queries.GetPercentageExpensesByProperty(ctx, sqlc.GetPercentageExpensesByPropertyParams{
		PropertyID: propertyID,
		StartDate:  startDate,
		EndDate:    endDate,
	})
	if err != nil {
		return nil, err
	}
	s.applyPercentageExpenses(pctRows, netIncome, &expenses)

	analytics, err := s.queries.GetBookingAnalyticsByProperty(ctx, sqlc.GetBookingAnalyticsByPropertyParams{
		PropertyID: propertyID,
		StartDate:  startDate,
		EndDate:    endDate,
	})
	if err != nil {
		return nil, err
	}

	taxRules, err := s.queries.ListTaxRulesByProperty(ctx, propertyID)
	if err != nil {
		return nil, err
	}

	profit := netIncome.Sub(expenses.Total)
	taxes, totalTax := s.calculateTaxes(ctx, taxRules, profit, baseCurrency, month, startDate, endDate)

	summary := buildSummary(month, baseCurrency, totalIncome, platformFees, netIncome,
		bookingsCount, totalNights, startDate, endDate, expenses,
		analytics.AvgLeadTimeDays, int(analytics.CancelledCount), int(analytics.TotalCount))
	summary.Taxes = taxes
	summary.TotalTax = totalTax
	summary.ProfitAfterTax = profit.Sub(totalTax)
	return summary, nil
}

func (s *DashboardService) GetOverallDashboard(ctx context.Context, userID pgtype.UUID, startDate, endDate pgtype.Date, currency string) (*models.DashboardSummary, error) {
	month := startDate.Time.Format("2006-01")

	baseCurrency := s.exchangeSvc.BaseCurrency()
	if currency != "" {
		baseCurrency = currency
	}

	bookingRows, err := s.queries.GetAllBookingsSummaryByCurrencyForUser(ctx, sqlc.GetAllBookingsSummaryByCurrencyForUserParams{
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return nil, err
	}

	totalIncome := decimal.Zero
	platformFees := decimal.Zero
	netIncome := decimal.Zero
	bookingsCount := 0
	totalNights := 0

	for _, row := range bookingRows {
		currency := row.Currency
		income, err := s.exchangeSvc.Convert(ctx, numericToDecimal(row.TotalIncome), currency, baseCurrency, month)
		if err != nil {
			return nil, err
		}
		fees, err := s.exchangeSvc.Convert(ctx, numericToDecimal(row.PlatformFees), currency, baseCurrency, month)
		if err != nil {
			return nil, err
		}
		net, err := s.exchangeSvc.Convert(ctx, numericToDecimal(row.NetIncome), currency, baseCurrency, month)
		if err != nil {
			return nil, err
		}
		totalIncome = totalIncome.Add(income)
		platformFees = platformFees.Add(fees)
		netIncome = netIncome.Add(net)
		bookingsCount += int(row.BookingsCount)
		totalNights += int(row.TotalNights)
	}

	expenseRows, err := s.queries.GetAllExpensesByCategoryCurrencyForUser(ctx, sqlc.GetAllExpensesByCategoryCurrencyForUserParams{
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return nil, err
	}

	expenses, err := s.aggregateUserExpenses(ctx, expenseRows, baseCurrency, month)
	if err != nil {
		return nil, err
	}

	pctRows, err := s.queries.GetAllPercentageExpensesForUser(ctx, sqlc.GetAllPercentageExpensesForUserParams{
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return nil, err
	}
	s.applyPercentageExpenses(pctRows, netIncome, &expenses)

	analytics, err := s.queries.GetBookingAnalyticsForUser(ctx, sqlc.GetBookingAnalyticsForUserParams{
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return nil, err
	}

	taxRules, err := s.queries.ListTaxRulesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	profit := netIncome.Sub(expenses.Total)
	taxes, totalTax := s.calculateTaxes(ctx, taxRules, profit, baseCurrency, month, startDate, endDate)

	summary := buildSummary(month, baseCurrency, totalIncome, platformFees, netIncome,
		bookingsCount, totalNights, startDate, endDate, expenses,
		analytics.AvgLeadTimeDays, int(analytics.CancelledCount), int(analytics.TotalCount))
	summary.Taxes = taxes
	summary.TotalTax = totalTax
	summary.ProfitAfterTax = profit.Sub(totalTax)
	return summary, nil
}

func (s *DashboardService) calculateTaxes(ctx context.Context, rules []sqlc.TaxRule, profit decimal.Decimal, baseCurrency, month string, startDate, endDate pgtype.Date) ([]models.TaxEstimate, decimal.Decimal) {
	hundred := decimal.NewFromInt(100)
	twelve := decimal.NewFromInt(12)

	// Calculate number of months in the range for scaling fixed taxes
	sy, sm, _ := startDate.Time.Date()
	ey, em, _ := endDate.Time.Date()
	numMonths := (ey-sy)*12 + int(em) - int(sm)
	if numMonths < 1 {
		numMonths = 1
	}
	monthsDecimal := decimal.NewFromInt(int64(numMonths))

	var taxes []models.TaxEstimate
	totalTax := decimal.Zero

	for _, rule := range rules {
		var amount decimal.Decimal
		if rule.Type == "percentage" {
			pct := numericToDecimal(rule.Percentage)
			amount = pct.Div(hundred).Mul(profit).Round(2)
		} else {
			fixedAnnual := numericToDecimal(rule.FixedAmount)
			monthly := fixedAnnual.Div(twelve).Round(2)
			scaled := monthly.Mul(monthsDecimal)
			converted, err := s.exchangeSvc.Convert(ctx, scaled, rule.Currency, baseCurrency, month)
			if err != nil {
				amount = scaled
			} else {
				amount = converted
			}
		}
		if amount.IsPositive() {
			taxes = append(taxes, models.TaxEstimate{
				Name:   rule.Name,
				Amount: amount,
			})
			totalTax = totalTax.Add(amount)
		}
	}

	if taxes == nil {
		taxes = []models.TaxEstimate{}
	}
	return taxes, totalTax
}

func (s *DashboardService) aggregateExpenses(ctx context.Context, rows []sqlc.GetExpensesByCategoryCurrencyRow, baseCurrency, month string) (models.ExpenseBreakdown, error) {
	expenses := models.ExpenseBreakdown{}
	totalExpenses := decimal.Zero

	for _, row := range rows {
		amount := numericToDecimal(row.Total)
		converted, err := s.exchangeSvc.Convert(ctx, amount, row.Currency, baseCurrency, month)
		if err != nil {
			return expenses, err
		}
		totalExpenses = totalExpenses.Add(converted)
		setCategoryDecimal(&expenses, row.Category, converted)
	}

	expenses.Total = totalExpenses
	return expenses, nil
}

func (s *DashboardService) aggregateUserExpenses(ctx context.Context, rows []sqlc.GetAllExpensesByCategoryCurrencyForUserRow, baseCurrency, month string) (models.ExpenseBreakdown, error) {
	expenses := models.ExpenseBreakdown{}
	totalExpenses := decimal.Zero

	for _, row := range rows {
		amount := numericToDecimal(row.Total)
		converted, err := s.exchangeSvc.Convert(ctx, amount, row.Currency, baseCurrency, month)
		if err != nil {
			return expenses, err
		}
		totalExpenses = totalExpenses.Add(converted)
		setCategoryDecimal(&expenses, row.Category, converted)
	}

	expenses.Total = totalExpenses
	return expenses, nil
}

func buildSummary(month, currency string, totalIncome, platformFees, netIncome decimal.Decimal,
	bookingsCount, totalNights int, startDate, endDate pgtype.Date, expenses models.ExpenseBreakdown,
	avgLeadTimeDays float64, cancelledCount, totalBookingCount int) *models.DashboardSummary {

	profit := netIncome.Sub(expenses.Total)

	daysInRange := endDate.Time.Sub(startDate.Time).Hours() / 24
	occupancyRate := 0.0
	if daysInRange > 0 {
		occupancyRate = math.Min(float64(totalNights)/daysInRange, 1.0)
	}

	// Estimated occupancy: extrapolate from elapsed days when the range hasn't ended yet
	var estimatedOccupancy *float64
	now := time.Now()
	if now.Before(endDate.Time) && now.After(startDate.Time) {
		daysElapsed := now.Sub(startDate.Time).Hours() / 24
		if daysElapsed >= 1 && daysInRange > 0 {
			est := math.Min(float64(totalNights)/daysElapsed, 1.0)
			est = math.Round(est*100) / 100
			estimatedOccupancy = &est
		}
	}

	// ADR = gross income / total nights
	adr := decimal.Zero
	if totalNights > 0 {
		adr = totalIncome.Div(decimal.NewFromInt(int64(totalNights))).Round(2)
	}

	// RevPAR = gross income / days in month
	revpar := decimal.Zero
	if daysInRange > 0 {
		revpar = totalIncome.Div(decimal.NewFromFloat(daysInRange)).Round(2)
	}

	// Avg stay length = total nights / bookings count (non-cancelled)
	avgStayLength := 0.0
	if bookingsCount > 0 {
		avgStayLength = math.Round(float64(totalNights)/float64(bookingsCount)*10) / 10
	}

	// Cancellation rate = cancelled / total bookings (including cancelled)
	cancellationRate := 0.0
	if totalBookingCount > 0 {
		cancellationRate = math.Round(float64(cancelledCount)/float64(totalBookingCount)*1000) / 1000
	}

	return &models.DashboardSummary{
		Month:                  month,
		Currency:               currency,
		TotalIncome:            totalIncome,
		PlatformFees:           platformFees,
		NetIncome:              netIncome,
		Expenses:               expenses,
		Profit:                 profit,
		OccupancyRate:          math.Round(occupancyRate*100) / 100,
		EstimatedOccupancyRate: estimatedOccupancy,
		BookingsCount:          bookingsCount,
		ADR:                    adr,
		RevPAR:                 revpar,
		AvgStayLength:          avgStayLength,
		AvgLeadTimeDays:        math.Round(avgLeadTimeDays*10) / 10,
		CancellationRate:       cancellationRate,
	}
}

func setCategoryDecimal(expenses *models.ExpenseBreakdown, category string, value decimal.Decimal) {
	switch category {
	case "electricity":
		expenses.Electricity = expenses.Electricity.Add(value)
	case "heating":
		expenses.Heating = expenses.Heating.Add(value)
	case "water":
		expenses.Water = expenses.Water.Add(value)
	case "internet":
		expenses.Internet = expenses.Internet.Add(value)
	case "maintenance":
		expenses.Maintenance = expenses.Maintenance.Add(value)
	case "cleaning":
		expenses.Cleaning = expenses.Cleaning.Add(value)
	case "other":
		expenses.Other = expenses.Other.Add(value)
	case "partner_payout":
		expenses.PartnerPayout = expenses.PartnerPayout.Add(value)
	}
}

func (s *DashboardService) applyPercentageExpenses(rows []sqlc.Expense, netIncome decimal.Decimal, expenses *models.ExpenseBreakdown) {
	hundred := decimal.NewFromInt(100)
	for _, row := range rows {
		pct := numericToDecimal(row.Percentage)
		amount := pct.Div(hundred).Mul(netIncome).Round(2)
		setCategoryDecimal(expenses, row.Category, amount)
		expenses.Total = expenses.Total.Add(amount)
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
