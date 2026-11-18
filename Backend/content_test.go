package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCMSRoutesRequireKey(t *testing.T) {
	t.Setenv("CMS_ADMIN_KEY", strings.Repeat("a", 32))
	router := routes(nil)
	for _, method := range []string{http.MethodGet, http.MethodPut} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(method, "/api/admin/content/", nil))
		if response.Code != http.StatusUnauthorized {
			t.Errorf("%s returned %d, want 401", method, response.Code)
		}
	}

	request := httptest.NewRequest(http.MethodPut, "/api/admin/content/", strings.NewReader(`{"revision":0,"fields":{"unexpected":"value"}}`))
	request.Header.Set("X-CMS-Key", strings.Repeat("a", 32))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("invalid content returned %d, want 400", response.Code)
	}
}

func TestContentEntries(t *testing.T) {
	published := contentEntry{ID: "one", Title: "First", Date: "2026-11-18", URL: "https://example.com", Published: true}
	draft := contentEntry{ID: "two", Title: "Second"}
	entries := []contentEntry{published, draft}
	if !validEntries(entries) {
		t.Fatal("valid entries were rejected")
	}
	if result := publishedEntries(entries); len(result) != 1 || result[0] != published {
		t.Fatalf("unexpected public entries: %#v", result)
	}
	if validEntries([]contentEntry{published, published}) {
		t.Error("duplicate ID accepted")
	}
	draft.URL = "javascript:alert(1)"
	if validEntries([]contentEntry{draft}) {
		t.Error("unsafe URL accepted")
	}
}
