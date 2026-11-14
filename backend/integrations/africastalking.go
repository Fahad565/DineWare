// Package integrations provides client wrappers for Africa's Talking SMS and USSD APIs.
//
// STRICT PRIVACY RULE: Outbound SMS messages must NEVER contain sensitive health or
// dietary data (allergy names, dietary conditions, etc.). Only itinerary reference code,
// guest count, arrival date, and a prompt to dial *384*44# may appear in SMS payloads.
// Sensitive manifest data is revealed exclusively inside authenticated USSD sessions.
package integrations

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// smsBaseURL is the Africa's Talking SMS endpoint.
const smsBaseURL = "https://api.africastalking.com/version1/messaging"

// sensitiveKeywords is the list of terms that must NEVER appear in an SMS payload.
// This list is used in tests and in the production send path as a safety belt.
var sensitiveKeywords = []string{
	"allerg", "peanut", "gluten", "shellfish", "lactose", "dairy", "nut",
	"anaphyla", "intoleran", "vegan", "vegetarian", "celiac", "soy", "wheat",
	"egg", "fish", "sesame", "mustard", "sulphite", "sulphur", "lupin",
	"mollusc", "dietary", "meal requirement", "health", "medical",
}

// ATClient is an Africa's Talking API client for SMS dispatch.
type ATClient struct {
	username   string
	apiKey     string
	httpClient *http.Client
	baseURL    string // overridable in tests
}

