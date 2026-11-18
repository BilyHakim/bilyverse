package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"
)

type siteContent struct {
	Revision int               `json:"revision"`
	Fields   map[string]string `json:"fields"`
}

var contentLimits = map[string]int{
	"interests": 100, "role": 36, "company": 45,
	"about": 6000, "projectTitle": 120, "projectSummary": 2000, "projectDetails": 6000,
	"certificates": 6000, "blog": 12000, "contactIntro": 1000,
}

func readSiteContent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var content siteContent
		var fields []byte
		err := db.QueryRowContext(r.Context(), "SELECT revision, fields FROM site_content WHERE id").Scan(&content.Revision, &fields)
		if err != nil || json.Unmarshal(fields, &content.Fields) != nil {
			writeJSON(w, http.StatusServiceUnavailable, "error", "Konten belum dapat dimuat.")
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		json.NewEncoder(w).Encode(content)
	}
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
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256<<10))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&content) != nil || decoder.Decode(new(any)) != io.EOF || !validSiteContent(content) {
			writeJSON(w, http.StatusBadRequest, "error", "Konten tidak valid atau melebihi batas karakter.")
			return
		}
		fields, err := json.Marshal(content.Fields)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, "error", "Konten gagal disimpan.")
			return
		}
		result, err := db.ExecContext(r.Context(), `UPDATE site_content
			SET fields = $1::jsonb, revision = revision + 1 WHERE id AND revision = $2`, string(fields), content.Revision)
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
	return true
}
