package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"dineware-backend/models"
)

func TestMigrationFilesEmbedded(t *testing.T) {
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		t.Fatalf("failed to read embedded migrations dir: %v", err)
	}

	if len(entries) == 0 {
		t.Fatalf("expected at least one embedded migration file, got 0")
	}

	foundInit := false
	for _, entry := range entries {
		if entry.Name() == "001_init.sql" {
			foundInit = true
			break
		}
	}

	if !foundInit {
		t.Errorf("001_init.sql not found in embedded migrations")
	}

	content, err := migrationFS.ReadFile("migrations/001_init.sql")
	if err != nil {
		t.Fatalf("failed to read 001_init.sql: %v", err)
	}

	if len(content) == 0 {
		t.Errorf("001_init.sql is empty")
	}
}

func TestMemoryStoreCRUD(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	// 1. Create itinerary
	req := models.CreateItineraryRequest{
		GroupName: "Safari Travelers",
		Stops: []models.StopInput{
			{LodgeName: "Serengeti Serena", ChefPhone: "+254700000001", CheckInDate: "2026-11-20"},
		},
		Guests: []models.GuestInput{
			{Name: "Alice", Allergies: "Peanuts", DietaryRequirements: "Halal"},
		},
	}

	created, err := store.CreateItinerary(ctx, req)
	if err != nil {
		t.Fatalf("failed to create itinerary: %v", err)
	}

	if created.ID == "" {
		t.Errorf("expected auto-generated ID, got empty")
	}

	// 2. Get itinerary
	got, err := store.GetItinerary(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get itinerary: %v", err)
	}
	if got.GroupName != "Safari Travelers" {
		t.Errorf("expected group name 'Safari Travelers', got %s", got.GroupName)
	}

	// 3. Update guests
	newGuests := []models.GuestInput{
		{Name: "Alice", Allergies: "Peanuts (Severe Anaphylaxis)", DietaryRequirements: "Halal"},
		{Name: "Bob", Allergies: "None", DietaryRequirements: "Vegetarian"},
	}
	updated, err := store.UpdateGuests(ctx, created.ID, newGuests)
	if err != nil {
		t.Fatalf("failed to update guests: %v", err)
	}
	if len(updated.Guests) != 2 {
		t.Fatalf("expected 2 guests, got %d", len(updated.Guests))
	}

	// 4. List itineraries
	list, err := store.ListItineraries(ctx)
	if err != nil {
		t.Fatalf("failed to list itineraries: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 itinerary in list, got %d", len(list))
	}
}

