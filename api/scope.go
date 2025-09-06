package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vnFuhung2903/vcs-user-management-service/dto"
	"github.com/vnFuhung2903/vcs-user-management-service/pkg/middlewares"
	"github.com/vnFuhung2903/vcs-user-management-service/usecases/services"
)

type scopeHandler struct {
	scopeService  services.IScopeService
	jwtMiddleware middlewares.IJWTMiddleware
}

func NewScopeHandler(scopeService services.IScopeService, jwtMiddleware middlewares.IJWTMiddleware) *scopeHandler {
	return &scopeHandler{scopeService, jwtMiddleware}
}

func (h *scopeHandler) SetupRoutes(r *gin.Engine) {
	scopeRoutes := r.Group("/scopes", h.jwtMiddleware.RequireScope("scope:manage"))
	{
		scopeRoutes.POST("/create", h.Create)
		scopeRoutes.GET("/list", h.ListAll)
		scopeRoutes.DELETE("/delete", h.Delete)
	}
}

// Create godoc
// @Summary Create a new scope
// @Description Create a scope (admin only)
// @Tags scopes
// @Accept json
// @Produce json
// @Param body body dto.CreateScopeRequest true "Scope creation request"
// @Success 201 {object} dto.APIResponse "New scope created successfully"
// @Failure 400 {object} dto.APIResponse "Bad request"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /scopes/create [post]
func (h *scopeHandler) Create(c *gin.Context) {
	var req dto.CreateScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	_, err := h.scopeService.Create(c.Request.Context(), req.ScopeName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to find scopes",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Code:    "SCOPE_CREATED",
		Message: "New scope created successfully",
	})
}

// ListAll godoc
// @Summary List all scopes
// @Description Retrieve all scopes (admin only)
// @Tags scopes
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse "Scopes retrieved successfully"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /scopes/ [get]
func (h *scopeHandler) ListAll(c *gin.Context) {
	scopes, err := h.scopeService.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to retrieve scopes",
			Error:   err.Error(),
		})
		return
	}

	names := make([]string, 0)
	for _, scope := range scopes {
		names = append(names, scope.Name)
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SCOPES_RETRIEVED",
		Message: "All scopes retrieved successfully",
		Data:    names,
	})
}

// Delete godoc
// @Summary Delete a scope
// @Description Delete a scope by name (admin only)
// @Tags scopes
// @Accept json
// @Produce json
// @Param body body dto.DeleteScopeRequest true "Scope deletion request"
// @Success 200 {object} dto.APIResponse "Scope deleted successfully"
// @Failure 400 {object} dto.APIResponse "Bad request"
// @Failure 500 {object} dto.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /scopes/delete [delete]
func (h *scopeHandler) Delete(c *gin.Context) {
	var req dto.DeleteScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Code:    "BAD_REQUEST",
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	err := h.scopeService.Delete(c.Request.Context(), req.ScopeName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to delete scope",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Code:    "SCOPE_DELETED",
		Message: "Scope deleted successfully",
	})
}
