package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"test123/errors"
	"test123/logger"
	"test123/models"
	"test123/requests"
	"test123/service"
	"test123/utils"

	"github.com/go-chi/chi/v5"
)

type UserHandlers struct {
	UserService *service.UserService
}

func NewUserHandler(us *service.UserService) *UserHandlers {
	return &UserHandlers{UserService: us}
}

// ----------------------------
// CREATE USER
// ----------------------------
func (h *UserHandlers) CreateUser(w http.ResponseWriter, r *http.Request) {

	logger.Info("CreateUser", "request received")

	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		logger.Warn("CreateUser", "Invalid JSON", map[string]interface{}{
			"error": err.Error(),
		})
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{
			"error": errors.ErrInvalidJSON.Error(),
		})
		return
	}

	if err := u.Validate(); err != nil {
		logger.Warn("CreateUser", "validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		utils.RespondJSON(w, utils.HttpStatusFromError(err), map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err := h.UserService.CreateUser(r.Context(), u); err != nil {
		logger.Error("CreateUser", "service failed", map[string]interface{}{
			"error": err.Error(),
		})
		utils.RespondJSON(w, utils.HttpStatusFromError(err), map[string]string{
			"error": err.Error(),
		})
		return
	}

	logger.Info("CreateUser", "User created successfully", map[string]interface{}{
		"email": u.Email,
	})

	utils.RespondJSON(w, utils.HttpStatusFromSuccess("created"), map[string]string{
		"message": "User created successfully",
	})
}

// ----------------------------
// GET ALL USERS (cursor pagination)
// ----------------------------
func (h *UserHandlers) GetAllUsers(w http.ResponseWriter, r *http.Request) {

	logger.Info("GetAllUsers", "request received")

	limitStr := r.URL.Query().Get("limit")
	cursorStr := r.URL.Query().Get("cursor")
	search := r.URL.Query().Get("search")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	limit := 5
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		} else {
			logger.Warn("GetAllUsers", "invalid limit value", map[string]interface{}{
				"limit": limitStr,
			})
		}
	}

	users, nextCursor, err := h.UserService.GetUsersWithFiltersCursor(
		r.Context(),
		limit,
		cursorStr,
		search,
		fromStr,
		toStr,
	)

	if err != nil {
		logger.Error("GetAllUsers", "service failed", map[string]interface{}{
			"error": err.Error(),
		})
		utils.RespondJSON(w, utils.HttpStatusFromError(err), map[string]string{
			"error": err.Error(),
		})
		return
	}

	logger.Info("GetAllUsers", "users retrieved", map[string]interface{}{
		"count": len(users),
	})

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"users":       users,
		"next_cursor": nextCursor,
	})
}

// ----------------------------
// UPDATE USER
// ----------------------------
func (h *UserHandlers) UpdateUser(w http.ResponseWriter, r *http.Request) {

	logger.Info("UpdateUser", "request received")

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logger.Warn("UpdateUser", "invalid id", map[string]interface{}{
			"id": idStr,
		})
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{
			"error": errors.ErrInvalidField.Error(),
		})
		return
	}

	var req requests.UserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("UpdateUser", "Invalid JSON", map[string]interface{}{
			"error": err.Error(),
		})
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{
			"error": errors.ErrInvalidJSON.Error(),
		})
		return
	}

	if err := req.Validate(); err != nil {
		logger.Warn("UpdateUser", "validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		utils.RespondJSON(w, utils.HttpStatusFromError(err), map[string]string{
			"error": err.Error(),
		})
		return
	}

	u := models.User{
		ID:           id,
		Name:         req.Name,
		Email:        req.Email,
		Username:     req.Username,
		MobileNumber: req.MobileNumber,
	}

	if err := h.UserService.UpdateUser(r.Context(), u); err != nil {
		logger.Error("UpdateUser", "service failed", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		utils.RespondJSON(w, utils.HttpStatusFromError(err), map[string]string{
			"error": err.Error(),
		})
		return
	}

	logger.Info("UpdateUser", "user updated", map[string]interface{}{
		"id": id,
	})

	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "User updated",
	})
}

// ----------------------------
// DELETE USER
// ----------------------------
func (h *UserHandlers) DeleteUser(w http.ResponseWriter, r *http.Request) {

	logger.Info("DeleteUser", "request received")

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logger.Warn("DeleteUser", "invalid id", map[string]interface{}{
			"id": idStr,
		})
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{
			"error": errors.ErrInvalidField.Error(),
		})
		return
	}

	if err := h.UserService.DeleteUser(r.Context(), id); err != nil {
		logger.Error("DeleteUser", "service failed", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		utils.RespondJSON(w, utils.HttpStatusFromError(err), map[string]string{
			"error": err.Error(),
		})
		return
	}

	logger.Info("DeleteUser", "user deleted", map[string]interface{}{
		"id": id,
	})

	utils.RespondJSON(w, utils.HttpStatusFromSuccess("deleted"), nil)
}

// ----------------------------
// CHECK USERNAME EXISTS
// ----------------------------
func (h *UserHandlers) CheckUsernameHandler(w http.ResponseWriter, r *http.Request) {

	username := chi.URLParam(r, "username")

	logger.Info("CheckUsernameHandler", "checking username", map[string]interface{}{
		"username": username,
	})

	exists, err := h.UserService.UsernameExists(r.Context(), username)
	if err != nil {
		logger.Error("CheckUsernameHandler", "service failed", map[string]interface{}{
			"error": err.Error(),
		})
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]bool{
		"exists": exists,
	})
}

// ----------------------------
// GET USER BY USERNAME
// ----------------------------
func (h *UserHandlers) GetUserByUsername(w http.ResponseWriter, r *http.Request) {

	username := chi.URLParam(r, "username")

	logger.Info("GetUserByUsername", "request received", map[string]interface{}{
		"username": username,
	})

	user, err := h.UserService.GetUserByUsername(r.Context(), username)
	if err != nil {
		logger.Warn("GetUserByUsername", "user not found", map[string]interface{}{
			"username": username,
		})
		utils.RespondJSON(w, http.StatusNotFound, map[string]string{
			"error": "user not found",
		})
		return
	}

	logger.Info("GetUserByUsername", "user found", map[string]interface{}{
		"id": user.ID,
	})

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"username": user.Username,
		"id":       user.ID,
	})
}

func (h *UserHandlers) GetUserById(w http.ResponseWriter, r *http.Request) {
	// Extract "id" from URL parameters
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "user id is required"})
		return
	}

	userId, err := strconv.Atoi(idStr)
	if err != nil {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	// Fetch user from service
	user, err := h.UserService.GetUserByID(r.Context(), userId)
	if err != nil {
		utils.RespondJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	// Respond with user data
	utils.RespondJSON(w, http.StatusOK, user)
}
