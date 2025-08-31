package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/vnFuhung2903/vcs-user-management-service/dto"
	"github.com/vnFuhung2903/vcs-user-management-service/entities"
	"github.com/vnFuhung2903/vcs-user-management-service/mocks/middlewares"
	"github.com/vnFuhung2903/vcs-user-management-service/mocks/services"
)

type UserHandlerSuite struct {
	suite.Suite
	ctrl         *gomock.Controller
	userHandler  *userHandler
	mockUserSvc  *services.MockIUserService
	mockScopeSvc *services.MockIScopeService
	mockJWT      *middlewares.MockIJWTMiddleware
	router       *gin.Engine
}

func (s *UserHandlerSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.ctrl = gomock.NewController(s.T())
	s.mockUserSvc = services.NewMockIUserService(s.ctrl)
	s.mockScopeSvc = services.NewMockIScopeService(s.ctrl)
	s.mockJWT = middlewares.NewMockIJWTMiddleware(s.ctrl)

	s.userHandler = NewUserHandler(s.mockScopeSvc, s.mockUserSvc, s.mockJWT)
	s.router = gin.New()

	// Mock the middleware to always pass
	s.mockJWT.EXPECT().RequireScope("user:manage").Return(func(c *gin.Context) {
		c.Next()
	}).AnyTimes()

	s.userHandler.SetupRoutes(s.router)
}

func (s *UserHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestUserHandlerSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerSuite))
}

func (s *UserHandlerSuite) TestCreate() {
	req := dto.CreateUserRequest{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
		Scopes:   []string{"read", "write"},
	}

	expectedScopes := []*entities.UserScope{
		{ID: 1, Name: "read"},
		{ID: 2, Name: "write"},
	}

	expectedUser := &entities.User{
		ID:       "user-123",
		Username: "testuser",
		Email:    "test@example.com",
		Scopes:   expectedScopes,
	}

	s.mockScopeSvc.EXPECT().FindMany(gomock.Any(), req.Scopes).Return(expectedScopes, nil)
	s.mockUserSvc.EXPECT().Create(req.Username, req.Password, req.Email, expectedScopes).Return(expectedUser, nil)

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/users/create", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(s.T(), http.StatusCreated, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.True(s.T(), response.Success)
	assert.Equal(s.T(), "USER_CREATED", response.Code)
	assert.Equal(s.T(), "New user created successfully", response.Message)
}

func (s *UserHandlerSuite) TestCreateInvalidInput() {
	req := dto.CreateUserRequest{
		Username: "",
		Password: "password123",
		Email:    "test@example.com",
		Scopes:   []string{"read"},
	}

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/users/create", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.False(s.T(), response.Success)
	assert.Equal(s.T(), "BAD_REQUEST", response.Code)
}

func (s *UserHandlerSuite) TestCreateScopeServiceError() {
	req := dto.CreateUserRequest{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
		Scopes:   []string{"read", "write"},
	}

	s.mockScopeSvc.EXPECT().FindMany(gomock.Any(), req.Scopes).Return(nil, errors.New("scope not found"))

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/users/create", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.False(s.T(), response.Success)
	assert.Equal(s.T(), "INTERNAL_SERVER_ERROR", response.Code)
	assert.Equal(s.T(), "Failed to find scopes", response.Message)
}

func (s *UserHandlerSuite) TestCreateUserServiceError() {
	req := dto.CreateUserRequest{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
		Scopes:   []string{"read"},
	}

	expectedScopes := []*entities.UserScope{
		{ID: 1, Name: "read"},
	}

	s.mockScopeSvc.EXPECT().FindMany(gomock.Any(), req.Scopes).Return(expectedScopes, nil)
	s.mockUserSvc.EXPECT().Create(req.Username, req.Password, req.Email, expectedScopes).Return(nil, errors.New("user creation failed"))

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/users/create", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.False(s.T(), response.Success)
	assert.Equal(s.T(), "INTERNAL_SERVER_ERROR", response.Code)
	assert.Equal(s.T(), "Failed to register user", response.Message)
}

