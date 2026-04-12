package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/oauth2"

	"github.com/cristianbutiri/property-manager-v2/internal/auth"
	"github.com/cristianbutiri/property-manager-v2/internal/database/sqlc"
)

type AuthHandler struct {
	queries     *sqlc.Queries
	jwtSvc      *auth.JWTService
	oauthConfig *oauth2.Config
}

func NewAuthHandler(q *sqlc.Queries, jwtSvc *auth.JWTService, oauthConfig *oauth2.Config) *AuthHandler {
	return &AuthHandler{queries: q, jwtSvc: jwtSvc, oauthConfig: oauthConfig}
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string   `json:"token"`
	User  userInfo `json:"user"`
}

type userInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	UserType  string `json:"user_type"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if !emailRegex.MatchString(req.Email) {
		respondError(w, http.StatusBadRequest, "invalid email format")
		return
	}
	if len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		respondInternalError(w, "failed to hash password", err)
		return
	}

	user, err := h.queries.CreateUser(r.Context(), sqlc.CreateUserParams{
		Email:        req.Email,
		PasswordHash: pgtype.Text{String: hash, Valid: true},
	})
	if err != nil {
		respondError(w, http.StatusConflict, "email already registered")
		return
	}

	uid, _ := uuid.Parse(uuidBytesToString(user.ID))
	token, err := h.jwtSvc.GenerateToken(uid, user.Email)
	if err != nil {
		respondInternalError(w, "failed to generate token", err)
		return
	}

	respondJSON(w, http.StatusCreated, authResponse{
		Token: token,
		User: userInfo{
			ID:        uuidBytesToString(user.ID),
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			UserType:  "free",
		},
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if !user.PasswordHash.Valid {
		respondError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if err := auth.VerifyPassword(user.PasswordHash.String, req.Password); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	uid, _ := uuid.Parse(uuidBytesToString(user.ID))
	token, err := h.jwtSvc.GenerateToken(uid, user.Email)
	if err != nil {
		respondInternalError(w, "failed to generate token", err)
		return
	}

	respondJSON(w, http.StatusOK, authResponse{
		Token: token,
		User: userInfo{
			ID:        uuidBytesToString(user.ID),
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			Name:      user.Name,
			AvatarURL: user.AvatarUrl,
			UserType:  user.UserType,
		},
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.queries.GetUserByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respondJSON(w, http.StatusOK, userInfo{
		ID:        uuidBytesToString(user.ID),
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		Name:      user.Name,
		AvatarURL: user.AvatarUrl,
		UserType:  user.UserType,
	})
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.oauthConfig == nil {
		respondError(w, http.StatusNotFound, "Google OAuth not configured")
		return
	}

	state, err := generateRandomState()
	if err != nil {
		respondInternalError(w, "failed to generate state", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	url := h.oauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if h.oauthConfig == nil {
		respondError(w, http.StatusNotFound, "Google OAuth not configured")
		return
	}

	// Validate state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value == "" {
		respondHTML(w, http.StatusBadRequest, errorHTML("Missing OAuth state"))
		return
	}
	if r.URL.Query().Get("state") != stateCookie.Value {
		respondHTML(w, http.StatusBadRequest, errorHTML("Invalid OAuth state"))
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Check for error from Google
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		respondHTML(w, http.StatusBadRequest, errorHTML("Google login denied"))
		return
	}

	// Exchange code for token
	code := r.URL.Query().Get("code")
	oauthToken, err := h.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		respondHTML(w, http.StatusBadRequest, errorHTML("Failed to exchange code"))
		return
	}

	// Fetch Google user info
	gUser, err := fetchGoogleUserInfo(r.Context(), h.oauthConfig, oauthToken)
	if err != nil {
		respondHTML(w, http.StatusInternalServerError, errorHTML("Failed to fetch Google user info"))
		return
	}

	// Find or create user
	ctx := r.Context()
	googleID := pgtype.Text{String: gUser.ID, Valid: true}

	// 1. Try find by google_id
	log.Printf("Searching user by Google ID")
	dbUser, err := h.queries.GetUserByGoogleID(ctx, googleID)
	if err == nil {
		log.Printf("Existing user found with Google ID")
		// Existing Google user
		h.respondWithTokenHTML(w, r, dbUser.ID, dbUser.Email, dbUser.Name, dbUser.AvatarUrl, dbUser.CreatedAt)
		return
	}

	// 2. Try find by email — link Google account
	log.Printf("Searching user by email")
	existingUser, err := h.queries.GetUserByEmail(ctx, gUser.Email)
	if err == nil {
		log.Printf("Linking Google account to existing user")
		name := gUser.Name
		if existingUser.Name != "" {
			name = existingUser.Name
		}
		_ = h.queries.LinkGoogleAccount(ctx, sqlc.LinkGoogleAccountParams{
			ID:        existingUser.ID,
			GoogleID:  googleID,
			Name:      name,
			AvatarUrl: gUser.Picture,
		})
		h.respondWithTokenHTML(w, r, existingUser.ID, existingUser.Email, name, gUser.Picture, existingUser.CreatedAt)
		return
	}

	// 3. Create new user
	log.Printf("Creating new user")
	newUser, err := h.queries.CreateGoogleUser(ctx, sqlc.CreateGoogleUserParams{
		Email:     gUser.Email,
		GoogleID:  googleID,
		Name:      gUser.Name,
		AvatarUrl: gUser.Picture,
	})
	if err != nil {
		respondHTML(w, http.StatusInternalServerError, errorHTML("Failed to create user"))
		return
	}

	h.respondWithTokenHTML(w, r, newUser.ID, newUser.Email, newUser.Name, newUser.AvatarUrl, newUser.CreatedAt)
}

func (h *AuthHandler) respondWithTokenHTML(w http.ResponseWriter, r *http.Request, id pgtype.UUID, email, name, avatarURL string, createdAt pgtype.Timestamptz) {
	uid, _ := uuid.Parse(uuidBytesToString(id))
	token, err := h.jwtSvc.GenerateToken(uid, email)
	if err != nil {
		respondHTML(w, http.StatusInternalServerError, errorHTML("Failed to generate token"))
		return
	}

	// Redirect to frontend with token — the frontend will read it from the URL.
	frontendURL := "http://localhost:5173"
	if r.Header.Get("X-Forwarded-Host") != "" {
		frontendURL = ""
	}
	http.Redirect(w, r, fmt.Sprintf("%s/auth/callback?token=%s", frontendURL, token), http.StatusTemporaryRedirect)
}

type googleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func fetchGoogleUserInfo(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (*googleUserInfo, error) {
	client := cfg.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var info googleUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (h *AuthHandler) Guest(w http.ResponseWriter, r *http.Request) {
	guestEmail := fmt.Sprintf("guest-%s@guest.local", uuid.New().String())

	user, err := h.queries.CreateGuestUser(r.Context(), guestEmail)
	if err != nil {
		respondInternalError(w, "failed to create guest user", err)
		return
	}

	// Create a demo property for the guest
	property, err := h.queries.CreateProperty(r.Context(), sqlc.CreatePropertyParams{
		Name:    "Demo Property",
		Address: "123 Demo Street",
		UserID:  user.ID,
	})
	if err != nil {
		respondInternalError(w, "failed to create demo property", err)
		return
	}

	// Seed demo data so guests see a populated dashboard
	if err := h.seedGuestData(r.Context(), property.ID); err != nil {
		log.Printf("WARNING: failed to seed guest demo data: %v", err)
	}

	uid, _ := uuid.Parse(uuidBytesToString(user.ID))
	token, err := h.jwtSvc.GenerateToken(uid, user.Email)
	if err != nil {
		respondInternalError(w, "failed to generate token", err)
		return
	}

	respondJSON(w, http.StatusCreated, authResponse{
		Token: token,
		User: userInfo{
			ID:        uuidBytesToString(user.ID),
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			UserType:  user.UserType,
		},
	})
}

func demoNumeric(amount float64) pgtype.Numeric {
	cents := int64(amount * 100)
	var n pgtype.Numeric
	n.Int = big.NewInt(cents)
	n.Exp = -2
	n.Valid = true
	return n
}

func demoDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

func (h *AuthHandler) seedGuestData(ctx context.Context, propertyID pgtype.UUID) error {
	now := time.Now()
	year, monthNum, _ := now.Date()
	loc := now.Location()
	monthStart := time.Date(year, monthNum, 1, 0, 0, 0, 0, loc)

	// Unique suffix so demo refs don't collide across guest accounts
	suffix := uuid.New().String()[:8]

	type booking struct {
		platform  string
		ref       string
		guest     string
		checkIn   time.Time
		checkOut  time.Time
		nights    int32
		gross     float64
		price     float64
		fee       float64
		net       float64
		currency  string
		status    string
	}

	bookings := []booking{
		{
			platform: "airbnb", ref: "HM-" + suffix + "-1", guest: "Alice Johnson",
			checkIn: monthStart.AddDate(0, 0, 1), checkOut: monthStart.AddDate(0, 0, 4),
			nights: 3, gross: 420, price: 420, fee: 63, net: 357, currency: "EUR", status: "completed",
		},
		{
			platform: "booking", ref: "BK-" + suffix + "-2", guest: "Marco Rossi",
			checkIn: monthStart.AddDate(0, 0, 5), checkOut: monthStart.AddDate(0, 0, 8),
			nights: 3, gross: 390, price: 390, fee: 58.50, net: 331.50, currency: "EUR", status: "completed",
		},
		{
			platform: "airbnb", ref: "HM-" + suffix + "-3", guest: "Sophie Mueller",
			checkIn: monthStart.AddDate(0, 0, 9), checkOut: monthStart.AddDate(0, 0, 11),
			nights: 2, gross: 310, price: 310, fee: 46.50, net: 263.50, currency: "EUR", status: "completed",
		},
		{
			platform: "booking", ref: "BK-" + suffix + "-4", guest: "James Smith",
			checkIn: monthStart.AddDate(0, 0, 12), checkOut: monthStart.AddDate(0, 0, 17),
			nights: 5, gross: 675, price: 675, fee: 101.25, net: 573.75, currency: "EUR", status: "completed",
		},
		{
			platform: "airbnb", ref: "HM-" + suffix + "-5", guest: "Elena Popescu",
			checkIn: monthStart.AddDate(0, 0, 18), checkOut: monthStart.AddDate(0, 0, 20),
			nights: 2, gross: 280, price: 280, fee: 42, net: 238, currency: "EUR", status: "confirmed",
		},
		{
			platform: "booking", ref: "BK-" + suffix + "-6", guest: "Hans Weber",
			checkIn: monthStart.AddDate(0, 0, 21), checkOut: monthStart.AddDate(0, 0, 25),
			nights: 4, gross: 540, price: 540, fee: 81, net: 459, currency: "EUR", status: "confirmed",
		},
		{
			platform: "airbnb", ref: "HM-" + suffix + "-7", guest: "Laura Chen",
			checkIn: monthStart.AddDate(0, 0, 14), checkOut: monthStart.AddDate(0, 0, 16),
			nights: 2, gross: 290, price: 290, fee: 43.50, net: 246.50, currency: "EUR", status: "cancelled",
		},
	}

	for _, b := range bookings {
		_, err := h.queries.CreateBooking(ctx, sqlc.CreateBookingParams{
			PropertyID:        propertyID,
			Platform:          b.platform,
			PlatformReference: b.ref,
			GuestName:         b.guest,
			CheckIn:           demoDate(b.checkIn),
			CheckOut:          demoDate(b.checkOut),
			Nights:            b.nights,
			GrossAmount:       demoNumeric(b.gross),
			Price:             demoNumeric(b.price),
			PlatformFee:       demoNumeric(b.fee),
			NetAmount:         demoNumeric(b.net),
			Currency:          b.currency,
			Status:            b.status,
		})
		if err != nil {
			return fmt.Errorf("create booking %s: %w", b.ref, err)
		}
	}

	// Electricity bill
	_, err := h.queries.CreateExpense(ctx, sqlc.CreateExpenseParams{
		PropertyID:  propertyID,
		Category:    "electricity",
		Description: "Monthly electricity bill",
		Amount:      demoNumeric(87.50),
		Currency:    "EUR",
		Date:        demoDate(monthStart.AddDate(0, 0, 14)),
		Recurring:   true,
	})
	if err != nil {
		return fmt.Errorf("create expense: %w", err)
	}

	// Income tax rule (10% of profit)
	_, err = h.queries.CreateTaxRule(ctx, sqlc.CreateTaxRuleParams{
		PropertyID: propertyID,
		Name:       "Income Tax (10%)",
		Type:       "percentage",
		Percentage: demoNumeric(10),
		Currency:   "EUR",
	})
	if err != nil {
		return fmt.Errorf("create tax rule: %w", err)
	}

	return nil
}

func (h *AuthHandler) UpgradeGuest(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req authRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if !emailRegex.MatchString(req.Email) {
		respondError(w, http.StatusBadRequest, "invalid email format")
		return
	}
	if len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		respondInternalError(w, "failed to hash password", err)
		return
	}

	user, err := h.queries.UpgradeGuestUser(r.Context(), sqlc.UpgradeGuestUserParams{
		ID:           userID,
		Email:        req.Email,
		PasswordHash: pgtype.Text{String: hash, Valid: true},
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, "upgrade failed — account may not be a guest or email already in use")
		return
	}

	uid, _ := uuid.Parse(uuidBytesToString(user.ID))
	token, err := h.jwtSvc.GenerateToken(uid, user.Email)
	if err != nil {
		respondInternalError(w, "failed to generate token", err)
		return
	}

	respondJSON(w, http.StatusOK, authResponse{
		Token: token,
		User: userInfo{
			ID:        uuidBytesToString(user.ID),
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			Name:      user.Name,
			AvatarURL: user.AvatarUrl,
			UserType:  user.UserType,
		},
	})
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func respondHTML(w http.ResponseWriter, status int, html string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprint(w, html)
}

func errorHTML(msg string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html><body><p>%s</p><script>
  setTimeout(function(){ window.close(); }, 3000);
</script></body></html>`, msg)
}
