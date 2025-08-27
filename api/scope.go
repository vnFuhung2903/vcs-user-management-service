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
	userRoutes := r.Group("/scopes", h.jwtMiddleware.RequireScope("scope:manage"))
	{
		userRoutes.POST("/create/:scope", h.Create)
	}
}

// CreateScope godoc
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
