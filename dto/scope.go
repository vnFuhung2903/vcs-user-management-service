package dto

type CreateScopeRequest struct {
	ScopeName string `json:"scope_name" binding:"required"`
}
