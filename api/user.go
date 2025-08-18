package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vnFuhung2903/vcs-user-management-service/dto"
	"github.com/vnFuhung2903/vcs-user-management-service/pkg/middlewares"
	"github.com/vnFuhung2903/vcs-user-management-service/usecases/services"
)

type UserHandler struct {
	userService   services.IUserService
	jwtMiddleware middlewares.IJWTMiddleware
}

func NewUserHandler(userService services.IUserService, jwtMiddleware middlewares.IJWTMiddleware) *UserHandler {
	return &UserHandler{userService, jwtMiddleware}
}

func (h *UserHandler) SetupRoutes(r *gin.Engine) {
	userRoutes := r.Group("/users", h.jwtMiddleware.RequireScope("user:manage"))
	{
		userRoutes.PUT("/update/role", h.UpdateRole)
		userRoutes.PUT("/update/scope", h.UpdateScope)
		userRoutes.DELETE("/delete", h.Delete)
	}
}

// UpdateRole godoc
// @Summary Update a user's role
// @Description Update role of a user (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Param body body dto.UpdateRoleRequest true "User ID and new role"
// @Success 200 {object} dto.APIResponse "Role updated successfully"
// @Failure 400 {object} dto.APIResponse "Bad request"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /users/update/role [put]
func (h *UserHandler) UpdateRole(c *gin.Context) {
	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := h.userService.UpdateRole(c.Request.Context(), req.UserId, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to update user role",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "ROLE_UPDATED",
		Message: "User role updated successfully",
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
func (h *UserHandler) UpdateScope(c *gin.Context) {
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

	if err := h.userService.UpdateScope(c.Request.Context(), req.UserId, req.Scope, req.IsAdded); err != nil {
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
		Code:    "SCOPE_UPDATED",
		Message: "User scope updated successfully",
	})
}

// Delete godoc
// @Summary Delete a user
// @Description Remove a user from the system (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Param body body dto.DeleteRequest true "User ID to delete"
// @Success 200 {object} dto.APIResponse "User deleted successfully"
// @Failure 400 {object} dto.APIResponse "Bad request"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /users/delete [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	var req dto.DeleteRequest
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
