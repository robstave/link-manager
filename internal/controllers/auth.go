package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/robstave/link-manager/internal/middleware"
	"github.com/robstave/link-manager/internal/services"
)

type AuthController struct{ service *services.AuthService }

func NewAuthController(service *services.AuthService) *AuthController {
	return &AuthController{service: service}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	token, expiresAt, err := c.service.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{Token: token, ExpiresAt: expiresAt})
}

func (c *AuthController) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"id": claims.UserID, "username": claims.Username, "role": claims.Role})
}
