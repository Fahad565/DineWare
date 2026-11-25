package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"dineware-backend/db"
	"dineware-backend/models"
	"dineware-backend/worker"
)

type recordingSMSSender struct {
	calls int
}

func (s *recordingSMSSender) SendSMS(context.Context, string, string, string, int, string) error {
	s.calls++
	return nil
}

func TestSetupRouterEndpoints(t *testing.T) {
	store := db.NewMemoryStore()
	router := SetupRouter(store, nil)

	// Test /health
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from /health, got %d", rec.Code)
	}

	headReq := httptest.NewRequest(http.MethodHead, "/health", nil)
	headRec := httptest.NewRecorder()
	router.ServeHTTP(headRec, headReq)
	if headRec.Code != http.StatusOK {
		t.Fatalf("expected 200 from HEAD /health, got %d", headRec.Code)
	}
	if headRec.Body.Len() != 0 {
		t.Fatalf("expected an empty body from HEAD /health, got %q", headRec.Body.String())
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

func TestCreateItineraryImmediatelyDispatchesDueStops(t *testing.T) {
	store := db.NewMemoryStore()
	sms := &recordingSMSSender{}
	dispatcher := worker.NewDispatcher(stopRepository{store: store}, sms, nil)
	router := SetupRouter(store, dispatcher)

	payload := models.CreateItineraryRequest{
		ID:        "ITIN-DISPATCH-001",
		GroupName: "Safari Group",
		Stops: []models.StopInput{
			{
				LodgeName: "Mara Lodge", ChefPhone: "+254700000001",
				CheckInDate: time.Now().Add(24 * time.Hour).Format("2006-01-02"),
			},
			{
				LodgeName: "Serengeti Lodge", ChefPhone: "+254700000002",
				CheckInDate: time.Now().Add(5 * 24 * time.Hour).Format("2006-01-02"),
			},
		},
		Guests: []models.GuestInput{{Name: "Guest", Allergies: "Peanuts"}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/itineraries", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}
	if sms.calls != 1 {
		t.Fatalf("expected one SMS for the stop within 72 hours, got %d", sms.calls)
	}
	created, err := store.GetItinerary(context.Background(), payload.ID)
	if err != nil {
		t.Fatalf("load created itinerary: %v", err)
	}
	if !created.Stops[0].NotificationSent {
		t.Error("expected due stop to be marked notification_sent")
	}
	if created.Stops[1].NotificationSent {
		t.Error("expected stop outside 72 hours to remain unnotified")
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

	SetupRouter(store, nil).ServeHTTP(rec, req)

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
