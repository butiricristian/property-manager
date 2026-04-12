package services

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/extrame/xls"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xuri/excelize/v2"

	"github.com/cristianbutiri/property-manager-v2/internal/database/sqlc"
)

var errSkipRow = fmt.Errorf("row skipped")

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

type PayoutData struct {
	Gross      float64
	Commission float64
	Net        float64
}

func (s *ImportService) ImportFile(ctx context.Context, propertyID pgtype.UUID, platform string, filename string, reader io.Reader) (*ImportResult, error) {
	var records [][]string
	var err error

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".xlsx":
		records, err = readExcelX(reader)
	case ".xls":
		records, err = readXLS(reader)
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
	log.Printf("[import] file=%s format=%s platform=%s rows=%d", filename, ext, platform, len(records)-1)
	log.Printf("[import] header: %v", header)
	result := &ImportResult{}

	for i, row := range records[1:] {
		var params sqlc.UpsertBookingParams
		var parseErr error

		switch platform {
		case "booking":
			params, parseErr = parseBookingComRow(header, row, propertyID, nil)
		case "airbnb":
			params, parseErr = parseAirbnbRow(header, row, propertyID)
		default:
			return nil, fmt.Errorf("unsupported platform: %s", platform)
		}

		if parseErr != nil {
			if errors.Is(parseErr, errSkipRow) {
				result.Skipped++
			} else {
				log.Printf("[import] row %d parse error: %v | row data: %v", i+2, parseErr, row)
				result.Errors++
			}
			continue
		}

		_, err := s.queries.UpsertBooking(ctx, params)
		if err != nil {
			log.Printf("[import] row %d db error: %v", i+2, err)
			result.Errors++
			continue
		}
		result.Imported++
	}

	return result, nil
}

func (s *ImportService) ImportFileWithPayouts(ctx context.Context, propertyID pgtype.UUID, platform string, filename string, reader io.Reader, payoutsFilename string, payoutsReader io.Reader) (*ImportResult, error) {
	payouts, err := parsePayoutsFile(payoutsFilename, payoutsReader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse payouts file: %w", err)
	}
	log.Printf("[import] parsed %d payout entries", len(payouts))

	var records [][]string
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".xlsx":
		records, err = readExcelX(reader)
	case ".xls":
		records, err = readXLS(reader)
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
	log.Printf("[import] file=%s format=%s platform=%s rows=%d", filename, ext, platform, len(records)-1)
	result := &ImportResult{}

	for i, row := range records[1:] {
		var params sqlc.UpsertBookingParams
		var parseErr error

		switch platform {
		case "booking":
			params, parseErr = parseBookingComRow(header, row, propertyID, payouts)
		case "airbnb":
			params, parseErr = parseAirbnbRow(header, row, propertyID)
		default:
			return nil, fmt.Errorf("unsupported platform: %s", platform)
		}

		if parseErr != nil {
			if errors.Is(parseErr, errSkipRow) {
				result.Skipped++
			} else {
				log.Printf("[import] row %d parse error: %v | row data: %v", i+2, parseErr, row)
				result.Errors++
			}
			continue
		}

		_, err := s.queries.UpsertBooking(ctx, params)
		if err != nil {
			log.Printf("[import] row %d db error: %v", i+2, err)
			result.Errors++
			continue
		}
		result.Imported++
	}

	return result, nil
}

func parsePayoutsFile(filename string, reader io.Reader) (map[string]PayoutData, error) {
	var records [][]string
	var err error

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".xlsx":
		records, err = readExcelX(reader)
	case ".xls":
		records, err = readXLS(reader)
	default:
		records, err = readCSV(reader)
	}
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return map[string]PayoutData{}, nil
	}

	header := records[0]
	refIdx := findColumn(header, "reference number", "booking number", "reservation number")
	amountIdx := findColumn(header, "amount")
	commissionIdx := findColumn(header, "commission")
	netIdx := findColumn(header, "net")

	if refIdx < 0 {
		return nil, fmt.Errorf("payouts file missing 'Reference number' column")
	}

	payouts := make(map[string]PayoutData, len(records)-1)
	for _, row := range records[1:] {
		ref := getField(row, refIdx)
		if ref == "" {
			continue
		}
		payouts[ref] = PayoutData{
			Gross:      parseAmount(getField(row, amountIdx)),
			Commission: parseAmount(getField(row, commissionIdx)),
			Net:        parseAmount(getField(row, netIdx)),
		}
	}
	return payouts, nil
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

