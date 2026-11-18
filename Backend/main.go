package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return errors.New("DATABASE_URL is required")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if len(os.Args) > 1 {
		if len(os.Args) != 2 || os.Args[1] != "migrate" {
			return errors.New("usage: go run . [migrate]")
		}
		goose.SetBaseFS(migrations)
		if err := goose.SetDialect("postgres"); err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return goose.UpContext(ctx, db, "migrations")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = db.PingContext(ctx)
	cancel()
	if err != nil {
		return err
	}

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           routes(db),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("Listening on %s", addr)
	return server.ListenAndServe()
}

func routes(db *sql.DB) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Route("/api", func(api chi.Router) {
		api.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := db.PingContext(ctx); err != nil {
				writeJSON(w, http.StatusServiceUnavailable, "error", "Database tidak tersedia.")
				return
			}
			writeJSON(w, http.StatusOK, "status", "ok")
		})
		api.Post("/contact", contactHandler(db))
		api.Get("/content", readSiteContent(db))
		api.Route("/admin/content", func(admin chi.Router) {
			admin.Use(requireCMSKey(os.Getenv("CMS_ADMIN_KEY")))
			admin.Get("/", readSiteContent(db))
			admin.Put("/", saveSiteContent(db))
		})
	})
	if dist := os.Getenv("STATIC_DIR"); dist != "" {
		router.Handle("/*", http.FileServer(http.Dir(dist)))
	}
	return router
}

type contactMessage struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

func contactHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || contentType != "application/json" {
			writeJSON(w, http.StatusUnsupportedMediaType, "error", "Gunakan format JSON.")
			return
		}

		var contact contactMessage
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&contact); err != nil || decoder.Decode(new(any)) != io.EOF {
			writeJSON(w, http.StatusBadRequest, "error", "Data pesan tidak valid.")
			return
		}
		contact.Name = strings.TrimSpace(contact.Name)
		contact.Email = strings.TrimSpace(contact.Email)
		contact.Message = strings.TrimSpace(contact.Message)
		if problem := contact.validationError(); problem != "" {
			writeJSON(w, http.StatusBadRequest, "error", problem)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		_, err = db.ExecContext(ctx,
			"INSERT INTO contact_messages (name, email, message) VALUES ($1, $2, $3)",
			contact.Name, contact.Email, contact.Message,
		)
		if err != nil {
			// Driver errors can contain submitted data.
			log.Printf("Saving contact message failed (%T)", err)
			writeJSON(w, http.StatusServiceUnavailable, "error", "Pesan belum bisa disimpan.")
			return
		}
		writeJSON(w, http.StatusCreated, "status", "created")
	}
}

func (contact contactMessage) validationError() string {
	if contact.Name == "" || utf8.RuneCountInString(contact.Name) > 100 || strings.ContainsRune(contact.Name, 0) {
		return "Isi nama dengan 1–100 karakter."
	}
	address, err := mail.ParseAddress(contact.Email)
	if err != nil || address.Address != contact.Email || len(contact.Email) > 254 {
		return "Masukkan alamat email yang valid."
	}
	if contact.Message == "" || utf8.RuneCountInString(contact.Message) > 3000 || strings.ContainsRune(contact.Message, 0) {
		return "Isi pesan dengan 1–3000 karakter."
	}
	return ""
}

func writeJSON(w http.ResponseWriter, status int, key, value string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{key: value})
}
