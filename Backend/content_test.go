package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"slices"
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
	if !validEntries([]contentEntry{entry}) {
		t.Fatal("legacy image was rejected")
	}
	if decodeContentImage("data:image/jpeg;base64,AAAA") != nil {
		t.Fatal("invalid JPEG accepted")
	}
	entries := publishedEntries([]contentEntry{entry})
	publicImageURLs(entries, "projects", 2)
	if entries[0].Image != "/api/content/image?collection=projects&id=a+%26+b&v=2" {
		t.Fatalf("unexpected public image URL %q", entries[0].Image)
	}

	entry.Image = ""
	entry.Images = []contentImage{{Src: encoded, Alt: "Front"}, {Src: encoded, Alt: "Back"}}
	if !validEntries([]contentEntry{entry}) {
		t.Fatal("valid gallery was rejected")
	}
	entries = publishedEntries([]contentEntry{entry})
	publicImageURLs(entries, "posts", 3)
	want := []contentImage{
		{Src: "/api/content/image?collection=posts&id=a+%26+b&index=0&v=3", Alt: "Front"},
		{Src: "/api/content/image?collection=posts&id=a+%26+b&index=1&v=3", Alt: "Back"},
	}
	if !slices.Equal(entries[0].Images, want) {
		t.Fatalf("unexpected public gallery: %#v", entries[0].Images)
	}
	entry.Images = slices.Repeat([]contentImage{{Src: encoded}}, contentGalleryLimit+1)
	if validEntries([]contentEntry{entry}) {
		t.Error("oversized gallery accepted")
	}
	entry.Images = []contentImage{{Src: encoded, Alt: strings.Repeat("é", 181)}}
	if validEntries([]contentEntry{entry}) {
		t.Error("oversized alt text accepted")
	}
	entry.Images = []contentImage{{Src: "https://example.com/image.jpg"}}
	if validEntries([]contentEntry{entry}) {
		t.Error("external image source accepted")
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
	if !validEntries(entries) {
		t.Fatal("valid entries were rejected")
	}
	if result := publishedEntries(entries); len(result) != 1 || result[0].ID != published.ID {
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
