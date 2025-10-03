package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// -------- API
	r.Route("/api", func(api chi.Router) {
		api.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"ok":true,"time":"` + time.Now().Format(time.RFC3339) + `"}`))
		})
	})

	// -------- Static frontend (built Vite app)
	dist := os.Getenv("WEB_DIST")
	if dist == "" {
		dist = filepath.Join(".", "web", "dist")
	}
	fs := http.Dir(dist)
	fileServer := http.FileServer(fs)

	// Serve assets (e.g. /assets/* from Vite)
	r.Handle("/assets/*", fileServer)

	// SPA fallback: any non-API path → index.html
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(dist, "index.html"))
	})

	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
