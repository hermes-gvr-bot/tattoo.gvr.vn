package handler

import (
	"encoding/json"
	"net/http"

	"tattoo-consultation/internal/auth"
	"tattoo-consultation/internal/config"
	"tattoo-consultation/internal/middleware"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	Pool *pgxpool.Pool
	Cfg  *config.Config
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	if req.Email == "" || req.Password == "" || req.Name == "" {
		writeJSON(w, 400, map[string]string{"error": "email, password, name required"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "internal error"})
		return
	}

	var userID string
	err = h.Pool.QueryRow(r.Context(),
		`INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3) RETURNING id`,
		req.Email, string(hash), req.Name,
	).Scan(&userID)
	if err != nil {
		writeJSON(w, 409, map[string]string{"error": "email already exists"})
		return
	}

	uid, _ := uuid.Parse(userID)
	token, _ := auth.GenerateToken(uid, req.Email, "client", h.Cfg.JWTSecret)

	writeJSON(w, 201, map[string]interface{}{
		"token":   token,
		"user_id": userID,
		"email":   req.Email,
		"role":    "client",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}

	var id, email, hash, role string
	err := h.Pool.QueryRow(r.Context(),
		`SELECT id, email, password_hash, role FROM users WHERE email = $1`, req.Email,
	).Scan(&id, &email, &hash, &role)
	if err != nil {
		writeJSON(w, 401, map[string]string{"error": "invalid credentials"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		writeJSON(w, 401, map[string]string{"error": "invalid credentials"})
		return
	}

	uid, _ := uuid.Parse(id)
	token, _ := auth.GenerateToken(uid, email, role, h.Cfg.JWTSecret)

	writeJSON(w, 200, map[string]interface{}{
		"token":   token,
		"user_id": id,
		"email":   email,
		"role":    role,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.ClaimsKey).(*auth.Claims)
	writeJSON(w, 200, map[string]interface{}{
		"user_id": claims.UserID,
		"email":   claims.Email,
		"role":    claims.Role,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
