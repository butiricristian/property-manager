package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/cristianbutiri/property-manager-v2/internal/database/sqlc"
)

type BackgroundSync struct {
	queries  *sqlc.Queries
	syncSvc  *SyncService
	interval time.Duration
	stopCh   chan struct{}
}

func NewBackgroundSync(q *sqlc.Queries, syncSvc *SyncService, interval time.Duration) *BackgroundSync {
	return &BackgroundSync{
		queries:  q,
		syncSvc:  syncSvc,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (bs *BackgroundSync) Start() {
	log.Printf("background sync started (interval: %s)", bs.interval)
	go func() {
		ticker := time.NewTicker(bs.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				bs.runOnce()
			case <-bs.stopCh:
				log.Println("background sync stopped")
				return
			}
		}
	}()
}

func (bs *BackgroundSync) Stop() {
	close(bs.stopCh)
}

func (bs *BackgroundSync) runOnce() {
	ctx := context.Background()
	integrations, err := bs.queries.ListEnabledIntegrations(ctx)
	if err != nil {
		log.Printf("background sync: failed to list integrations: %v", err)
		return
	}

	log.Printf("background sync: processing %d enabled integrations", len(integrations))

	for _, integration := range integrations {
		result := bs.syncSvc.SyncIntegration(ctx, integration)

		now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
		var syncError pgtype.Text
		if result.Errors > 0 {
			syncError = pgtype.Text{String: result.Message, Valid: true}
		}

		if err := bs.queries.UpdateIntegrationSyncStatus(ctx, sqlc.UpdateIntegrationSyncStatusParams{
			ID:            integration.ID,
			LastSyncedAt:  now,
			LastSyncError: syncError,
		}); err != nil {
			log.Printf("background sync: failed to update sync status for %s: %v",
				uuidToString(integration.ID), err)
		}
	}
}

func uuidToString(u pgtype.UUID) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", u.Bytes[0:4], u.Bytes[4:6], u.Bytes[6:8], u.Bytes[8:10], u.Bytes[10:16])
}
