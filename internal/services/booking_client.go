package services

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	defaultBookingBaseURL = "https://secure-supply-xml.booking.com"
	tokenExchangeURL      = "https://connectivity-authentication.booking.com/token-based-authentication/exchange"
	tokenRefreshBuffer    = 5 * time.Minute
)

type BookingClient struct {
	httpClient *http.Client
	baseURL    string

	mu          sync.Mutex
	token       string
	tokenExpiry time.Time
	clientID    string
	clientSecret string
}

func NewBookingClient() *BookingClient {
	return &BookingClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    defaultBookingBaseURL,
	}
}

// BookingReservation represents a parsed reservation from the B.XML API.
type BookingReservation struct {
	ID          string
	GuestName   string
	CheckIn     time.Time
	CheckOut    time.Time
	TotalAmount float64
	Commission  float64
	Currency    string
	Status      string
}

// bxmlRequest is the XML request body for the B.XML API.
type bxmlRequest struct {
	XMLName xml.Name `xml:"request"`
	HotelID string   `xml:"hotel_id"`
}

// tokenExchangeRequest is the JSON body for the token exchange endpoint.
type tokenExchangeRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// tokenExchangeResponse is the JSON response from the token exchange endpoint.
type tokenExchangeResponse struct {
	JWT  string `json:"jwt"`
	RUID string `json:"ruid"`
}

// getToken returns a cached JWT or exchanges credentials for a new one.
func (c *BookingClient) getToken(ctx context.Context, clientID, clientSecret string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Return cached token if still valid and credentials match
	if c.token != "" && c.clientID == clientID && c.clientSecret == clientSecret &&
		time.Now().Before(c.tokenExpiry.Add(-tokenRefreshBuffer)) {
		return c.token, nil
	}

	log.Printf("Booking.com: exchanging credentials for new JWT token")

	body, err := json.Marshal(tokenExchangeRequest{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})
	if err != nil {
		return "", fmt.Errorf("marshaling token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenExchangeURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("creating token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling token exchange: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token exchange failed (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var tokenResp tokenExchangeResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return "", fmt.Errorf("parsing token response: %w", err)
	}

	if tokenResp.JWT == "" {
		return "", fmt.Errorf("token exchange returned empty JWT")
	}

	c.token = tokenResp.JWT
	c.tokenExpiry = time.Now().Add(1 * time.Hour)
	c.clientID = clientID
	c.clientSecret = clientSecret

	log.Printf("Booking.com: obtained new JWT token (expires %s)", c.tokenExpiry.Format(time.RFC3339))
	return c.token, nil
}

// FetchReservations calls the B.XML reservations endpoint for a given hotel.
func (c *BookingClient) FetchReservations(ctx context.Context, clientID, clientSecret, hotelID string) ([]BookingReservation, error) {
	token, err := c.getToken(ctx, clientID, clientSecret)
	if err != nil {
		return nil, fmt.Errorf("authenticating with Booking.com: %w", err)
	}

	xmlReq := bxmlRequest{
		HotelID: hotelID,
	}
	reqBody, err := xml.Marshal(xmlReq)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/hotels/xml/reservations",
		bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling B.XML API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Try to parse fault from XML response body for a cleaner error
		var faultResp bxmlResponse
		if xml.Unmarshal(body, &faultResp) == nil && faultResp.Fault != nil {
			return nil, fmt.Errorf("B.XML API error: %s", faultResp.Fault.String)
		}
		return nil, fmt.Errorf("B.XML API returned HTTP %d", resp.StatusCode)
	}

	return parseBookingXMLResponse(body)
}

type bxmlFault struct {
	Code   string `xml:"code,attr"`
	String string `xml:"string,attr"`
}

// XML response structures for B.XML reservations endpoint.
type bxmlResponse struct {
	XMLName      xml.Name          `xml:"reservations"`
	Fault        *bxmlFault        `xml:"fault"`
	Reservations []bxmlReservation `xml:"reservation"`
}

type bxmlReservation struct {
	ID        string       `xml:"id,attr"`
	Status    string       `xml:"status"`
	Guest     bxmlGuest    `xml:"customer"`
	Room      []bxmlRoom   `xml:"room"`
	TotalPrice string      `xml:"totalprice"`
	Commission string      `xml:"commission_amount"`
	Currency   string      `xml:"currencycode"`
}

type bxmlGuest struct {
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
}

type bxmlRoom struct {
	Arrival   string `xml:"arrival_date"`
	Departure string `xml:"departure_date"`
}

func parseBookingXMLResponse(data []byte) ([]BookingReservation, error) {
	var resp bxmlResponse
	if err := xml.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing B.XML response: %w", err)
	}

	if resp.Fault != nil {
		return nil, fmt.Errorf("B.XML API error %s: %s", resp.Fault.Code, resp.Fault.String)
	}

	var reservations []BookingReservation
	for _, r := range resp.Reservations {
		var checkIn, checkOut time.Time
		if len(r.Room) > 0 {
			var err error
			checkIn, err = time.Parse("2006-01-02", r.Room[0].Arrival)
			if err != nil {
				continue
			}
			checkOut, err = time.Parse("2006-01-02", r.Room[0].Departure)
			if err != nil {
				continue
			}
		} else {
			continue
		}

		total := parseAmount(r.TotalPrice)
		commission := parseAmount(r.Commission)

		status := "confirmed"
		switch r.Status {
		case "cancelled", "no_show":
			status = "cancelled"
		}

		guestName := r.Guest.FirstName
		if r.Guest.LastName != "" {
			if guestName != "" {
				guestName += " "
			}
			guestName += r.Guest.LastName
		}

		currency := r.Currency
		if currency == "" {
			currency = "EUR"
		}

		reservations = append(reservations, BookingReservation{
			ID:          r.ID,
			GuestName:   guestName,
			CheckIn:     checkIn,
			CheckOut:    checkOut,
			TotalAmount: total,
			Commission:  commission,
			Currency:    currency,
			Status:      status,
		})
	}

	return reservations, nil
}
