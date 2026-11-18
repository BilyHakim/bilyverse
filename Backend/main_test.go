package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContactRejectsMalformedRequests(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		status      int
	}{
		{"form body", "text/plain", "", http.StatusUnsupportedMediaType},
		{"empty message", "application/json", `{"name":"Bily","email":"bily@example.com","message":" "}`, http.StatusBadRequest},
		{"invalid email", "application/json", `{"name":"Bily","email":"Bily <bily@example.com>","message":"Halo"}`, http.StatusBadRequest},
		{"extra value", "application/json", `{"name":"Bily","email":"bily@example.com","message":"Halo"} {}`, http.StatusBadRequest},
		{"oversized body", "application/json", strings.Repeat(" ", 33<<10), http.StatusBadRequest},
	}
	router := routes(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/contact", strings.NewReader(tt.body))
			request.Header.Set("Content-Type", tt.contentType)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != tt.status {
				t.Fatalf("got %d, want %d", response.Code, tt.status)
			}
		})
	}
}
