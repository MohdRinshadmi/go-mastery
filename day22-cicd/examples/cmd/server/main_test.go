// Day 22 — unit tests for the server handlers.
// Run with:  go test -race -v ./...
// CI runs:   go test -race -coverprofile=coverage.out -covermode=atomic ./...
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	// Create a test server using the same handler as production.
	h := healthHandler("v1.2.3", "abc1234")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", body["status"])
	}
	if body["version"] != "v1.2.3" {
		t.Errorf("expected version=v1.2.3, got %q", body["version"])
	}
	if body["git_commit"] != "abc1234" {
		t.Errorf("expected git_commit=abc1234, got %q", body["git_commit"])
	}
}

func TestParsePort(t *testing.T) {
	tests := []struct {
		input    string
		fallback int
		want     int
	}{
		{"8080", 3000, 8080},
		{"", 3000, 3000},        // empty → default
		{"99999", 3000, 3000},   // out of range → default
		{"abc", 3000, 3000},     // non-numeric → default
		{"443", 9090, 443},
	}

	for _, tc := range tests {
		got := parsePort(tc.input, tc.fallback)
		if got != tc.want {
			t.Errorf("parsePort(%q, %d) = %d, want %d", tc.input, tc.fallback, got, tc.want)
		}
	}
}

func TestOrderStatusString(t *testing.T) {
	tests := []struct {
		status OrderStatus
		want   string
	}{
		{OrderPending, "pending"},
		{OrderPaid, "paid"},
		{OrderShipped, "shipped"},
		{OrderDelivered, "delivered"},
		{OrderStatus(99), "unknown"},
	}

	for _, tc := range tests {
		if got := tc.status.String(); got != tc.want {
			t.Errorf("OrderStatus(%d).String() = %q, want %q", tc.status, got, tc.want)
		}
	}
}
