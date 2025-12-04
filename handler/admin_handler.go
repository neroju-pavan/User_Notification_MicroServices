package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"test123/service"
	"test123/utils"

	"github.com/go-chi/chi/v5"
)

type AdminHandler struct {
	Service     *service.RoleService
	RoleService *service.UserRoleService
	UserService *service.UserService
}

func NewAdminHandler(service *service.RoleService, roleService *service.UserRoleService, userService *service.UserService) *AdminHandler {
	return &AdminHandler{
		Service:     service,
		RoleService: roleService,
		UserService: userService,
	}
}
func (h *AdminHandler) CreateRole(w http.ResponseWriter, r *http.Request) {

	type req struct {
		Name string `json:"name"`
	}

	var body req

	// Parse JSON
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	// Validate
	if body.Name == "" {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "role name is required",
		})
		return
	}

	// Call service
	err := h.Service.CreateRole(r.Context(), body.Name)
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	// Success
	utils.RespondJSON(w, http.StatusCreated, map[string]string{
		"message": "role created successfully",
	})
}
func (h *AdminHandler) AddRoleToUser(w http.ResponseWriter, r *http.Request) {

	type req struct {
		UserID int    `json:"user_id"`
		Role   string `json:"role"`
	}

	var body req

	// parse body safely
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	// validate
	if body.UserID == 0 || body.Role == "" {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "user_id and role are required",
		})
		return
	}

	// call repo
	err := h.RoleService.AddUserRole(r.Context(), body.Role, body.UserID)
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Dtabase Failure",
		})
		return
	}

	// success response
	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "role assigned to user",
	})
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {

	userIdString := chi.URLParam(r, "Id")
	if userIdString == "" {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "missing userId param",
		})
		return
	}

	userId, err := strconv.Atoi(userIdString)
	if err != nil {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid userId",
		})
		return
	}

	err = h.UserService.DeleteUser(r.Context(), userId)
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "user deleted successfully",
	})
}

func (h *AdminHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {

	userIdString := chi.URLParam(r, "Id")
	if userIdString == "" {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "missing userId param",
		})
		return
	}

	userId, err := strconv.Atoi(userIdString)
	if err != nil {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid userId",
		})
		return
	}

	user, err := h.UserService.GetUserByID(r.Context(), userId)
	if err != nil {
		utils.RespondJSON(w, http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
		return
	}

	utils.RespondJSON(w, http.StatusOK, user)
}
