package dto

type CreateUserRequest struct {
	Username string   `json:"username" binding:"required"`
	Password string   `json:"password" binding:"required"`
	Email    string   `json:"email" binding:"required,email"`
	Scopes   []string `json:"scopes" binding:"required"`
}

type UpdateScopeRequest struct {
	UserId  string `json:"user_id" binding:"required"`
	IsAdded bool   `json:"is_added"`
	Scope   string `json:"scopes" binding:"required"`
}

type DeleteRequest struct {
	UserId string `json:"user_id" binding:"required"`
}
