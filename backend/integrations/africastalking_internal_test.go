package integrations

import "testing"

func TestNewATClientSelectsSandboxEndpointCaseInsensitively(t *testing.T) {
	tests := []struct {
		username string
		wantURL  string
		wantUser string
	}{
		{
			username: " sandbox ",
			wantURL:  smsSandboxBaseURL,
			wantUser: "sandbox",
		},
		{
			username: "SaNdBoX",
			wantURL:  smsSandboxBaseURL,
			wantUser: "sandbox",
		},
		{
			username: "production-user",
			wantURL:  smsBaseURL,
			wantUser: "production-user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.username, func(t *testing.T) {
			client := NewATClient(tt.username, "api-key")
			if client.baseURL != tt.wantURL {
				t.Errorf("expected endpoint %q, got %q", tt.wantURL, client.baseURL)
			}
			if client.username != tt.wantUser {
				t.Errorf("expected username %q, got %q", tt.wantUser, client.username)
			}
		})
	}
}
