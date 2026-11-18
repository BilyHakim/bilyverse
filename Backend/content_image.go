package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"image/jpeg"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

const contentImageLimit = 96 * 1024
const contentGalleryLimit = 8

var contentImageID = regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)

type contentImage struct {
	ID  string `json:"id,omitempty"`
	Src string `json:"src"`
	Alt string `json:"alt,omitempty"`
}

func validContentImages(entry contentEntry) bool {
	if len(entry.Images) > contentGalleryLimit || (entry.Image != "" && len(entry.Images) > 0) {
		return false
	}
	if entry.Image != "" && decodeContentImage(entry.Image) == nil {
		return false
	}
	seen := make(map[string]bool, len(entry.Images))
	for _, image := range entry.Images {
		if strings.ContainsRune(image.Alt, 0) || utf8.RuneCountInString(image.Alt) > 180 || decodeContentImage(image.Src) == nil {
			return false
		}
		if image.ID != "" {
			if !contentImageID.MatchString(image.ID) || seen[image.ID] {
				return false
			}
			seen[image.ID] = true
		}
	}
	return true
}

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
	for index := range entries {
		entry := &entries[index]
		query := url.Values{"collection": {collection}, "id": {entry.ID}, "v": {strconv.Itoa(revision)}}
		if entry.Image != "" {
			entry.Image = "/api/content/image?" + query.Encode()
		}
		for position := range entry.Images {
			query.Del("asset")
			query.Del("index")
			if imageID := entry.Images[position].ID; imageID != "" {
				query.Set("asset", imageID)
			} else {
				query.Set("index", strconv.Itoa(position))
			}
			entry.Images[position].Src = "/api/content/image?" + query.Encode()
		}
	}
}

func readContentImage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		collection, id := query.Get("collection"), query.Get("id")
		if (collection != "projects" && collection != "certificates" && collection != "posts") || id == "" || len(id) > 256 {
			http.NotFound(w, r)
			return
		}
		asset := query.Get("asset")
		if query.Has("asset") && !contentImageID.MatchString(asset) {
			http.NotFound(w, r)
			return
		}
		index := -1
		if query.Has("index") {
			var err error
			index, err = strconv.Atoi(query.Get("index"))
			if err != nil || index < 0 || index >= contentGalleryLimit {
				http.NotFound(w, r)
				return
			}
		}
		var encoded string
		err := db.QueryRowContext(r.Context(), `SELECT CASE
				WHEN $4::text <> '' THEN (SELECT image->>'src' FROM jsonb_array_elements(entry->'images') AS gallery(image) WHERE image->>'id' = $4)
				WHEN $3::int = -1 THEN COALESCE(entry->>'image', entry->'images'->0->>'src')
				ELSE entry->'images'->$3::int->>'src' END
			FROM site_content, jsonb_array_elements(collections -> $1::text) AS entries(entry)
			WHERE id AND entry->>'id' = $2 AND entry->'published' = 'true'::jsonb`, collection, id, index, asset).Scan(&encoded)
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