func (s *UserHandlerSuite) TestUpdateScope() {
	req := dto.UpdateScopeRequest{
		UserId:  "user-123",
		Scope:   "admin",
		IsAdded: true,
	}

	expectedScope := &entities.UserScope{
		ID:   1,
		Name: "admin",
	}

	s.mockScopeSvc.EXPECT().FindOne(gomock.Any(), req.Scope).Return(expectedScope, nil)
	s.mockUserSvc.EXPECT().UpdateScope(gomock.Any(), req.UserId, expectedScope, req.IsAdded).Return(nil)

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("PUT", "/users/update/scope", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.True(s.T(), response.Success)
	assert.Equal(s.T(), "USER_SCOPE_UPDATED", response.Code)
	assert.Equal(s.T(), "User scope updated successfully", response.Message)
}

func (s *UserHandlerSuite) TestUpdateScopeInvalidJSON() {
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("PUT", "/users/update/scope", bytes.NewBufferString("invalid json"))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.False(s.T(), response.Success)
	assert.Equal(s.T(), "BAD_REQUEST", response.Code)
}

func (s *UserHandlerSuite) TestUpdateScopeNotFound() {
	req := dto.UpdateScopeRequest{
		UserId:  "user-123",
		Scope:   "nonexistent",
		IsAdded: true,
	}

	s.mockScopeSvc.EXPECT().FindOne(gomock.Any(), req.Scope).Return(nil, errors.New("scope not found"))

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("PUT", "/users/update/scope", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.False(s.T(), response.Success)
	assert.Equal(s.T(), "INTERNAL_SERVER_ERROR", response.Code)
	assert.Equal(s.T(), "Failed to find scope", response.Message)
}

func (s *UserHandlerSuite) TestUpdateScopeUserServiceError() {
	req := dto.UpdateScopeRequest{
		UserId:  "user-123",
		Scope:   "admin",
		IsAdded: true,
	}

	expectedScope := &entities.UserScope{
		ID:   1,
		Name: "admin",
	}

	s.mockScopeSvc.EXPECT().FindOne(gomock.Any(), req.Scope).Return(expectedScope, nil)
	s.mockUserSvc.EXPECT().UpdateScope(gomock.Any(), req.UserId, expectedScope, req.IsAdded).Return(errors.New("update failed"))

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("PUT", "/users/update/scope", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.False(s.T(), response.Success)
	assert.Equal(s.T(), "INTERNAL_SERVER_ERROR", response.Code)
	assert.Equal(s.T(), "Failed to update user scope", response.Message)
}

func (s *UserHandlerSuite) TestDelete() {
	req := dto.DeleteRequest{
		UserId: "user-123",
	}

	s.mockUserSvc.EXPECT().Delete(gomock.Any(), req.UserId).Return(nil)

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("DELETE", "/users/delete", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.True(s.T(), response.Success)
	assert.Equal(s.T(), "USER_DELETED", response.Code)
	assert.Equal(s.T(), "User deleted successfully", response.Message)
}

func (s *UserHandlerSuite) TestDeleteInvalidInput() {
	req := dto.DeleteRequest{
		UserId: "",
	}

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("DELETE", "/users/delete", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.False(s.T(), response.Success)
	assert.Equal(s.T(), "BAD_REQUEST", response.Code)
}

func (s *UserHandlerSuite) TestDeleteUserServiceError() {
	req := dto.DeleteRequest{
		UserId: "user-123",
	}

	s.mockUserSvc.EXPECT().Delete(gomock.Any(), req.UserId).Return(errors.New("delete failed"))

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("DELETE", "/users/delete", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.False(s.T(), response.Success)
	assert.Equal(s.T(), "INTERNAL_SERVER_ERROR", response.Code)
	assert.Equal(s.T(), "Failed to delete user", response.Message)
}
