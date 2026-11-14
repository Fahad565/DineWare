package integrations_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dineware-backend/integrations"
)

// ─── SMS Privacy Tests ────────────────────────────────────────────────────────

// TestSMSNeverLeaksAllergyKeywords verifies that BuildPrivacySafeSMS never
// embeds any sensitive dietary or health term in the outbound SMS payload.
func TestSMSNeverLeaksAllergyKeywords(t *testing.T) {
	t.Parallel()

	msg, err := integrations.BuildPrivacySafeSMS("ITIN-2026-001", "Smith Family Safari", 4, "2026-12-01")
	if err != nil {
		t.Fatalf("BuildPrivacySafeSMS returned unexpected error: %v", err)
	}

	// Every keyword that must NEVER appear in a plain SMS
	forbidden := []string{
		"allerg", "peanut", "gluten", "shellfish", "lactose", "dairy", "nut",
		"anaphyla", "intoleran", "vegan", "vegetarian", "celiac", "soy", "wheat",
		"egg", "fish", "sesame", "mustard", "sulphite", "sulphur", "lupin",
		"mollusc", "dietary", "meal requirement", "health", "medical",
	}
	lower := strings.ToLower(msg)
	for _, kw := range forbidden {
		if strings.Contains(lower, kw) {
			t.Errorf("SMS payload leaks sensitive keyword %q:\n  Payload: %s", kw, msg)
		}
	}
}

// TestSMSContainsRequiredFields verifies that the SMS payload contains the
// four mandatory non-sensitive tracking elements.
func TestSMSContainsRequiredFields(t *testing.T) {
	t.Parallel()

	itinID := "ITIN-2026-007"
	group := "Okafor Safari Group"
	date := "2026-12-15"
	guests := 6

	msg, err := integrations.BuildPrivacySafeSMS(itinID, group, guests, date)
	if err != nil {
		t.Fatalf("BuildPrivacySafeSMS error: %v", err)
	}

	checks := map[string]string{
		"itinerary reference code": itinID,
		"group name":               group,
		"arrival date":             date,
		"USSD shortcode prompt":    "*384*44#",
	}
	for label, substr := range checks {
		if !strings.Contains(msg, substr) {
			t.Errorf("SMS missing required field (%s=%q):\n  Payload: %s", label, substr, msg)
		}
	}
}

// TestSMSBlocksInjectedAllergyContent verifies the safety belt: if sensitive
// content were somehow injected into an SMS build call, it must be blocked.
func TestSMSBlocksInjectedAllergyContent(t *testing.T) {
	t.Parallel()

	// Use a group name crafted to try to smuggle a sensitive term
	// (the safety check inside BuildPrivacySafeSMS must catch it)
	_, err := integrations.BuildPrivacySafeSMS("ITIN-X", "Peanut-Free Lovers Tour", 3, "2026-12-01")
	if err == nil {
		t.Error("expected BuildPrivacySafeSMS to block a payload containing 'peanut', but it succeeded")
	}
}

