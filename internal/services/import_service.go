package services

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"math/big"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xuri/excelize/v2"

	"github.com/cristianbutiri/property-manager-v2/internal/database/sqlc"
)

type ImportService struct {
	queries *sqlc.Queries
}

func NewImportService(q *sqlc.Queries) *ImportService {
	return &ImportService{queries: q}
}

type ImportResult struct {
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
	Errors   int `json:"errors"`
}

func (s *ImportService) ImportFile(ctx context.Context, propertyID pgtype.UUID, platform string, filename string, reader io.Reader) (*ImportResult, error) {
	var records [][]string
	var err error

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".xls", ".xlsx":
		records, err = readExcel(reader)
	default:
		records, err = readCSV(reader)
	}
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return &ImportResult{}, nil
	}

	header := records[0]
	result := &ImportResult{}

	for _, row := range records[1:] {
		var params sqlc.UpsertBookingParams
		var parseErr error

		switch platform {
		case "booking":
			params, parseErr = parseBookingComRow(header, row, propertyID)
		case "airbnb":
			params, parseErr = parseAirbnbRow(header, row, propertyID)
		default:
			return nil, fmt.Errorf("unsupported platform: %s", platform)
		}

		if parseErr != nil {
			result.Errors++
			continue
		}

		_, err := s.queries.UpsertBooking(ctx, params)
		if err != nil {
			result.Errors++
			continue
		}
		result.Imported++
	}

	return result, nil
}

func readCSV(reader io.Reader) ([][]string, error) {
	csvReader := csv.NewReader(reader)
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}
	return records, nil
}

func readExcel(reader io.Reader) ([][]string, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("failed to read Excel sheet: %w", err)
	}
	return rows, nil
}

func findColumn(header []string, names ...string) int {
	for i, h := range header {
		normalized := strings.ToLower(strings.TrimSpace(h))
		for _, name := range names {
			if normalized == name {
				return i
			}
		}
	}
	return -1
}

func getField(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func parseAmount(s string) float64 {
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.TrimLeft(s, "€$£")
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func parseBookingComRow(header, row []string, propertyID pgtype.UUID) (sqlc.UpsertBookingParams, error) {
	refIdx := findColumn(header, "reservation number", "booking number", "reservation id", "book number")
	guestIdx := findColumn(header, "booker name", "guest name", "guest")
	checkInIdx := findColumn(header, "arrival", "check-in", "checkin")
	checkOutIdx := findColumn(header, "departure", "check-out", "checkout")
	amountIdx := findColumn(header, "total payment", "total price", "price", "amount", "total")
	commissionIdx := findColumn(header, "commission", "commission amount")
	statusIdx := findColumn(header, "status", "booking status")

	ref := getField(row, refIdx)
	if ref == "" {
		return sqlc.UpsertBookingParams{}, fmt.Errorf("missing booking reference")
	}

	checkIn, err := parseDate(getField(row, checkInIdx))
	if err != nil {
		return sqlc.UpsertBookingParams{}, fmt.Errorf("invalid check-in date: %w", err)
	}
	checkOut, err := parseDate(getField(row, checkOutIdx))
	if err != nil {
		return sqlc.UpsertBookingParams{}, fmt.Errorf("invalid check-out date: %w", err)
	}

	nights := int32(math.Round(checkOut.Sub(checkIn).Hours() / 24))
	gross := parseAmount(getField(row, amountIdx))
	commission := parseAmount(getField(row, commissionIdx))
	net := gross - commission

	status := "confirmed"
	if statusIdx >= 0 {
		s := strings.ToLower(getField(row, statusIdx))
		if strings.Contains(s, "cancel") {
			status = "cancelled"
		}
	}

	return sqlc.UpsertBookingParams{
		PropertyID:        propertyID,
		Platform:          "booking",
		PlatformReference: ref,
		GuestName:         getField(row, guestIdx),
		CheckIn:           pgtype.Date{Time: checkIn, Valid: true},
		CheckOut:          pgtype.Date{Time: checkOut, Valid: true},
		Nights:            nights,
		GrossAmount:       decimalFromFloat(gross),
		PlatformFee:       decimalFromFloat(commission),
		NetAmount:         decimalFromFloat(net),
		Currency:          "EUR",
		Status:            status,
	}, nil
}

func parseAirbnbRow(header, row []string, propertyID pgtype.UUID) (sqlc.UpsertBookingParams, error) {
	refIdx := findColumn(header, "confirmation code", "reservation code")
	guestIdx := findColumn(header, "guest", "guest name", "guest first name")
	checkInIdx := findColumn(header, "start date", "check-in", "date")
	checkOutIdx := findColumn(header, "end date", "check-out")
	amountIdx := findColumn(header, "amount", "total", "gross earnings")
	hostFeeIdx := findColumn(header, "host fee", "service fee", "host service fee")
	statusIdx := findColumn(header, "status", "type")
	nightsIdx := findColumn(header, "nights", "number of nights")

	ref := getField(row, refIdx)
	if ref == "" {
		return sqlc.UpsertBookingParams{}, fmt.Errorf("missing confirmation code")
	}

	checkIn, err := parseDate(getField(row, checkInIdx))
	if err != nil {
		return sqlc.UpsertBookingParams{}, fmt.Errorf("invalid start date: %w", err)
	}
	checkOut, err := parseDate(getField(row, checkOutIdx))
	if err != nil {
		return sqlc.UpsertBookingParams{}, fmt.Errorf("invalid end date: %w", err)
	}

	nights := int32(math.Round(checkOut.Sub(checkIn).Hours() / 24))
	if nightsIdx >= 0 {
		if n, err := strconv.Atoi(getField(row, nightsIdx)); err == nil {
			nights = int32(n)
		}
	}

	gross := parseAmount(getField(row, amountIdx))
	fee := parseAmount(getField(row, hostFeeIdx))
	net := gross - fee

	status := "confirmed"
	if statusIdx >= 0 {
		s := strings.ToLower(getField(row, statusIdx))
		if strings.Contains(s, "cancel") {
			status = "cancelled"
		}
	}

	return sqlc.UpsertBookingParams{
		PropertyID:        propertyID,
		Platform:          "airbnb",
		PlatformReference: ref,
		GuestName:         getField(row, guestIdx),
		CheckIn:           pgtype.Date{Time: checkIn, Valid: true},
		CheckOut:          pgtype.Date{Time: checkOut, Valid: true},
		Nights:            nights,
		GrossAmount:       decimalFromFloat(gross),
		PlatformFee:       decimalFromFloat(fee),
		NetAmount:         decimalFromFloat(net),
		Currency:          "EUR",
		Status:            status,
	}, nil
}

func parseDate(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"02/01/2006",
		"2006/01/02",
		"Jan 2, 2006",
		"2 Jan 2006",
		"02-01-2006",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized date format: %s", s)
}

func decimalFromFloat(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	scaled := int64(math.Round(f * 100))
	n.Valid = true
	n.Exp = -2
	n.Int = big.NewInt(scaled)
	return n
}
