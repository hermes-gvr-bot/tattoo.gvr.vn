package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"tattoo-consultation/internal/auth"
	"tattoo-consultation/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConsultationHandler struct {
	Pool      *pgxpool.Pool
	UploadDir string
}

func (h *ConsultationHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.ClaimsKey).(*auth.Claims)

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
		writeJSON(w, 400, map[string]string{"error": "file too large (max 10MB)"})
		return
	}

	file, header, err := r.FormFile("body_photo")
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "body_photo required"})
		return
	}
	defer file.Close()

	ideaText := r.FormValue("idea_text")
	if ideaText == "" {
		writeJSON(w, 400, map[string]string{"error": "idea_text required"})
		return
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	photoName := uuid.New().String() + ext
	photoPath := filepath.Join(h.UploadDir, photoName)

	dst, err := os.Create(photoPath)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "failed to save file"})
		return
	}
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		writeJSON(w, 500, map[string]string{"error": "failed to write file"})
		return
	}
	dst.Close()

	var consID string
	err = h.Pool.QueryRow(r.Context(),
		`INSERT INTO consultations (user_id, body_photo_path, idea_text, status)
		 VALUES ($1, $2, $3, 'pending_payment') RETURNING id`,
		claims.UserID, photoPath, ideaText,
	).Scan(&consID)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": fmt.Sprintf("db insert: %v", err)})
		return
	}

	writeJSON(w, 201, map[string]interface{}{
		"id":              consID,
		"status":          "pending_payment",
		"body_photo_path": "/uploads/" + photoName,
		"idea_text":       ideaText,
	})
}

func (h *ConsultationHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.ClaimsKey).(*auth.Claims)

	rows, err := h.Pool.Query(r.Context(),
		`SELECT id, body_photo_path, idea_text, status, body_part, skin_tone, created_at
		 FROM consultations WHERE user_id = $1 ORDER BY created_at DESC LIMIT 50`,
		claims.UserID,
	)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "db query failed"})
		return
	}
	defer rows.Close()

	type ConsultationRow struct {
		ID            string  `json:"id"`
		BodyPhotoPath string  `json:"body_photo_path"`
		IdeaText      string  `json:"idea_text"`
		Status        string  `json:"status"`
		BodyPart      *string `json:"body_part"`
		SkinTone      *string `json:"skin_tone"`
		CreatedAt     string  `json:"created_at"`
	}

	var cons []ConsultationRow
	for rows.Next() {
		var c ConsultationRow
		var createdAt interface{}
		if err := rows.Scan(&c.ID, &c.BodyPhotoPath, &c.IdeaText, &c.Status, &c.BodyPart, &c.SkinTone, &createdAt); err != nil {
			continue
		}
		if t, ok := createdAt.(interface{ Format(string) string }); ok {
			c.CreatedAt = t.Format("2006-01-02T15:04:05Z")
		} else {
			c.CreatedAt = fmt.Sprintf("%v", createdAt)
		}
		cons = append(cons, c)
	}
	if cons == nil {
		cons = []ConsultationRow{}
	}

	writeJSON(w, 200, cons)
}

func (h *ConsultationHandler) Get(w http.ResponseWriter, r *http.Request) {
	consID := chi.URLParam(r, "id")

	var c struct {
		ID             string    `json:"id"`
		UserID         string    `json:"user_id"`
		BodyPhotoPath  string    `json:"body_photo_path"`
		IdeaText       string    `json:"idea_text"`
		Status         string    `json:"status"`
		BodyPart       *string   `json:"body_part"`
		SkinTone       *string   `json:"skin_tone"`
		SizeEstimation *string   `json:"size_estimation"`
		VisionAnalysis *string   `json:"vision_analysis"`
		AdminNotes     *string   `json:"admin_notes"`
		CreatedAt      time.Time `json:"created_at"`
	}

	var createdAt time.Time
	err := h.Pool.QueryRow(r.Context(),
		`SELECT id, user_id, body_photo_path, idea_text, status, body_part, skin_tone,
		        size_estimation, vision_analysis, admin_notes, created_at
		 FROM consultations WHERE id = $1`, consID,
	).Scan(&c.ID, &c.UserID, &c.BodyPhotoPath, &c.IdeaText, &c.Status,
		&c.BodyPart, &c.SkinTone, &c.SizeEstimation, &c.VisionAnalysis, &c.AdminNotes, &createdAt)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	c.CreatedAt = createdAt

	// Fetch variants
	rows, err := h.Pool.Query(r.Context(),
		`SELECT id, variant_number, prompt_used, sketch_path, final_path, sketch_status, final_status
		 FROM variants WHERE consultation_id = $1 ORDER BY variant_number`, consID)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "variants query failed"})
		return
	}
	defer rows.Close()

	type Variant struct {
		ID            string  `json:"id"`
		VariantNumber int     `json:"variant_number"`
		PromptUsed    string  `json:"prompt_used"`
		SketchPath    *string `json:"sketch_path"`
		FinalPath     *string `json:"final_path"`
		SketchStatus  string  `json:"sketch_status"`
		FinalStatus   string  `json:"final_status"`
	}

	var variants []Variant
	for rows.Next() {
		var v Variant
		rows.Scan(&v.ID, &v.VariantNumber, &v.PromptUsed, &v.SketchPath, &v.FinalPath, &v.SketchStatus, &v.FinalStatus)
		variants = append(variants, v)
	}
	if variants == nil {
		variants = []Variant{}
	}

	writeJSON(w, 200, map[string]interface{}{
		"consultation": c,
		"variants":     variants,
	})
}
