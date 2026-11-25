package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dineware-backend/db"
	"dineware-backend/handlers"
	"dineware-backend/models"
)

func newTestMux() (http.Handler, db.Store) {
	store := db.NewMemoryStore()
	h := handlers.NewHandler(store, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /api/v1/itineraries", h.CreateItinerary)
	mux.HandleFunc("GET /api/v1/itineraries", h.ListItineraries)
	mux.HandleFunc("GET /api/v1/itineraries/{id}", h.GetItinerary)
	mux.HandleFunc("PUT /api/v1/itineraries/{id}", h.UpdateItinerary)
	mux.HandleFunc("DELETE /api/v1/itineraries/{id}", h.DeleteItinerary)
	mux.HandleFunc("PATCH /api/v1/itineraries/{id}/guests", h.UpdateGuests)

	return handlers.CORSMiddleware(mux), store
}

func TestHealthEndpoint(t *testing.T) {
	router, _ := newTestMux()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	var resp models.HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode health response: %v", err)
	}

	if resp.Status != "OK" {
		t.Errorf("expected status 'OK', got %q", resp.Status)
	}
}

func TestItineraryLifecycle(t *testing.T) {
	router, _ := newTestMux()

	// 1. Create a multi-lodge itinerary with guest allergy manifest
	createPayload := models.CreateItineraryRequest{
		ID:        "ITIN-TEST-001",
		GroupName: "Kilimanjaro Expedition Alpha",
		Stops: []models.StopInput{
			{
				LodgeName:   "Amboseli Serena Safari Lodge",
				ChefPhone:   "+254711000111",
				CheckInDate: "2026-12-01",
			},
			{
				LodgeName:   "Ngorongoro Crater Lodge",
				ChefPhone:   "+254722000222",
				CheckInDate: "2026-12-05",
			},
		},
		Guests: []models.GuestInput{
			{
				Name:                "Sarah Connor",
				Allergies:           "Peanuts (Severe Anaphylaxis), Shellfish",
				DietaryRequirements: "Strict Vegan",
			},
			{
				Name:                "John Connor",
				Allergies:           "Gluten Intolerance",
				DietaryRequirements: "Gluten-Free",
			},
		},
	}

	body, err := json.Marshal(createPayload)
	if err != nil {
		t.Fatalf("failed to marshal create payload: %v", err)
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
		t.Fatalf("failed to decode created itinerary: %v", err)
	}

	if created.ID != "ITIN-TEST-001" {
		t.Errorf("expected ID ITIN-TEST-001, got %s", created.ID)
	}
	if len(created.Stops) != 2 {
		t.Fatalf("expected 2 stops, got %d", len(created.Stops))
	}
	if len(created.Guests) != 2 {
		t.Fatalf("expected 2 guests, got %d", len(created.Guests))
	}
	if created.Guests[0].Allergies != "Peanuts (Severe Anaphylaxis), Shellfish" {
		t.Errorf("unexpected allergy recorded: %s", created.Guests[0].Allergies)
	}

	// 2. List itineraries
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/itineraries", nil)
	listRec := httptest.NewRecorder()

	router.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from list, got %d", listRec.Code)
	}

	var list []models.Itinerary
	if err := json.NewDecoder(listRec.Body).Decode(&list); err != nil {
		t.Fatalf("failed to decode list: %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("expected 1 itinerary in list, got %d", len(list))
	}

	// 3. Update guest count or allergies mid-trip via PATCH
	patchPayload := models.UpdateGuestsRequest{
		Guests: []models.GuestInput{
			{
				Name:                "Sarah Connor",
				Allergies:           "Peanuts (Severe Anaphylaxis), Shellfish, Tree Nuts",
				DietaryRequirements: "Strict Vegan / Organic Only",
			},
			{
				Name:                "John Connor",
				Allergies:           "Gluten Intolerance",
				DietaryRequirements: "Gluten-Free",
			},
			{
				Name:                "Kyle Reese (Late Addition)",
				Allergies:           "Dairy / Lactose",
				DietaryRequirements: "Dairy-Free",
			},
		},
	}

	patchBody, _ := json.Marshal(patchPayload)
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/itineraries/ITIN-TEST-001/guests", bytes.NewReader(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchRec := httptest.NewRecorder()

	router.ServeHTTP(patchRec, patchReq)

	if patchRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from patch, got %d: %s", patchRec.Code, patchRec.Body.String())
	}

	var updated models.Itinerary
	if err := json.NewDecoder(patchRec.Body).Decode(&updated); err != nil {
		t.Fatalf("failed to decode updated itinerary: %v", err)
	}

	if len(updated.Guests) != 3 {
		t.Fatalf("expected 3 guests after mid-trip update, got %d", len(updated.Guests))
	}
	if updated.Guests[0].Allergies != "Peanuts (Severe Anaphylaxis), Shellfish, Tree Nuts" {
		t.Errorf("expected updated allergies for guest 0, got %s", updated.Guests[0].Allergies)
	}
	if updated.Guests[2].Name != "Kyle Reese (Late Addition)" {
		t.Errorf("expected added guest 2, got %s", updated.Guests[2].Name)
	}

	// 4. Test 404 for non-existent itinerary PATCH
	notFoundReq := httptest.NewRequest(http.MethodPatch, "/api/v1/itineraries/NON-EXISTENT/guests", bytes.NewReader(patchBody))
	notFoundReq.Header.Set("Content-Type", "application/json")
	notFoundRec := httptest.NewRecorder()

	router.ServeHTTP(notFoundRec, notFoundReq)

	if notFoundRec.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", notFoundRec.Code)
	}
}