func readExcelX(reader io.Reader) ([][]string, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to open xlsx file: %w", err)
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("failed to read xlsx sheet: %w", err)
	}
	return rows, nil
}

func readXLS(reader io.Reader) ([][]string, error) {
	// extrame/xls requires a file on disk, so write to a temp file
	tmp, err := os.CreateTemp("", "import-*.xls")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := io.Copy(tmp, reader); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmp.Close()

	wb, err := xls.Open(tmp.Name(), "utf-8")
	if err != nil {
		return nil, fmt.Errorf("failed to open xls file: %w", err)
	}

	sheet := wb.GetSheet(0)
	if sheet == nil {
		return nil, fmt.Errorf("xls file has no sheets")
	}

	// First pass: determine max columns from all rows
	maxCols := 0
	for i := 0; i <= int(sheet.MaxRow); i++ {
		row := sheet.Row(i)
		if row == nil {
			continue
		}
		if lc := row.LastCol(); lc > maxCols {
			maxCols = lc
		}
	}

	log.Printf("[import/xls] sheets=%d maxRow=%d maxCols=%d", wb.NumSheets(), sheet.MaxRow, maxCols)

	// If LastCol didn't work, try reading with a fixed column count
	if maxCols == 0 {
		maxCols = 20 // reasonable upper bound for booking exports
	}

	var records [][]string
	for i := 0; i <= int(sheet.MaxRow); i++ {
		row := sheet.Row(i)
		if row == nil {
			continue
		}
		var cells []string
		allEmpty := true
		for j := 0; j < maxCols; j++ {
			val := row.Col(j)
			if val != "" {
				allEmpty = false
			}
			cells = append(cells, val)
		}
		if !allEmpty {
			records = append(records, cells)
		}
	}

	if len(records) > 0 {
		log.Printf("[import/xls] first row: %v", records[0])
	}

	return records, nil
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
	s = strings.TrimSpace(s)
	// Strip currency codes/symbols (prefix or suffix)
	for _, code := range []string{"EUR", "USD", "GBP", "RON", "LEI"} {
		s = strings.TrimPrefix(s, code)
		s = strings.TrimSuffix(s, code)
	}
	s = strings.ReplaceAll(s, " ", "")
	s = strings.TrimLeft(s, "€$£")
	s = strings.TrimRight(s, "€$£")

	// Detect format by finding the last separator character.
	// The last separator is always the decimal separator;
	// any earlier separators are thousands grouping.
	lastComma := strings.LastIndex(s, ",")
	lastDot := strings.LastIndex(s, ".")

	switch {
	case lastComma > lastDot:
		// European: "1.234,56" → remove dots, replace comma with dot
		s = strings.ReplaceAll(s, ".", "")
		s = strings.Replace(s, ",", ".", 1)
	case lastDot > lastComma:
		// US/UK: "1,234.56" → remove commas
		s = strings.ReplaceAll(s, ",", "")
	default:
		// No separators, or only one type — just replace comma with dot
		s = strings.ReplaceAll(s, ",", ".")
	}

	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func parseBookingComRow(header, row []string, propertyID pgtype.UUID, payouts map[string]PayoutData) (sqlc.UpsertBookingParams, error) {
	typeIdx := findColumn(header, "type")
	refIdx := findColumn(header, "reservation number", "booking number", "reservation id", "book number", "reference number")
	guestIdx := findColumn(header, "booker name", "guest name", "guest name(s)", "guest", "booked by")
	checkInIdx := findColumn(header, "arrival", "check-in", "checkin")
	checkOutIdx := findColumn(header, "departure", "check-out", "checkout")
	amountIdx := findColumn(header, "original amount", "total payment", "total price", "price", "amount", "total")
	priceIdx := findColumn(header, "final amount")
	commissionIdx := findColumn(header, "commission", "commission amount")
	commissionPctIdx := findColumn(header, "commission %", "commission%")
	netIdx := findColumn(header, "net")
	currencyIdx := findColumn(header, "currency")
	statusIdx := findColumn(header, "status", "booking status", "reservation status")
	nightsIdx := findColumn(header, "duration (nights)", "nights", "number of nights")

	// Detect reservations page format: no "type" column, has "commission %" column,
	// and no "final amount"/"net" columns (statement has those)
	isReservationsPage := typeIdx < 0 && commissionPctIdx >= 0 && priceIdx < 0 && netIdx < 0

	// Skip non-reservation rows in reservations statement exports
	if typeIdx >= 0 {
		rowType := strings.ToLower(getField(row, typeIdx))
		if rowType != "" && rowType != "reservation" {
			return sqlc.UpsertBookingParams{}, errSkipRow
		}
	}

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
	if nightsIdx >= 0 {
		if n, err := strconv.Atoi(getField(row, nightsIdx)); err == nil && n > 0 {
			nights = int32(n)
		}
	}

	var estimatedGross pgtype.Numeric // null by default

	gross := parseAmount(getField(row, amountIdx))
	price := gross
	if priceIdx >= 0 {
		if parsed := parseAmount(getField(row, priceIdx)); parsed != 0 {
			price = parsed
		}
	}
	commission := parseAmount(getField(row, commissionIdx))

	if isReservationsPage {
		// Price column is the room rate (no tourist tax included)
		// estimated_gross = price + (price / 1.11 × 0.03)  — 3% tourist tax on pre-VAT amount
		roomRate := price
		estGross := roomRate + (roomRate / 1.11 * 0.03)
		estGross = math.Round(estGross*100) / 100
		estimatedGross = decimalFromFloat(estGross)

		// Use estimated gross as best available gross_amount
		gross = estGross

		// Compute commission from commission % if commission amount column is missing
		if commission == 0 && commissionPctIdx >= 0 {
			pct := parseAmount(getField(row, commissionPctIdx))
			if pct > 0 {
				commission = math.Round(roomRate*pct) / 100
			}
		}
	}

	net := gross - commission
	if netIdx >= 0 {
		if parsed := parseAmount(getField(row, netIdx)); parsed != 0 {
			net = parsed
		}
	}

	if payouts != nil {
		if pd, ok := payouts[ref]; ok {
			gross = pd.Gross
			commission = pd.Commission
			net = pd.Net
			price = gross
		}
	}

	status := "confirmed"
	if statusIdx >= 0 {
		s := strings.ToLower(getField(row, statusIdx))
		if strings.Contains(s, "cancel") {
			status = "cancelled"
		}
	}

	// Try to detect currency from the amount field if no dedicated currency column
	currency := getField(row, currencyIdx)
	if currency == "" {
		currency = detectCurrency(getField(row, amountIdx))
	}

	return sqlc.UpsertBookingParams{
		PropertyID:           propertyID,
		Platform:             "booking",
		PlatformReference:    ref,
		GuestName:            getField(row, guestIdx),
		CheckIn:              pgtype.Date{Time: checkIn, Valid: true},
		CheckOut:             pgtype.Date{Time: checkOut, Valid: true},
		Nights:               nights,
		GrossAmount:          decimalFromFloat(gross),
		PlatformFee:          decimalFromFloat(commission),
		NetAmount:            decimalFromFloat(net),
		Currency:             currencyOrDefault(currency, "EUR"),
		Status:               status,
		Price:                decimalFromFloat(price),
		EstimatedGrossAmount: estimatedGross,
	}, nil
}

func detectCurrency(amountStr string) string {
	s := strings.ToUpper(strings.TrimSpace(amountStr))
	if strings.Contains(s, "€") || strings.Contains(s, "EUR") {
		return "EUR"
	}
	if strings.Contains(s, "$") || strings.Contains(s, "USD") {
		return "USD"
	}
	if strings.Contains(s, "£") || strings.Contains(s, "GBP") {
		return "GBP"
	}
	if strings.Contains(s, "RON") || strings.Contains(s, "LEI") {
		return "RON"
	}
	return ""
}

func currencyOrDefault(val, fallback string) string {
	val = strings.TrimSpace(strings.ToUpper(val))
	if val == "" {
		return fallback
	}
	return val
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
		Price:             decimalFromFloat(gross),
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
		"2 January 2006",
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
