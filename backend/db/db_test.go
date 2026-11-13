package db

import (
	"context"
	"testing"

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
