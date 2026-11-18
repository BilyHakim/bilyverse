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
