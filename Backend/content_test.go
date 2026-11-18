package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContentImage(t *testing.T) {
	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, 12, 8)), nil); err != nil {
		t.Fatal(err)
	}
	encoded := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buffer.Bytes())
	entry := contentEntry{ID: "a & b", Title: "Project", Published: true, Image: encoded}
	if !validEntries([]contentEntry{entry}, true) || validEntries([]contentEntry{entry}, false) {
		t.Fatal("image was accepted outside its allowed collections or rejected in an allowed collection")
	}
	if decodeContentImage("data:image/jpeg;base64,AAAA") != nil {
		t.Fatal("invalid JPEG accepted")
	}
	entries := publishedEntries([]contentEntry{entry})
	publicImageURLs(entries, "projects", 2)
	if entries[0].Image != "/api/content/image?collection=projects&id=a+%26+b&v=2" {
		t.Fatalf("unexpected public image URL %q", entries[0].Image)
	}
}

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
	if !validEntries(entries, false) {
		t.Fatal("valid entries were rejected")
	}
	if result := publishedEntries(entries); len(result) != 1 || result[0] != published {
		t.Fatalf("unexpected public entries: %#v", result)
	}
	if validEntries([]contentEntry{published, published}, false) {
		t.Error("duplicate ID accepted")
	}
	draft.URL = "javascript:alert(1)"
	if validEntries([]contentEntry{draft}, false) {
		t.Error("unsafe URL accepted")
	}
}
