package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"tattoo-consultation/internal/config"
	"tattoo-consultation/internal/db"
	"tattoo-consultation/internal/handler"
	"tattoo-consultation/internal/middleware"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer pool.Close()
	log.Println("Database connected")

	if err := db.RunMigrations(ctx, pool); err != nil {
		log.Printf("Migration warning (non-fatal): %v", err)
	} else {
		log.Println("Migrations applied")
	}

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5201", "https://tattoo.gvr.vn"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Auth routes (public)
	authHandler := &handler.AuthHandler{Pool: pool, Cfg: cfg}
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthRequired(cfg.JWTSecret))
		r.Get("/api/auth/me", authHandler.Me)

		// Consultations
		consHandler := &handler.ConsultationHandler{Pool: pool, UploadDir: cfg.UploadDir}
		r.Post("/api/consultations", consHandler.Create)
		r.Get("/api/consultations", consHandler.List)
		r.Get("/api/consultations/{id}", consHandler.Get)
	})

	// Serve uploaded files
	uploadsDir := http.Dir(cfg.UploadDir)
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(uploadsDir)))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"tattoo-consultation","version":"0.1.0"}`))
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}
	log.Println("Server stopped")
}
