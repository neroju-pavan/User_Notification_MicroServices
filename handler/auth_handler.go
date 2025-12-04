package handler

import (
	"encoding/json"
	"net/http"

	"test123/requests"
	"test123/service"
	"test123/utils"
)

type AuthHandler struct {
	AuthService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		AuthService: authService,
	}
}

// POST /auth/passwordreset/generate
// Send reset-token for password reset
func (h *AuthHandler) GenerateResetToken(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Email string `json:"email"`
	}

	var body req
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	status, resp := h.AuthService.GenerateResetToken(r.Context(), body.Email)
	utils.RespondJSON(w, status, resp)
}

// POST /auth/reset-password?token=xyz
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Password string `json:"password"`
	}

	var body req
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "token required"})
		return
	}

	status, resp := h.AuthService.ResetPassword(r.Context(), token, body.Password)
	utils.RespondJSON(w, status, resp)
}

// POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body requests.LoginReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	status, resp := h.AuthService.Login(r.Context(), body.Username, body.Password)
	utils.RespondJSON(w, status, resp)
}

// POST /auth/logout
func (h *AuthHandler) WipeOutSession(w http.ResponseWriter, r *http.Request) {
	var body requests.LoginTokenRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	status, resp := h.AuthService.WipeOutSession(r.Context(), body.AccessToken, body.RefreshToken)
	utils.RespondJSON(w, status, resp)
}

// POST /auth/token/refresh
func (h *AuthHandler) GenerateAccessToken(w http.ResponseWriter, r *http.Request) {
	token, err := utils.ExtractToken(r)
	if err != nil || token == "" {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or missing token"})
		return
	}

	status, resp := h.AuthService.GenerateAccessToken(r.Context(), token)
	utils.RespondJSON(w, status, resp)
}
