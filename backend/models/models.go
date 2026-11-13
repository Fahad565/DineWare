package models

import "time"

// Itinerary represents a safari booking itinerary spanning one or more lodges.
type Itinerary struct {
	ID        string    `json:"id" example:"ITIN-2026-001"`
	GroupName string    `json:"group_name" example:"Smith Family Safari"`
	CreatedAt time.Time `json:"created_at" example:"2026-11-13T12:00:00Z"`
	UpdatedAt time.Time `json:"updated_at,omitempty" example:"2026-11-13T12:00:00Z"`
	Stops     []Stop    `json:"stops"`
	Guests    []Guest   `json:"guests"`
}

// Stop represents an individual lodge destination in an itinerary.
type Stop struct {
	ID               int    `json:"id,omitempty" example:"1"`
	ItineraryID      string `json:"itinerary_id,omitempty" example:"ITIN-2026-001"`
	LodgeName        string `json:"lodge_name" example:"Mara Bush Camp"`
	ChefPhone        string `json:"chef_phone" example:"+254711223344"`
	CheckInDate      string `json:"check_in_date" example:"2026-12-01"`
	NotificationSent bool   `json:"notification_sent" example:"false"`
	ChefConfirmed    bool   `json:"chef_confirmed" example:"false"`
}

// Guest holds personal dietary and allergy specifications (protected health info).
type Guest struct {
	ID                  int    `json:"id,omitempty" example:"1"`
	ItineraryID         string `json:"itinerary_id,omitempty" example:"ITIN-2026-001"`
	Name                string `json:"name" example:"Alice Smith"`
	Allergies           string `json:"allergies" example:"Peanut (Severe Anaphylaxis), Shellfish"`
	DietaryRequirements string `json:"dietary_requirements,omitempty" example:"Strict Vegetarian"`
}

// StopInput specifies incoming payload data for an itinerary stop.
type StopInput struct {
	LodgeName   string `json:"lodge_name" example:"Mara Bush Camp"`
	ChefPhone   string `json:"chef_phone" example:"+254711223344"`
	CheckInDate string `json:"check_in_date" example:"2026-12-01"`
}

// GuestInput specifies incoming payload data for a guest's dietary & allergy requirements.
type GuestInput struct {
	Name                string `json:"name" example:"Alice Smith"`
	Allergies           string `json:"allergies" example:"Peanut (Severe Anaphylaxis), Shellfish"`
	DietaryRequirements string `json:"dietary_requirements,omitempty" example:"Strict Vegetarian"`
}

// CreateItineraryRequest represents the payload to register a new multi-lodge safari itinerary.
type CreateItineraryRequest struct {
	ID        string       `json:"id,omitempty" example:"ITIN-2026-001"`
	GroupName string       `json:"group_name" example:"Smith Family Safari"`
	Stops     []StopInput  `json:"stops"`
	Guests    []GuestInput `json:"guests"`
}

// UpdateGuestsRequest represents the payload to modify guest counts or allergy manifests mid-trip.
type UpdateGuestsRequest struct {
	Guests []GuestInput `json:"guests"`
}

// HealthResponse represents system health status.
type HealthResponse struct {
	Status    string `json:"status" example:"OK"`
	Timestamp string `json:"timestamp,omitempty" example:"2026-11-13T19:50:00Z"`
}

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Error   string `json:"error" example:"invalid request body"`
	Details string `json:"details,omitempty"`
}
