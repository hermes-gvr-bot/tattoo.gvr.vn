package handler

import (
	"context"
	"log"
	"net/http"

	"tattoo-consultation/internal/auth"
	"tattoo-consultation/internal/middleware"
	"tattoo-consultation/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GenerateHandler handles the admin-triggered AI generation endpoint.
type GenerateHandler struct {
	Pool      *pgxpool.Pool
	UploadDir string
	Generator *service.Generator
}

// TriggerGenerate starts the AI pipeline for a consultation (admin only).
// POST /api/admin/consultations/{id}/generate
func (h *GenerateHandler) TriggerGenerate(w http.ResponseWriter, r *http.Request) {
	consultationID := chi.URLParam(r, "id")

	// Verify consultation exists and is in a valid state
	var status string
	err := h.Pool.QueryRow(r.Context(),
		`SELECT status FROM consultations WHERE id = $1`, consultationID,
	).Scan(&status)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "consultation not found"})
		return
	}

	if status == "generating" {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "generation already in progress"})
		return
	}

	if status != "pending_payment" && status != "payment_confirmed" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "consultation must be in pending_payment or payment_confirmed status, current: " + status,
		})
		return
	}

	// Check admin role
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*auth.Claims)
	if !ok || claims.Role != "admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin only"})
		return
	}

	// Start generation in background (use context.Background(), NOT r.Context())
	go func() {
		log.Printf("[generate] Starting pipeline for consultation %s", consultationID)
		if err := h.Generator.GenerateConsultation(context.Background(), consultationID); err != nil {
			log.Printf("[generate] Pipeline failed for %s: %v", consultationID, err)
		} else {
			log.Printf("[generate] Pipeline completed for %s", consultationID)
		}
	}()

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":          "generating",
		"consultation_id": consultationID,
	})
}
