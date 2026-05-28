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
	"tattoo-consultation/internal/service"
	"tattoo-consultation/internal/storage"

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

	// Init S3/MinIO storage
	s3Client, err := storage.NewS3Client(cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket, cfg.S3PublicURL, cfg.S3UseSSL)
	if err != nil {
		log.Printf("S3 storage init failed (falling back to local disk): %v", err)
	}

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5201", "http://100.86.223.10:5201", "http://tmds-server-local.tail3840e.ts.net:5201", "https://tattoo.gvr.vn"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Auth routes (public)
	authHandler := &handler.AuthHandler{Pool: pool, Cfg: cfg}
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)

	// Protect routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthRequired(cfg.JWTSecret))
		r.Get("/api/auth/me", authHandler.Me)

		// Create LoRA service for body preview generation
		loraService := &service.LoraService{
			Pool:         pool,
			UploadDir:    cfg.UploadDir,
			ReplicateKey: os.Getenv("REPLICATE_API_KEY"),
			S3:           s3Client,
		}

		// Create generator (used by both consultation create + admin trigger)
		generator := &service.Generator{
			Pool:          pool,
			UploadDir:     cfg.UploadDir,
			OpenRouterKey: os.Getenv("OPENROUTER_API_KEY"),
			TogetherKey:   os.Getenv("TOGETHER_API_KEY"),
			ReplicateKey:  os.Getenv("REPLICATE_API_KEY"),
			S3:            s3Client,
			Lora:          loraService,
		}

		// Consultations (auto-generate on create)
		consHandler := &handler.ConsultationHandler{Pool: pool, UploadDir: cfg.UploadDir, S3: s3Client, Generator: generator, Lora: loraService}
		r.Post("/api/consultations", consHandler.Create)
		r.Get("/api/consultations", consHandler.List)
		r.Get("/api/consultations/{id}", consHandler.Get)
		r.Post("/api/consultations/{id}/photos", consHandler.UploadPhotos)
		r.Get("/api/consultations/{id}/lora-status", consHandler.LoraStatus)

		// Admin-only routes
		genHandler := &handler.GenerateHandler{Pool: pool, UploadDir: cfg.UploadDir, Generator: generator}
		r.Get("/api/admin/consultations", consHandler.AdminList)
		r.Get("/api/admin/stats", consHandler.AdminStats)
		r.Post("/api/admin/consultations/{id}/generate", genHandler.TriggerGenerate)
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
