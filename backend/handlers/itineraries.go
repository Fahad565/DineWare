package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"dineware-backend/db"
	"dineware-backend/models"
)

// CreateItinerary creates a new multi-lodge safari itinerary with guest allergy manifests.
// @Summary      Create a new safari itinerary
// @Description  Allows tour operators to register groups, multi-lodge stops, and protected guest allergy manifests.
// @Tags         itineraries
// @Accept       json
// @Produce      json
// @Param        itinerary  body      models.CreateItineraryRequest  true  "Itinerary registration payload"
// @Success      201        {object}  models.Itinerary
// @Failure      400        {object}  models.ErrorResponse
// @Failure      500        {object}  models.ErrorResponse
// @Router       /api/v1/itineraries [post]
func (h *Handler) CreateItinerary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}

	var req models.CreateItineraryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if strings.TrimSpace(req.GroupName) == "" {
		WriteError(w, http.StatusBadRequest, "group_name is required", "")
		return
	}

	if len(req.Stops) == 0 {
		WriteError(w, http.StatusBadRequest, "at least one lodge stop is required", "")
		return
	}

	for i, s := range req.Stops {
		if strings.TrimSpace(s.LodgeName) == "" {
			WriteError(w, http.StatusBadRequest, "lodge_name is required for each stop", "")
			return
		}
		if strings.TrimSpace(s.ChefPhone) == "" {
			WriteError(w, http.StatusBadRequest, "chef_phone is required for each stop", "")
			return
		}
		if _, err := time.Parse("2006-01-02", s.CheckInDate); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid check_in_date format (expected YYYY-MM-DD)", "stop index: "+string(rune('0'+i)))
			return
		}
	}

	if len(req.Guests) == 0 {
		WriteError(w, http.StatusBadRequest, "at least one guest is required in the manifest", "")
		return
	}

	for _, g := range req.Guests {
		if strings.TrimSpace(g.Name) == "" {
			WriteError(w, http.StatusBadRequest, "name is required for each guest", "")
			return
		}
	}

	itinerary, err := h.store.CreateItinerary(r.Context(), req)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to create itinerary", err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, itinerary)
}

// ListItineraries lists all active safari itineraries with their stops and guest allergy manifests.
// @Summary      List all safari itineraries
// @Description  Retrieves multi-lodge itineraries including guest allergy and dietary manifests.
// @Tags         itineraries
// @Produce      json
// @Success      200  {array}   models.Itinerary
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/itineraries [get]
func (h *Handler) ListItineraries(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}

	list, err := h.store.ListItineraries(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to fetch itineraries", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, list)
}

// GetItinerary retrieves a single safari itinerary by its ID.
// @Summary      Get safari itinerary by ID
// @Description  Retrieves a specific itinerary with all stops and guest dietary manifests.
// @Tags         itineraries
// @Produce      json
// @Param        id   path      string  true  "Itinerary ID"
// @Success      200  {object}  models.Itinerary
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/itineraries/{id} [get]
func (h *Handler) GetItinerary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}

	id := ExtractID(r, "/api/v1/itineraries/", "")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "missing itinerary ID", "")
		return
	}

	it, err := h.store.GetItinerary(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "itinerary not found", "")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to retrieve itinerary", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, it)
}

// UpdateGuests updates guest counts or allergy manifests for an active safari itinerary mid-trip.
// @Summary      Update guest count or allergies mid-trip
// @Description  Dynamically updates the guest allergy manifest and count for an active multi-lodge itinerary.
// @Tags         itineraries
// @Accept       json
// @Produce      json
// @Param        id       path      string                      true  "Itinerary ID"
// @Param        payload  body      models.UpdateGuestsRequest  true  "Updated guests payload"
// @Success      200      {object}  models.Itinerary
// @Failure      400      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/itineraries/{id}/guests [patch]
func (h *Handler) UpdateGuests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}

	id := ExtractID(r, "/api/v1/itineraries/", "/guests")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "missing itinerary ID", "")
		return
	}

	var req models.UpdateGuestsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if len(req.Guests) == 0 {
		WriteError(w, http.StatusBadRequest, "at least one guest is required in the manifest", "")
		return
	}

	for _, g := range req.Guests {
		if strings.TrimSpace(g.Name) == "" {
			WriteError(w, http.StatusBadRequest, "name is required for each guest", "")
			return
		}
	}

	updated, err := h.store.UpdateGuests(r.Context(), id, req.Guests)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "itinerary not found", "")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to update guests", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, updated)
}
