package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"image/jpeg"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const contentImageLimit = 96 * 1024

func decodeContentImage(encoded string) []byte {
	value, ok := strings.CutPrefix(encoded, "data:image/jpeg;base64,")
	if !ok || len(value) > base64.StdEncoding.EncodedLen(contentImageLimit) {
		return nil
	}
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil || len(data) > contentImageLimit {
		return nil
	}
	config, err := jpeg.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width < 1 || config.Height < 1 || config.Width > 1200 || config.Height > 1200 {
		return nil
	}
	return data
}

func publicImageURLs(entries []contentEntry, collection string, revision int) {
	for index, entry := range entries {
		if entry.Image != "" {
			query := url.Values{"collection": {collection}, "id": {entry.ID}, "v": {strconv.Itoa(revision)}}
			entries[index].Image = "/api/content/image?" + query.Encode()
		}
	}
}

func readContentImage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		collection, id := query.Get("collection"), query.Get("id")
		if (collection != "projects" && collection != "certificates") || id == "" || len(id) > 256 {
			http.NotFound(w, r)
			return
		}
		var encoded string
		err := db.QueryRowContext(r.Context(), `SELECT entry->>'image'
			FROM site_content, jsonb_array_elements(collections -> $1::text) AS entries(entry)
			WHERE id AND entry->>'id' = $2 AND entry->'published' = 'true'::jsonb AND entry ? 'image'`, collection, id).Scan(&encoded)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		data := decodeContentImage(encoded)
		if data == nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		w.Write(data)
	}
}
