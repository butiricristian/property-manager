package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/cristianbutiri/property-manager-v2/internal/database/sqlc"
	"github.com/cristianbutiri/property-manager-v2/internal/models"
)

type SyncService struct {
	queries       *sqlc.Queries
	bookingClient *BookingClient
}

func NewSyncService(q *sqlc.Queries) *SyncService {
	return &SyncService{
		queries:       q,
		bookingClient: NewBookingClient(),
	}
}

func (s *SyncService) SyncIntegration(ctx context.Context, integration sqlc.Integration) models.SyncResult {
	switch integration.Platform {
	case "booking":
		return s.syncBookingCom(ctx, integration)
	case "airbnb":
		return s.syncAirbnb(ctx, integration)
	default:
		return models.SyncResult{Message: "Unknown platform: " + integration.Platform}
	}
}

func (s *SyncService) syncBookingCom(ctx context.Context, integration sqlc.Integration) models.SyncResult {
	log.Printf("Booking.com sync triggered for integration %s", uuidToString(integration.ID))

	// Parse credentials
	var creds map[string]string
	if err := json.Unmarshal(integration.Credentials, &creds); err != nil {
		return models.SyncResult{
			Errors:  1,
			Message: "Failed to parse integration credentials",
		}
	}

	clientID := creds["client_id"]
	clientSecret := creds["client_secret"]
	if clientID == "" || clientSecret == "" {
		return models.SyncResult{
			Errors:  1,
			Message: "Missing client_id or client_secret in integration credentials",
		}
	}

	// Get property mappings for this integration
	mappings, err := s.queries.ListMappingsByIntegration(ctx, integration.ID)
	if err != nil {
		return models.SyncResult{
			Errors:  1,
			Message: fmt.Sprintf("Failed to load property mappings: %v", err),
		}
	}

	if len(mappings) == 0 {
		return models.SyncResult{
			Message: "No properties linked to this integration. Link a property with its Hotel ID first.",
		}
	}

	result := models.SyncResult{}

	for _, mapping := range mappings {
		reservations, err := s.bookingClient.FetchReservations(ctx, clientID, clientSecret, mapping.ExternalID)
		if err != nil {
			log.Printf("Booking.com sync error for hotel %s: %v", mapping.ExternalID, err)
			result.Errors++
			continue
		}

		for _, res := range reservations {
			nights := int32(math.Round(res.CheckOut.Sub(res.CheckIn).Hours() / 24))
			net := res.TotalAmount - res.Commission

			_, err := s.queries.UpsertBooking(ctx, sqlc.UpsertBookingParams{
				PropertyID:        mapping.PropertyID,
				Platform:          "booking",
				PlatformReference: res.ID,
				GuestName:         res.GuestName,
				CheckIn:           pgtype.Date{Time: res.CheckIn, Valid: true},
				CheckOut:          pgtype.Date{Time: res.CheckOut, Valid: true},
				Nights:            nights,
				GrossAmount:       decimalFromFloat(res.TotalAmount),
				PlatformFee:       decimalFromFloat(res.Commission),
				NetAmount:         decimalFromFloat(net),
				Currency:          res.Currency,
				Status:            res.Status,
				Price:             decimalFromFloat(res.TotalAmount),
			})
			if err != nil {
				log.Printf("Booking.com upsert error for reservation %s: %v", res.ID, err)
				result.Errors++
				continue
			}
			result.Imported++
		}
	}

	result.Message = fmt.Sprintf("Booking.com sync completed: %d imported, %d errors across %d properties",
		result.Imported, result.Errors, len(mappings))
	return result
}

func (s *SyncService) syncAirbnb(ctx context.Context, integration sqlc.Integration) models.SyncResult {
	log.Printf("Airbnb sync triggered for integration %s (stub)", uuidToString(integration.ID))
	return models.SyncResult{
		Message: "Airbnb API integration not yet implemented. Please use CSV import.",
	}
}
