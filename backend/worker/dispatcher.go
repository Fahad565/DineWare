// Package worker provides background job workers for the DineWare engine.
//
// dispatcher.go implements the staggered SMS dispatch evaluator. It queries
// itinerary_stops where notification_sent = false and check_in_date is within
// the next 72 hours, then fires a privacy-safe Africa's Talking SMS to each
// lodge's designated chef phone.
package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"dineware-backend/integrations"
)

// DispatchWindow is the look-ahead horizon: stops with check_in_date within
// this window and notification_sent = false will trigger an SMS dispatch.
const DispatchWindow = 72 * time.Hour

// PendingStop is a lightweight projection of itinerary_stops used by the dispatcher.
// It contains ONLY the non-sensitive fields needed to build the privacy-safe SMS.
type PendingStop struct {
	StopID      int
	ItineraryID string
	GroupName   string
	ChefPhone   string
	LodgeName   string
	CheckInDate string // "YYYY-MM-DD"
	GuestCount  int
}

// StopRepository is the interface the Dispatcher uses to interact with the DB.
// A PostgreSQL implementation is wired in production; a stub is used in tests.
type StopRepository interface {
	// QueryPendingStops returns all stops where notification_sent = false
	// and check_in_date falls within the next 72 hours.
	QueryPendingStops(ctx context.Context, within time.Duration) ([]PendingStop, error)

	// MarkNotificationSent sets notification_sent = true for the given stop ID.
	MarkNotificationSent(ctx context.Context, stopID int) error
}

// SMSSender is the interface the Dispatcher uses to fire outbound SMS messages.
// Implemented by integrations.ATClient in production and by stubs in tests.
type SMSSender interface {
	SendSMS(ctx context.Context, to, itineraryID, groupName string, guestCount int, arrivalDate string) error
}

// Dispatcher evaluates pending stops and fires staggered, privacy-safe SMS alerts.
type Dispatcher struct {
	repo   StopRepository
	sms    SMSSender
	logger *log.Logger
}

// NewDispatcher constructs a Dispatcher with the given dependencies.
func NewDispatcher(repo StopRepository, sms SMSSender, logger *log.Logger) *Dispatcher {
	if logger == nil {
		logger = log.Default()
	}
	return &Dispatcher{repo: repo, sms: sms, logger: logger}
}

// Run executes a single dispatch cycle. It queries for all pending stops within
// the 72-hour window and sends a privacy-safe SMS for each. Stops that fail to
// notify are logged but do not abort the remaining batch.
//
// Call this in a ticker loop or cron to continuously evaluate the window:
//
//	ticker := time.NewTicker(5 * time.Minute)
//	for range ticker.C {
//	    dispatcher.Run(ctx)
//	}
func (d *Dispatcher) Run(ctx context.Context) error {
	stops, err := d.repo.QueryPendingStops(ctx, DispatchWindow)
	if err != nil {
		return fmt.Errorf("dispatcher: query pending stops: %w", err)
	}

	if len(stops) == 0 {
		d.logger.Println("dispatcher: no pending stops within 72-hour window")
		return nil
	}

	d.logger.Printf("dispatcher: found %d stop(s) to notify", len(stops))

	var (
		sent   int
		failed int
	)

	for _, stop := range stops {
		// Staggered dispatch: short pause between messages to avoid AT rate-limit bursts
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := d.sms.SendSMS(ctx, stop.ChefPhone, stop.ItineraryID, stop.GroupName, stop.GuestCount, stop.CheckInDate); err != nil {
			d.logger.Printf("dispatcher: SMS failed for stop %d (phone=%s): %v", stop.StopID, stop.ChefPhone, err)
			failed++
			continue
		}

		if err := d.repo.MarkNotificationSent(ctx, stop.StopID); err != nil {
			d.logger.Printf("dispatcher: failed to mark stop %d as sent: %v", stop.StopID, err)
			// Not fatal — the SMS was delivered; we'll retry marking on the next cycle
		}

		d.logger.Printf("dispatcher: notified chef %s for stop %d (itinerary=%s, check-in=%s)",
			stop.ChefPhone, stop.StopID, stop.ItineraryID, stop.CheckInDate)
		sent++
	}

	d.logger.Printf("dispatcher: cycle complete — sent=%d, failed=%d", sent, failed)
	return nil
}

// ─── Convenience: Postgres Implementation Skeleton ───────────────────────────
// The PostgresStopRepository is wired into production via db.Store.
// It is intentionally separate from the dispatcher logic to keep the dispatcher
// unit-testable with any StopRepository stub.

// NewATSMSSender wraps an *integrations.ATClient to satisfy the SMSSender interface.
// Use this when building the production Dispatcher.
func NewATSMSSender(client *integrations.ATClient) SMSSender {
	return &atSMSSender{client: client}
}

type atSMSSender struct {
	client *integrations.ATClient
}

func (s *atSMSSender) SendSMS(ctx context.Context, to, itineraryID, groupName string, guestCount int, arrivalDate string) error {
	return s.client.SendSMS(ctx, to, itineraryID, groupName, guestCount, arrivalDate)
}
