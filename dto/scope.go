package dto

type CreateScopeRequest struct {
	ScopeName string `json:"scope_name" binding:"required"`
}

type DeleteScopeRequest struct {
	ScopeName string `json:"scope_name" binding:"required"`
}
