package dto

import "github.com/vnFuhung2903/vcs-user-management-service/entities"

type UpdateRoleRequest struct {
	UserId string            `json:"user_id" binding:"required"`
	Role   entities.UserRole `json:"role" binding:"required"`
}

type UpdateScopeRequest struct {
	UserId  string `json:"user_id" binding:"required"`
	IsAdded bool   `json:"is_added"`
	Scope   string `json:"scopes" binding:"required"`
}

type DeleteRequest struct {
	UserId string `json:"user_id" binding:"required"`
}
