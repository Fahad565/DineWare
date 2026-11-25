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
	calls      int
	recipients []string
}

func (s *recordingSMSSender) SendSMS(_ context.Context, to, _, _ string, _ int, _ string) error {
	s.calls++
	s.recipients = append(s.recipients, to)
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

func TestUpdateItineraryImmediatelyDispatchesChangedDueStop(t *testing.T) {
	store := db.NewMemoryStore()
	sms := &recordingSMSSender{}
	dispatcher := worker.NewDispatcher(stopRepository{store: store}, sms, nil)
	router := SetupRouter(store, dispatcher)
	payload := models.CreateItineraryRequest{
		ID:        "ITIN-UPDATE-DISPATCH",
		GroupName: "Safari Group",
		Stops: []models.StopInput{{
			LodgeName: "Mara Lodge", ChefPhone: "+254700000001",
			CheckInDate: time.Now().Add(5 * 24 * time.Hour).Format("2006-01-02"),
		}},
		Guests: []models.GuestInput{{Name: "Guest", Allergies: "None"}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal create request: %v", err)
	}
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/itineraries", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", createRec.Code, createRec.Body.String())
	}
	var created models.Itinerary
	if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
		t.Fatalf("decode created itinerary: %v", err)
	}
	if sms.calls != 0 {
		t.Fatalf("expected no SMS for stop outside 72 hours, got %d", sms.calls)
	}

	updatePayload := models.UpdateItineraryRequest{
		GroupName: "Safari Group",
		Stops: []models.UpdateStopInput{{
			ID: created.Stops[0].ID, LodgeName: "Mara Lodge",
			ChefPhone:   "+254700000009",
			CheckInDate: time.Now().Add(24 * time.Hour).Format("2006-01-02"),
		}},
		Guests: []models.GuestInput{{Name: "Guest", Allergies: "None"}},
	}
	body, err = json.Marshal(updatePayload)
	if err != nil {
		t.Fatalf("marshal update request: %v", err)
	}
	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/itineraries/"+created.ID, bytes.NewReader(body))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", updateRec.Code, updateRec.Body.String())
	}
	if sms.calls != 1 || len(sms.recipients) != 1 || sms.recipients[0] != "+254700000009" {
		t.Fatalf("expected updated stop to receive one SMS, calls=%d recipients=%v", sms.calls, sms.recipients)
	}
	var updated models.Itinerary
	if err := json.NewDecoder(updateRec.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated itinerary: %v", err)
	}
	if !updated.Stops[0].NotificationSent {
		t.Error("expected updated stop to be marked notification_sent")
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
		"serviceCode": {"*384*95853#"},
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
