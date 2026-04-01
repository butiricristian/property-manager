package handlers

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/cristianbutiri/property-manager-v2/internal/auth"
)

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("error encoding response: %v", err)
		}
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func respondInternalError(w http.ResponseWriter, message string, err error) {
	log.Printf("ERROR: %s: %v", message, err)
	respondJSON(w, http.StatusInternalServerError, map[string]string{"error": message})
}

func decodeJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func parseUUID(s string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	err := uuid.Scan(s)
	return uuid, err
}

func getURLParamUUID(r *http.Request, param string) (pgtype.UUID, error) {
	return parseUUID(chi.URLParam(r, param))
}

func parseMonthDates(month string) (startDate, endDate pgtype.Date, err error) {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return startDate, endDate, err
	}
	startDate = pgtype.Date{Time: t, Valid: true}
	endDate = pgtype.Date{Time: t.AddDate(0, 1, 0), Valid: true}
	return startDate, endDate, nil
}

func getMonthParam(r *http.Request) string {
	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	return month
}

func numericToDecimal(n pgtype.Numeric) decimal.Decimal {
	if !n.Valid {
		return decimal.Zero
	}
	f, _ := n.Float64Value()
	return decimal.NewFromFloat(f.Float64)
}

func decimalToNumeric(d decimal.Decimal) pgtype.Numeric {
	// Convert to cents (2 decimal places) for storage
	cents := d.Mul(decimal.NewFromInt(100)).IntPart()
	var n pgtype.Numeric
	n.Int = big.NewInt(cents)
	n.Exp = -2
	n.Valid = true
	return n
}

func getUserID(r *http.Request) (pgtype.UUID, error) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		return pgtype.UUID{}, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

func uuidBytesToString(u pgtype.UUID) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", u.Bytes[0:4], u.Bytes[4:6], u.Bytes[6:8], u.Bytes[8:10], u.Bytes[10:16])
}

func ServeStaticFiles(r chi.Router) {
	distPath := "web/dist"
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		log.Println("web/dist not found, skipping static file serving")
		return
	}
	fsys := os.DirFS(distPath)
	fileServer := http.FileServer(http.FS(fsys))

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file directly
		path := r.URL.Path
		if path == "/" {
			path = "index.html"
		} else {
			path = path[1:] // Remove leading slash
		}

		// Check if file exists
		if _, err := fs.Stat(fsys, path); os.IsNotExist(err) {
			// SPA fallback: serve index.html for non-API routes
			http.ServeFile(w, r, distPath+"/index.html")
			return
		}

		fileServer.ServeHTTP(w, r)
	})
}
