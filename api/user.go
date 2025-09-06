package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vnFuhung2903/vcs-user-management-service/dto"
	"github.com/vnFuhung2903/vcs-user-management-service/pkg/middlewares"
	"github.com/vnFuhung2903/vcs-user-management-service/usecases/services"
)

type userHandler struct {
	scopeService  services.IScopeService
	userService   services.IUserService
	jwtMiddleware middlewares.IJWTMiddleware
}

func NewUserHandler(scopeService services.IScopeService, userService services.IUserService, jwtMiddleware middlewares.IJWTMiddleware) *userHandler {
	return &userHandler{scopeService, userService, jwtMiddleware}
}

func (h *userHandler) SetupRoutes(r *gin.Engine) {
	userRoutes := r.Group("/users", h.jwtMiddleware.RequireScope("user:manage"))
	{
		userRoutes.POST("/create", h.Create)
		userRoutes.GET("/list", h.ListAll)
		userRoutes.PUT("/update/scope", h.UpdateScope)
		userRoutes.DELETE("/delete", h.Delete)
	}
}

// Create godoc
// @Summary Create a new user
// @Description Create a user (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Param body body dto.CreateUserRequest true "User creation request"
// @Success 201 {object} dto.APIResponse "New user created successfully"
// @Failure 400 {object} dto.APIResponse "Bad request"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /users/create [post]
func (h *userHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	scopes, err := h.scopeService.FindMany(c.Request.Context(), req.Scopes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to find scopes",
			Error:   err.Error(),
		})
		return
	}

	_, err = h.userService.Create(req.Username, req.Password, req.Email, scopes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to register user",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Code:    "USER_CREATED",
		Message: "New user created successfully",
	})
}

// ListAll godoc
// @Summary List all users
// @Description Retrieve all users (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse "Users retrieved successfully"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /users/list [get]
func (h *userHandler) ListAll(c *gin.Context) {
	users, err := h.userService.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to retrieve users",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "USERS_RETRIEVED",
		Message: "All users retrieved successfully",
		Data:    users,
	})
}

// UpdateScope godoc
// @Summary Update a user's scope
// @Description Update permission scope of a user (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Param body body dto.UpdateScopeRequest true "User ID, scopes, and whether to add or remove"
// @Success 200 {object} dto.APIResponse "Scope updated successfully"
// @Failure 400 {object} dto.APIResponse "Bad request"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /users/update/scope [put]
func (h *userHandler) UpdateScope(c *gin.Context) {
	var req dto.UpdateScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	scope, err := h.scopeService.FindOne(c.Request.Context(), req.Scope)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to find scope",
			Error:   err.Error(),
		})
		return
	}

	if err := h.userService.UpdateScope(c.Request.Context(), req.UserId, scope, req.IsAdded); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to update user scope",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "USER_SCOPE_UPDATED",
		Message: "User scope updated successfully",
	})
}

// Delete godoc
// @Summary Delete a user
// @Description Remove a user from the system (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Param body body dto.DeleteUserRequest true "User ID to delete"
// @Success 200 {object} dto.APIResponse "User deleted successfully"
// @Failure 400 {object} dto.APIResponse "Bad request"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /users/delete [delete]
func (h *userHandler) Delete(c *gin.Context) {
	var req dto.DeleteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := h.userService.Delete(c.Request.Context(), req.UserId); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to delete user",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "USER_DELETED",
		Message: "User deleted successfully",
	})
}
