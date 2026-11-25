package integrations

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewATClientSelectsSandboxEndpointCaseInsensitively(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		username    string
		wantURL     string
		wantUser    string
	}{
		{
			name:     "username fallback sandbox",
			username: " sandbox ",
			wantURL:  smsSandboxBaseURL,
			wantUser: "sandbox",
		},
		{
			name:     "username fallback case insensitive",
			username: "SaNdBoX",
			wantURL:  smsSandboxBaseURL,
			wantUser: "sandbox",
		},
		{
			name:        "explicit sandbox",
			environment: " SANDBOX ",
			username:    "live-account",
			wantURL:     smsSandboxBaseURL,
			wantUser:    "live-account",
		},
		{
			name:        "explicit live",
			environment: "LIVE",
			username:    "live-account",
			wantURL:     smsBaseURL,
			wantUser:    "live-account",
		},
		{
			name:        "production overrides sandbox username",
			environment: "production",
			username:    "sandbox",
			wantURL:     smsBaseURL,
			wantUser:    "sandbox",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewATClientWithConfig(ATConfig{
				Environment: tt.environment,
				Username:    tt.username,
				APIKey:      "api-key",
			})
			if client.baseURL != tt.wantURL {
				t.Errorf("expected endpoint %q, got %q", tt.wantURL, client.baseURL)
			}
			if client.username != tt.wantUser {
				t.Errorf("expected username %q, got %q", tt.wantUser, client.username)
			}
		})
	}
}

func TestATClientConfiguredSenderAndMessageLinks(t *testing.T) {
	var capturedFrom, capturedMessage string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		capturedFrom = r.FormValue("from")
		capturedMessage = r.FormValue("message")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"SMSMessageData":{"Recipients":[{"statusCode":101,"number":"+254711000001"}]}}`))
	}))
	defer srv.Close()

	client := NewATClientWithConfig(ATConfig{
		Username: "sandbox", APIKey: "test-api-key", SenderID: " DineWare ",
		USSDCode: "*123#", PortalURL: " https://portal.example.test/ ",
	})
	client.baseURL = srv.URL
	if err := client.SendSMS(context.Background(), "+254711000001", "ITIN-2026-001", "Safari Group", 4, "2026-12-01"); err != nil {
		t.Fatalf("send SMS: %v", err)
	}
	if capturedFrom != "DineWare" {
		t.Errorf("expected trimmed sender ID, got %q", capturedFrom)
	}
	if !strings.Contains(capturedMessage, "Dial *123#") {
		t.Errorf("expected configured USSD code in message, got %q", capturedMessage)
	}
	if !strings.Contains(capturedMessage, "https://portal.example.test/manifest") {
		t.Errorf("expected configured portal link in message, got %q", capturedMessage)
	}
}

func TestATClientAcceptsRecipientSuccessStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"SMSMessageData":{"Recipients":[{"number":"+254711000001","status":"Success"}]}}`))
	}))
	defer srv.Close()

	client := NewATClientWithBaseURL("sandbox", "test-api-key", srv.URL)
	if err := client.SendSMS(context.Background(), "+254711000001", "ITIN-2026-001", "Safari Group", 4, "2026-12-01"); err != nil {
		t.Fatalf("expected status=Success to be accepted, got %v", err)
	}
}

func TestBuildPrivacySafeSMSUsesDefaults(t *testing.T) {
	message, err := BuildPrivacySafeSMS("ITIN-1", "Safari", 2, "2026-12-01")
	if err != nil {
		t.Fatalf("build SMS: %v", err)
	}
	if !strings.Contains(message, defaultUSSDCode) || !strings.Contains(message, defaultPortalURL+"/manifest") {
		t.Errorf("default links missing from message: %q", message)
	}
}