func TestMemoryStoreUpdateItineraryDispatchResetAndDelete(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	checkIn := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	created, err := store.CreateItinerary(ctx, models.CreateItineraryRequest{
		GroupName: "Original trip",
		Stops: []models.StopInput{
			{LodgeName: "First Lodge", ChefPhone: "+254700000001", CheckInDate: checkIn},
			{LodgeName: "Removed Lodge", ChefPhone: "+254700000002", CheckInDate: checkIn},
		},
		Guests: []models.GuestInput{{Name: "Guest", Allergies: "None"}},
	})
	if err != nil {
		t.Fatalf("create itinerary: %v", err)
	}
	if err := store.MarkNotificationSent(ctx, created.Stops[0].ID); err != nil {
		t.Fatalf("mark stop notified: %v", err)
	}
	if err := store.ConfirmManifest(ctx, created.ID, created.Stops[0].ChefPhone); err != nil {
		t.Fatalf("confirm first stop: %v", err)
	}

	updated, dispatch, err := store.UpdateItinerary(ctx, created.ID, models.UpdateItineraryRequest{
		GroupName: "Updated trip",
		Stops: []models.UpdateStopInput{
			{
				ID: created.Stops[0].ID, LodgeName: "First Lodge Updated",
				ChefPhone: "+254700000003", CheckInDate: checkIn,
			},
			{
				LodgeName: "New Lodge", ChefPhone: "+254700000004",
				CheckInDate: time.Now().Add(5 * 24 * time.Hour).Format("2006-01-02"),
			},
		},
		Guests: []models.GuestInput{{Name: "Updated Guest", Allergies: "Peanuts", DietaryRequirements: "Vegan"}},
	})
	if err != nil {
		t.Fatalf("update itinerary: %v", err)
	}
	if !dispatch {
		t.Fatal("expected changed due stop to trigger dispatch")
	}
	if updated.GroupName != "Updated trip" || len(updated.Stops) != 2 || len(updated.Guests) != 1 {
		t.Fatalf("unexpected updated itinerary: %+v", updated)
	}
	if updated.Stops[0].NotificationSent || updated.Stops[0].ChefConfirmed {
		t.Errorf("expected changed stop notifications and confirmation to reset: %+v", updated.Stops[0])
	}
	if updated.Stops[0].ChefPhone != "+254700000003" || updated.Stops[0].LodgeName != "First Lodge Updated" {
		t.Errorf("stop details were not updated: %+v", updated.Stops[0])
	}

	updated, dispatch, err = store.UpdateItinerary(ctx, created.ID, models.UpdateItineraryRequest{
		GroupName: "Renamed trip",
		Stops: []models.UpdateStopInput{
			{
				ID: updated.Stops[0].ID, LodgeName: updated.Stops[0].LodgeName,
				ChefPhone: updated.Stops[0].ChefPhone, CheckInDate: updated.Stops[0].CheckInDate,
			},
			{
				ID: updated.Stops[1].ID, LodgeName: updated.Stops[1].LodgeName,
				ChefPhone: updated.Stops[1].ChefPhone, CheckInDate: updated.Stops[1].CheckInDate,
			},
		},
		Guests: []models.GuestInput{{Name: "Updated Guest", Allergies: "Peanuts"}},
	})
	if err != nil {
		t.Fatalf("update group name: %v", err)
	}
	if dispatch {
		t.Error("group-name-only update should not trigger dispatch")
	}

	if err := store.DeleteItinerary(ctx, created.ID); err != nil {
		t.Fatalf("delete itinerary: %v", err)
	}
	if _, err := store.GetItinerary(ctx, created.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected deleted itinerary to be missing, got %v", err)
	}
	itineraries, err := store.ListItineraries(ctx)
	if err != nil || len(itineraries) != 0 {
		t.Errorf("expected no itineraries after deletion, itineraries=%v err=%v", itineraries, err)
	}
}

func TestMemoryStoreUSSDAndDispatcherQueries(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	checkIn := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	itinerary, err := store.CreateItinerary(ctx, models.CreateItineraryRequest{
		GroupName: "Safari Group",
		Stops: []models.StopInput{{
			LodgeName: "Mara Lodge", ChefPhone: "+254700000001", CheckInDate: checkIn,
		}},
		Guests: []models.GuestInput{{Name: "Guest", Allergies: "None"}},
	})
	if err != nil {
		t.Fatalf("create itinerary: %v", err)
	}

	stop, err := store.GetStopByItineraryAndPhone(ctx, itinerary.ID, "+254700000001")
	if err != nil {
		t.Fatalf("get USSD stop: %v", err)
	}
	if stop.GuestCount != 1 || stop.LodgeName != "Mara Lodge" {
		t.Fatalf("unexpected USSD stop: %+v", stop)
	}
	guests, err := store.GetGuestsByItinerary(ctx, itinerary.ID)
	if err != nil || len(guests) != 1 || guests[0].Name != "Guest" {
		t.Fatalf("unexpected USSD guests: guests=%+v err=%v", guests, err)
	}
	if err := store.ConfirmManifest(ctx, itinerary.ID, "+254700000001"); err != nil {
		t.Fatalf("confirm manifest: %v", err)
	}
	stop, err = store.GetStopByItineraryAndPhone(ctx, itinerary.ID, "+254700000001")
	if err != nil || !stop.ChefConfirmed {
		t.Fatalf("expected manifest confirmation, stop=%+v err=%v", stop, err)
	}

	pending, err := store.QueryPendingStops(ctx, 72*time.Hour)
	if err != nil || len(pending) != 1 || pending[0].ItineraryID != itinerary.ID {
		t.Fatalf("unexpected pending stops: stops=%+v err=%v", pending, err)
	}
	if err := store.MarkNotificationSent(ctx, pending[0].StopID); err != nil {
		t.Fatalf("mark notification sent: %v", err)
	}
	pending, err = store.QueryPendingStops(ctx, 72*time.Hour)
	if err != nil || len(pending) != 0 {
		t.Fatalf("expected no pending stops after marking sent: stops=%+v err=%v", pending, err)
	}
}
