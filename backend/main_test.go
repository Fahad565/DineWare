package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"dineware-backend/db"
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
