package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Nino-Prog/shortr/internal/handler"
	"github.com/Nino-Prog/shortr/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("cannot connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}

	s := store.New(pool)
	h := handler.New(s)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	// Auth
	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)

	// Redirect (rate-limited)
	r.With(handler.RateLimit).Get("/{code}", h.Redirect)

	// Authenticated API
	r.Group(func(r chi.Router) {
		r.Use(h.RequireAuth)
		r.Post("/api/shorten", h.Shorten)
		r.Get("/api/links", h.ListLinks)
		r.Delete("/api/links/{code}", h.DeleteLink)
		r.Get("/api/analytics/{code}", h.Analytics)
	})

	// Static frontend
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/index.html")
	})
	r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/dashboard.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("shortr listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
