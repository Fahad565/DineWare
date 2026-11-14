package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"dineware-backend/db"
	"dineware-backend/models"
)

func TestSetupRouterEndpoints(t *testing.T) {
	store := db.NewMemoryStore()
	router := SetupRouter(store)

	// Test /health
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from /health, got %d", rec.Code)
	}

	// Test /swagger redirect
	reqSwagger := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	recSwagger := httptest.NewRecorder()
	router.ServeHTTP(recSwagger, reqSwagger)

	if recSwagger.Code != http.StatusMovedPermanently {
		t.Errorf("expected 301 redirect from /swagger, got %d", recSwagger.Code)
	}

	// Test /swagger/index.html
	reqSwaggerUI := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	recSwaggerUI := httptest.NewRecorder()
	router.ServeHTTP(recSwaggerUI, reqSwaggerUI)

	if recSwaggerUI.Code != http.StatusOK {
		t.Fatalf("expected 200 from /swagger/index.html, got %d", recSwaggerUI.Code)
	}
}

func TestUSSDWebhookRoute(t *testing.T) {
	store := db.NewMemoryStore()
	itinerary, err := store.CreateItinerary(context.Background(), models.CreateItineraryRequest{
		GroupName: "Safari Group",
		Stops: []models.StopInput{{
			LodgeName: "Mara Lodge", ChefPhone: "+254700000001",
			CheckInDate: "2026-11-20",
		}},
		Guests: []models.GuestInput{{Name: "Guest", Allergies: "None"}},
	})
	if err != nil {
		t.Fatalf("create itinerary: %v", err)
	}

	form := url.Values{
		"sessionId":   {"session-1"},
		"serviceCode": {"*384*44#"},
		"phoneNumber": {"+254700000001"},
		"text":        {itinerary.ID},
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ussd", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	SetupRouter(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from USSD webhook, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/plain") {
		t.Fatalf("expected text/plain response, got %q", got)
	}
	if !strings.HasPrefix(rec.Body.String(), "CON ") {
		t.Fatalf("expected CON response, got %q", rec.Body.String())
	}
}
