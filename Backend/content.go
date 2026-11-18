package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

type siteContent struct {
	Revision    int                 `json:"revision"`
	Fields      map[string]string   `json:"fields"`
	Collections *contentCollections `json:"collections,omitempty"`
}

type contentEntry struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Summary   string         `json:"summary"`
	Body      string         `json:"body"`
	Date      string         `json:"date"`
	URL       string         `json:"url"`
	Published bool           `json:"published"`
	Image     string         `json:"image,omitempty"`
	ImageAlt  string         `json:"imageAlt,omitempty"`
	Images    []contentImage `json:"images,omitempty"`
}

type contentCollections struct {
	Projects     []contentEntry `json:"projects"`
	Certificates []contentEntry `json:"certificates"`
	Posts        []contentEntry `json:"posts"`
}

var contentLimits = map[string]int{
	"interests": 100, "role": 36, "company": 45,
	"about": 6000, "projectTitle": 120, "projectSummary": 2000, "projectDetails": 6000,
	"certificates": 6000, "blog": 12000, "contactIntro": 1000,
}

func readSiteContent(db *sql.DB, includeDrafts bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var content siteContent
		var fields, collections []byte
		err := db.QueryRowContext(r.Context(), "SELECT revision, fields, collections FROM site_content WHERE id").Scan(&content.Revision, &fields, &collections)
		if err != nil || json.Unmarshal(fields, &content.Fields) != nil ||
			(collections != nil && json.Unmarshal(collections, &content.Collections) != nil) {
			writeJSON(w, http.StatusServiceUnavailable, "error", "Konten belum dapat dimuat.")
			return
		}
		if !includeDrafts && content.Collections != nil {
			content.Collections.Projects = publishedEntries(content.Collections.Projects)
			content.Collections.Certificates = publishedEntries(content.Collections.Certificates)
			content.Collections.Posts = publishedEntries(content.Collections.Posts)
			publicImageURLs(content.Collections.Projects, "projects", content.Revision)
			publicImageURLs(content.Collections.Certificates, "certificates", content.Revision)
			publicImageURLs(content.Collections.Posts, "posts", content.Revision)
			for _, field := range []string{"projectTitle", "projectSummary", "projectDetails", "certificates", "blog"} {
				delete(content.Fields, field)
			}
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		json.NewEncoder(w).Encode(content)
	}
}

func publishedEntries(entries []contentEntry) []contentEntry {
	published := make([]contentEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Published {
			published = append(published, entry)
		}
	}
	return published
}

func requireCMSKey(key string) func(http.Handler) http.Handler {
	expected := sha256.Sum256([]byte(key))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(key) < 32 {
				writeJSON(w, http.StatusServiceUnavailable, "error", "Dashboard belum dikonfigurasi.")
				return
			}
			provided := sha256.Sum256([]byte(r.Header.Get("X-CMS-Key")))
			if subtle.ConstantTimeCompare(provided[:], expected[:]) != 1 {
				writeJSON(w, http.StatusUnauthorized, "error", "Kunci admin tidak valid.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func saveSiteContent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var content siteContent
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<20))
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&content)
		if err == nil {
			err = decoder.Decode(new(any))
			if err == nil {
				writeJSON(w, http.StatusBadRequest, "error", "Kirim satu dokumen JSON untuk setiap penyimpanan.")
				return
			}
		} else if err == io.EOF {
			writeJSON(w, http.StatusBadRequest, "error", "Konten kosong.")
			return
		}
		var sizeError *http.MaxBytesError
		if errors.As(err, &sizeError) {
			writeJSON(w, http.StatusRequestEntityTooLarge, "error", "Konten melebihi batas penyimpanan 32 MiB. Kurangi jumlah gambar lalu coba lagi.")
			return
		}
		if err != io.EOF || !validSiteContent(content) {
			writeJSON(w, http.StatusBadRequest, "error", "Konten tidak valid atau melebihi batas karakter.")
			return
		}
		fields, err := json.Marshal(content.Fields)
		collections, collectionErr := json.Marshal(content.Collections)
		if err != nil || collectionErr != nil {
			writeJSON(w, http.StatusInternalServerError, "error", "Konten gagal disimpan.")
			return
		}
		result, err := db.ExecContext(r.Context(), `UPDATE site_content
			SET fields = $1::jsonb,
			    collections = COALESCE(NULLIF($2::jsonb, 'null'::jsonb), collections),
			    revision = revision + 1 WHERE id AND revision = $3`, string(fields), string(collections), content.Revision)
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, "error", "Konten gagal disimpan.")
			return
		}
		count, err := result.RowsAffected()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, "error", "Penyimpanan gagal dikonfirmasi.")
			return
		}
		if count == 0 {
			writeJSON(w, http.StatusConflict, "error", "Konten sudah berubah. Salin perubahanmu lalu buka ulang editor.")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func validSiteContent(content siteContent) bool {
	if content.Revision < 0 || len(content.Fields) != len(contentLimits) {
		return false
	}
	for field, value := range content.Fields {
		limit, exists := contentLimits[field]
		if !exists || strings.ContainsRune(value, 0) || utf8.RuneCountInString(value) > limit {
			return false
		}
	}
	if content.Collections == nil {
		return true
	}
	return validEntries(content.Collections.Projects) &&
		validEntries(content.Collections.Certificates) && validEntries(content.Collections.Posts)
}

func validEntries(entries []contentEntry) bool {
	if entries == nil || len(entries) > 50 {
		return false
	}
	seen := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if strings.TrimSpace(entry.ID) == "" || strings.TrimSpace(entry.Title) == "" || seen[entry.ID] {
			return false
		}
		seen[entry.ID] = true
		for _, field := range []struct {
			value string
			limit int
		}{{entry.ID, 64}, {entry.Title, 120}, {entry.Summary, 2000}, {entry.Body, 12000}, {entry.Date, 10}, {entry.URL, 2048}, {entry.ImageAlt, 180}} {
			if strings.ContainsRune(field.value, 0) || utf8.RuneCountInString(field.value) > field.limit {
				return false
			}
		}
		if !validContentImages(entry) {
			return false
		}
		if entry.Date != "" {
			if _, err := time.Parse("2006-01-02", entry.Date); err != nil {
				return false
			}
		}
		if entry.URL != "" {
			link, err := url.Parse(entry.URL)
			if err != nil || (link.Scheme != "https" && link.Scheme != "http") || link.Hostname() == "" || link.User != nil {
				return false
			}
		}
	}
	return true
}