// TestSendSMSCallsATEndpoint verifies the HTTP layer sends to Africa's Talking
// with the correct headers and form body, without leaking sensitive data.
func TestSendSMSCallsATEndpoint(t *testing.T) {
	t.Parallel()

	var capturedBody, capturedAPIKey string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAPIKey = r.Header.Get("ApiKey")
		if err := r.ParseForm(); err != nil {
			t.Errorf("server: ParseForm error: %v", err)
		}
		capturedBody = r.FormValue("message")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"SMSMessageData":{"Message":"Sent to 1/1 Total Cost: KES 0.8000"}}`))
	}))
	defer srv.Close()

	client := integrations.NewATClientWithBaseURL("sandbox", "test-api-key", srv.URL)
	err := client.SendSMS(context.Background(), "+254711000001", "ITIN-2026-001", "Smith Safari", 4, "2026-12-01")
	if err != nil {
		t.Fatalf("SendSMS returned error: %v", err)
	}

	if capturedAPIKey != "test-api-key" {
		t.Errorf("expected ApiKey header 'test-api-key', got %q", capturedAPIKey)
	}

	lower := strings.ToLower(capturedBody)
	for _, kw := range []string{"allerg", "peanut", "gluten", "shellfish", "dietary", "health"} {
		if strings.Contains(lower, kw) {
			t.Errorf("SMS body dispatched to AT API contains sensitive keyword %q: %s", kw, capturedBody)
		}
	}
}

// ─── USSD State Machine Tests ─────────────────────────────────────────────────

// stubProvider implements GuestManifestProvider for USSD state machine tests.
type stubProvider struct {
	stop       *integrations.StopRecord
	guests     []integrations.GuestRecord
	confirmErr error
}

func (s *stubProvider) GetStopByItineraryAndPhone(_ context.Context, itinID, phone string) (*integrations.StopRecord, error) {
	if s.stop != nil && s.stop.ItineraryID == itinID && s.stop.ChefPhone == phone {
		return s.stop, nil
	}
	return nil, nil
}

func (s *stubProvider) GetGuestsByItinerary(_ context.Context, _ string) ([]integrations.GuestRecord, error) {
	return s.guests, nil
}

func (s *stubProvider) ConfirmManifest(_ context.Context, _, _ string) error {
	return s.confirmErr
}

func makeStub() *stubProvider {
	return &stubProvider{
		stop: &integrations.StopRecord{
			ItineraryID: "ITIN-2026-001",
			GroupName:   "Okafor Safari",
			LodgeName:   "Mara Bush Camp",
			ChefPhone:   "+254711000001",
			CheckInDate: "2026-12-01",
			GuestCount:  4,
		},
		guests: []integrations.GuestRecord{
			{Name: "Alice Smith", Allergies: "Peanut (Severe Anaphylaxis)", DietaryRequirements: "Vegan"},
			{Name: "Bob Smith", Allergies: "None", DietaryRequirements: ""},
		},
	}
}

// TestUSSDLevel0RootMenu verifies the entry-level welcome prompt.
func TestUSSDLevel0RootMenu(t *testing.T) {
	t.Parallel()
	req := integrations.USSDRequest{
		SessionID: "AT-SID-001", ServiceCode: "*384*44#", PhoneNumber: "+254711000001", Text: "",
	}
	resp := integrations.HandleUSSD(context.Background(), req, makeStub())
	if !strings.HasPrefix(resp, "CON") {
		t.Errorf("expected CON response at level 0, got: %s", resp)
	}
	if !strings.Contains(resp, "Reference Code") {
		t.Errorf("level 0 response should prompt for reference code, got: %s", resp)
	}
}

// TestUSSDLevel1ValidCode verifies that a valid itinerary code + phone returns
// the action menu (CON) and does NOT leak allergy data.
func TestUSSDLevel1ValidCode(t *testing.T) {
	t.Parallel()
	req := integrations.USSDRequest{
		SessionID: "AT-SID-001", ServiceCode: "*384*44#", PhoneNumber: "+254711000001", Text: "ITIN-2026-001",
	}
	resp := integrations.HandleUSSD(context.Background(), req, makeStub())
	if !strings.HasPrefix(resp, "CON") {
		t.Errorf("expected CON response at level 1 for valid code, got: %s", resp)
	}
	// Must offer both menu options
	if !strings.Contains(resp, "1.") || !strings.Contains(resp, "2.") {
		t.Errorf("level 1 response must present both menu options, got: %s", resp)
	}
	// MUST NOT contain allergy data at this level
	lower := strings.ToLower(resp)
	for _, kw := range []string{"peanut", "anaphyla", "vegan"} {
		if strings.Contains(lower, kw) {
			t.Errorf("level 1 CON response leaks sensitive keyword %q: %s", kw, resp)
		}
	}
}

// TestUSSDLevel1InvalidCode verifies that an unknown reference code returns END.
func TestUSSDLevel1InvalidCode(t *testing.T) {
	t.Parallel()
	req := integrations.USSDRequest{
		SessionID: "AT-SID-002", ServiceCode: "*384*44#", PhoneNumber: "+254711000001", Text: "ITIN-WRONG",
	}
	resp := integrations.HandleUSSD(context.Background(), req, makeStub())
	if !strings.HasPrefix(resp, "END") {
		t.Errorf("expected END for invalid code, got: %s", resp)
	}
}

// TestUSSDLevel2Option1ViewManifest verifies that menu option 1 returns the
// full allergy manifest ONLY inside the authenticated USSD session (END).
func TestUSSDLevel2Option1ViewManifest(t *testing.T) {
	t.Parallel()
	req := integrations.USSDRequest{
		SessionID: "AT-SID-001", ServiceCode: "*384*44#", PhoneNumber: "+254711000001",
		Text: "ITIN-2026-001*1",
	}
	resp := integrations.HandleUSSD(context.Background(), req, makeStub())
	if !strings.HasPrefix(resp, "END") {
		t.Errorf("expected END after viewing manifest, got: %s", resp)
	}
	// Allergy data IS allowed inside the authenticated USSD END response
	if !strings.Contains(resp, "Alice Smith") {
		t.Errorf("manifest response missing guest name: %s", resp)
	}
	if !strings.Contains(strings.ToLower(resp), "peanut") {
		t.Errorf("manifest response missing allergy data: %s", resp)
	}
}

// TestUSSDLevel2Option2ConfirmManifest verifies that menu option 2 flips
// chef_confirmed and returns a success END response.
func TestUSSDLevel2Option2ConfirmManifest(t *testing.T) {
	t.Parallel()
	req := integrations.USSDRequest{
		SessionID: "AT-SID-001", ServiceCode: "*384*44#", PhoneNumber: "+254711000001",
		Text: "ITIN-2026-001*2",
	}
	resp := integrations.HandleUSSD(context.Background(), req, makeStub())
	if !strings.HasPrefix(resp, "END") {
		t.Errorf("expected END after confirming manifest, got: %s", resp)
	}
	if !strings.Contains(strings.ToLower(resp), "confirm") {
		t.Errorf("confirmation response should mention confirmation, got: %s", resp)
	}
}

// TestUSSDLevel2InvalidOption verifies that an unrecognised selection returns END.
func TestUSSDLevel2InvalidOption(t *testing.T) {
	t.Parallel()
	req := integrations.USSDRequest{
		SessionID: "AT-SID-001", ServiceCode: "*384*44#", PhoneNumber: "+254711000001",
		Text: "ITIN-2026-001*9",
	}
	resp := integrations.HandleUSSD(context.Background(), req, makeStub())
	if !strings.HasPrefix(resp, "END") {
		t.Errorf("expected END for invalid selection, got: %s", resp)
	}
}

// TestUSSDWrongPhoneNumberDenied verifies that a matching itinerary code but
// wrong chef phone is rejected at level 1.
func TestUSSDWrongPhoneNumberDenied(t *testing.T) {
	t.Parallel()
	req := integrations.USSDRequest{
		SessionID: "AT-SID-003", ServiceCode: "*384*44#", PhoneNumber: "+254799999999", // wrong phone
		Text: "ITIN-2026-001",
	}
	resp := integrations.HandleUSSD(context.Background(), req, makeStub())
	if !strings.HasPrefix(resp, "END") {
		t.Errorf("expected END for unregistered phone, got: %s", resp)
	}
}