func TestUpdateAndDeleteItinerary(t *testing.T) {
	router, _ := newTestMux()
	createPayload := models.CreateItineraryRequest{
		ID:        "ITIN-EDIT-DELETE",
		GroupName: "Original trip",
		Stops: []models.StopInput{{
			LodgeName: "Original Lodge", ChefPhone: "+254700000001",
			CheckInDate: time.Now().Add(5 * 24 * time.Hour).Format("2006-01-02"),
		}},
		Guests: []models.GuestInput{{Name: "Original Guest", Allergies: "None"}},
	}
	createBody, err := json.Marshal(createPayload)
	if err != nil {
		t.Fatalf("marshal create payload: %v", err)
	}
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/itineraries", bytes.NewReader(createBody))
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
	updatePayload := models.UpdateItineraryRequest{
		GroupName: "Updated trip",
		Stops: []models.UpdateStopInput{{
			ID: created.Stops[0].ID, LodgeName: "Updated Lodge",
			ChefPhone: "+254700000002", CheckInDate: created.Stops[0].CheckInDate,
		}},
		Guests: []models.GuestInput{{
			Name: "Updated Guest", Allergies: "Peanuts", DietaryRequirements: "Vegan",
		}},
	}
	updateBody, err := json.Marshal(updatePayload)
	if err != nil {
		t.Fatalf("marshal update payload: %v", err)
	}
	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/itineraries/"+created.ID, bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", updateRec.Code, updateRec.Body.String())
	}

	var updated models.Itinerary
	if err := json.NewDecoder(updateRec.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated itinerary: %v", err)
	}
	if updated.GroupName != "Updated trip" || updated.Stops[0].LodgeName != "Updated Lodge" ||
		updated.Stops[0].ChefPhone != "+254700000002" || updated.Guests[0].Name != "Updated Guest" {
		t.Fatalf("unexpected updated itinerary: %+v", updated)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/itineraries/"+created.ID, nil)
	deleteRec := httptest.NewRecorder()
	router.ServeHTTP(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusNoContent || deleteRec.Body.Len() != 0 {
		t.Fatalf("expected 204 No Content, got %d with body %q", deleteRec.Code, deleteRec.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/itineraries/"+created.ID, nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after deletion, got %d", getRec.Code)
	}
}

func TestValidationErrors(t *testing.T) {
	router, _ := newTestMux()

	cases := []struct {
		name       string
		payload    models.CreateItineraryRequest
		expectedSt int
	}{
		{
			name: "missing group name",
			payload: models.CreateItineraryRequest{
				GroupName: "",
				Stops: []models.StopInput{
					{LodgeName: "Lodge A", ChefPhone: "+254700000000", CheckInDate: "2026-12-01"},
				},
				Guests: []models.GuestInput{
					{Name: "Guest A", Allergies: "None"},
				},
			},
			expectedSt: http.StatusBadRequest,
		},
		{
			name: "missing stops",
			payload: models.CreateItineraryRequest{
				GroupName: "Safari Group",
				Stops:     []models.StopInput{},
				Guests: []models.GuestInput{
					{Name: "Guest A", Allergies: "None"},
				},
			},
			expectedSt: http.StatusBadRequest,
		},
		{
			name: "invalid check-in date",
			payload: models.CreateItineraryRequest{
				GroupName: "Safari Group",
				Stops: []models.StopInput{
					{LodgeName: "Lodge A", ChefPhone: "+254700000000", CheckInDate: "not-a-date"},
				},
				Guests: []models.GuestInput{
					{Name: "Guest A", Allergies: "None"},
				},
			},
			expectedSt: http.StatusBadRequest,
		},
		{
			name: "missing guests",
			payload: models.CreateItineraryRequest{
				GroupName: "Safari Group",
				Stops: []models.StopInput{
					{LodgeName: "Lodge A", ChefPhone: "+254700000000", CheckInDate: "2026-12-01"},
				},
				Guests: []models.GuestInput{},
			},
			expectedSt: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := json.Marshal(tc.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/itineraries", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tc.expectedSt {
				t.Errorf("expected status %d, got %d", tc.expectedSt, rec.Code)
			}
		})
	}
}

func TestCORSPreflight(t *testing.T) {
	router, _ := newTestMux()

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/itineraries", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content for OPTIONS, got %d", rec.Code)
	}

	if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("expected CORS origin '*', got %q", origin)
	}
}