// NewATClient constructs an ATClient from the given credentials.
func NewATClient(username, apiKey string) *ATClient {
	return &ATClient{
		username:   username,
		apiKey:     apiKey,
		baseURL:    smsBaseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// NewATClientWithBaseURL constructs an ATClient with an overridden base URL.
// Use this in tests to point the client at an httptest.Server.
func NewATClientWithBaseURL(username, apiKey, baseURL string) *ATClient {
	return &ATClient{
		username:   username,
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// BuildPrivacySafeSMS constructs the redacted SMS body for a lodge notification.
//
// The payload contains ONLY:
//   - Property-specific itinerary reference code.
//   - Number of arriving guests.
//   - Scheduled arrival date.
//   - A prompt to retrieve the full manifest via USSD (*384*44#) or the web portal.
//
// No allergy, dietary, or health information is included.
func BuildPrivacySafeSMS(itineraryID, groupName string, guestCount int, arrivalDate string) (string, error) {
	msg := fmt.Sprintf(
		"DineWare Alert: Group '%s' (Ref: %s) — %d guest(s) arriving %s. "+
			"Dial *384*44# on your GSM phone to view the full manifest securely. "+
			"Web portal: https://dineware.app/manifest",
		groupName, itineraryID, guestCount, arrivalDate,
	)

	// Safety belt: enforce zero leakage of sensitive terms
	if err := assertNoSensitiveData(msg); err != nil {
		return "", fmt.Errorf("SMS privacy violation blocked: %w", err)
	}
	return msg, nil
}

// assertNoSensitiveData returns an error if the payload contains any sensitive keyword.
// This is called both in production and in tests.
func assertNoSensitiveData(payload string) error {
	lower := strings.ToLower(payload)
	for _, kw := range sensitiveKeywords {
		if strings.Contains(lower, kw) {
			return fmt.Errorf("sensitive keyword %q detected in SMS payload", kw)
		}
	}
	return nil
}

// SendSMS dispatches a privacy-safe SMS to a chef's phone number.
//
// Parameters:
//   - ctx: request context.
//   - to: recipient phone number in E.164 format (e.g. "+254711223344").
//   - itineraryID: property-specific itinerary reference code.
//   - groupName: safari group name.
//   - guestCount: number of arriving guests.
//   - arrivalDate: formatted arrival date string (YYYY-MM-DD).
func (c *ATClient) SendSMS(ctx context.Context, to, itineraryID, groupName string, guestCount int, arrivalDate string) error {
	msg, err := BuildPrivacySafeSMS(itineraryID, groupName, guestCount, arrivalDate)
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("username", c.username)
	params.Set("to", to)
	params.Set("message", msg)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("failed to build SMS request: %w", err)
	}
	req.Header.Set("ApiKey", c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("SMS dispatch failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("AT SMS API returned HTTP %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// ─── USSD Handler ────────────────────────────────────────────────────────────

// USSDRequest holds the parsed fields from an Africa's Talking USSD webhook POST.
type USSDRequest struct {
	SessionID   string
	ServiceCode string
	PhoneNumber string
	Text        string // cumulative input, levels joined by "*"
}

// GuestManifestProvider is the interface the USSD handler uses to fetch data.
// Implemented by the db.Store layer in production and by stubs in tests.
type GuestManifestProvider interface {
	// GetStopByItineraryAndPhone returns the stop matching both keys, used to
	// verify the chef's identity and locate the manifest.
	GetStopByItineraryAndPhone(ctx context.Context, itineraryID, chefPhone string) (*StopRecord, error)

	// GetGuestsByItinerary returns all guest records for detailed manifest view.
	// This is called only inside authenticated USSD sessions — never in SMS.
	GetGuestsByItinerary(ctx context.Context, itineraryID string) ([]GuestRecord, error)

	// ConfirmManifest sets chef_confirmed = true for the given stop.
	ConfirmManifest(ctx context.Context, itineraryID, chefPhone string) error
}

// StopRecord is a lightweight projection used by the USSD handler.
type StopRecord struct {
	ItineraryID  string
	GroupName    string
	LodgeName    string
	ChefPhone    string
	CheckInDate  string
	GuestCount   int
	ChefConfirmed bool
}

// GuestRecord is a lightweight projection containing sensitive manifest data.
// This struct is ONLY used inside authenticated USSD sessions.
type GuestRecord struct {
	Name                string
	Allergies           string
	DietaryRequirements string
}

// HandleUSSD processes an inbound USSD webhook from Africa's Talking and returns
// the response string prefixed with "CON" (continue session) or "END" (terminate).
//
// Session state machine (cumulative `text` field, levels joined by "*"):
//
//	Level 0 (text == ""):            Root menu — prompt for itinerary reference code.
//	Level 1 (one segment):           Verify chef phone + itinerary code → show action menu.
//	Level 2 (two segments):
//	  Option "1" → View allergy & dietary manifest (sensitive, inside authenticated session).
//	  Option "2" → Confirm manifest receipt (sets chef_confirmed = true).
func HandleUSSD(ctx context.Context, req USSDRequest, provider GuestManifestProvider) string {
	parts := strings.Split(req.Text, "*")

	// Level 0: root entry — chef just dialled the shortcode
	if req.Text == "" {
		return "CON Welcome to DineWare Kitchen Portal.\nEnter your Itinerary Reference Code:"
	}

	itineraryCode := strings.TrimSpace(parts[0])

	// Level 1: chef entered the reference code — authenticate via phone + code
	if len(parts) == 1 {
		stop, err := provider.GetStopByItineraryAndPhone(ctx, itineraryCode, req.PhoneNumber)
		if err != nil || stop == nil {
			return "END Access Denied: Itinerary code not found or your number is not registered for this manifest. Please contact your tour operator."
		}
		return fmt.Sprintf(
			"CON Manifest verified for %s (%s).\n%d guest(s) arriving %s.\n1. View Full Kitchen Manifest\n2. Confirm Manifest Receipt",
			stop.LodgeName, itineraryCode, stop.GuestCount, stop.CheckInDate,
		)
	}

	// Level 2: deep menu action
	if len(parts) == 2 {
		action := strings.TrimSpace(parts[1])

		// Re-authenticate at every deep level
		stop, err := provider.GetStopByItineraryAndPhone(ctx, itineraryCode, req.PhoneNumber)
		if err != nil || stop == nil {
			return "END Session expired or invalid. Please dial again."
		}

		switch action {
		case "1":
			// Return detailed allergy/dietary manifest — ONLY safe inside USSD session
			guests, err := provider.GetGuestsByItinerary(ctx, itineraryCode)
			if err != nil || len(guests) == 0 {
				return "END No guest manifest found. Contact tour operator."
			}
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("END Manifest for %s (%d guests):\n", stop.GroupName, len(guests)))
			for i, g := range guests {
				sb.WriteString(fmt.Sprintf("Guest %d: %s\n  Allergies: %s", i+1, g.Name, g.Allergies))
				if g.DietaryRequirements != "" {
					sb.WriteString(fmt.Sprintf("\n  Diet: %s", g.DietaryRequirements))
				}
				if i < len(guests)-1 {
					sb.WriteString("\n")
				}
			}
			return sb.String()

		case "2":
			// Confirm manifest receipt — sets chef_confirmed = true in DB
			if err := provider.ConfirmManifest(ctx, itineraryCode, req.PhoneNumber); err != nil {
				return "END System error. Please try again or use the web portal."
			}
			return "END Confirmed. Manifest receipt logged for your lodge. Safe travels!"

		default:
			return "END Invalid selection. Please dial *384*44# and choose 1 or 2."
		}
	}

	return "END Invalid input. Please dial *384*44# to start again."
}
