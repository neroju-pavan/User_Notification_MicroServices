package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"test123/errors"
	"test123/logger"
	"test123/models"
	"test123/service"
	"test123/utils"

	"github.com/go-chi/chi/v5"
)

type ProfileHandler struct {
	ProfileService *service.ProfileService
}

func NewProfileHandler(ps *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{ProfileService: ps}
}

// Create or update profile for a user
func (h *ProfileHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(userIdStr)
	if err != nil || id <= 0 {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": errors.ErrInvalidField.Error()})
		return
	}

	var p models.UserProfile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		logger.Warn("CreateProfile", "invalid json", map[string]interface{}{"error": err.Error()})
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": errors.ErrInvalidJSON.Error()})
		return
	}

	p.UserID = id

	if err := p.Validate(); err != nil {
		utils.RespondJSON(w, utils.HttpStatusFromError(err), map[string]string{"error": err.Error()})
		return
	}

	// Try to see if profile exists
	_, gerr := h.ProfileService.GetProfileByUserID(r.Context(), id)
	if gerr != nil {
		// create
		if err := h.ProfileService.CreateProfile(r.Context(), p); err != nil {
			logger.Error("CreateProfile", "service failed", map[string]interface{}{"error": err.Error()})
			utils.RespondJSON(w, utils.HttpStatusFromError(err), map[string]string{"error": err.Error()})
			return
		}
		utils.RespondJSON(w, http.StatusCreated, map[string]string{"message": "profile created"})
		return
	}

	// update
	if err := h.ProfileService.UpdateProfile(r.Context(), p); err != nil {
		logger.Error("UpdateProfile", "service failed", map[string]interface{}{"error": err.Error()})
		utils.RespondJSON(w, utils.HttpStatusFromError(err), map[string]string{"error": err.Error()})
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{"message": "profile updated"})
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(userIdStr)
	if err != nil || id <= 0 {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": errors.ErrInvalidField.Error()})
		return
	}

	p, err := h.ProfileService.GetProfileByUserID(r.Context(), id)
	if err != nil {
		utils.RespondJSON(w, utils.HttpStatusFromError(err), map[string]string{"error": err.Error()})
		return
	}

	utils.RespondJSON(w, http.StatusOK, p)
}
