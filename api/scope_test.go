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

type ScopeHandlerSuite struct {
	suite.Suite
	ctrl         *gomock.Controller
	scopeHandler *scopeHandler
	mockScopeSvc *services.MockIScopeService
	mockJWT      *middlewares.MockIJWTMiddleware
	router       *gin.Engine
}

func (s *ScopeHandlerSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.ctrl = gomock.NewController(s.T())
	s.mockScopeSvc = services.NewMockIScopeService(s.ctrl)
	s.mockJWT = middlewares.NewMockIJWTMiddleware(s.ctrl)

	s.scopeHandler = NewScopeHandler(s.mockScopeSvc, s.mockJWT)
	s.router = gin.New()

	s.mockJWT.EXPECT().RequireScope("scope:manage").Return(func(c *gin.Context) {
		c.Next()
	}).AnyTimes()

	s.scopeHandler.SetupRoutes(s.router)
}

func (s *ScopeHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestScopeHandlerSuite(t *testing.T) {
	suite.Run(t, new(ScopeHandlerSuite))
}

func (s *ScopeHandlerSuite) TestCreate() {
	req := dto.CreateScopeRequest{
		ScopeName: "test:read",
	}

	expectedScope := &entities.UserScope{
		ID:   1,
		Name: "test:read",
	}

	s.mockScopeSvc.EXPECT().Create(gomock.Any(), req.ScopeName).Return(expectedScope, nil)

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/scopes/create", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(s.T(), http.StatusCreated, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.True(s.T(), response.Success)
	assert.Equal(s.T(), "SCOPE_CREATED", response.Code)
	assert.Equal(s.T(), "New scope created successfully", response.Message)
}

func (s *ScopeHandlerSuite) TestCreateInvalidInput() {
	req := dto.CreateScopeRequest{
		ScopeName: "",
	}

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/scopes/create", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.False(s.T(), response.Success)
	assert.Equal(s.T(), "BAD_REQUEST", response.Code)
	assert.Equal(s.T(), "Invalid request data", response.Message)
}

func (s *ScopeHandlerSuite) TestCreateServiceError() {
	req := dto.CreateScopeRequest{
		ScopeName: "test:read",
	}

	s.mockScopeSvc.EXPECT().Create(gomock.Any(), req.ScopeName).Return(nil, errors.New("scope already exists"))

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/scopes/create", bytes.NewBuffer(body))
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

func (s *ScopeHandlerSuite) TestListAll() {
	expectedScopes := []*entities.UserScope{
		{ID: 1, Name: "test:read"},
		{ID: 2, Name: "test:write"},
	}

	s.mockScopeSvc.EXPECT().FindAll(gomock.Any()).Return(expectedScopes, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/scopes/list", nil)

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.True(s.T(), response.Success)
	assert.Equal(s.T(), "SCOPES_RETRIEVED", response.Code)
	assert.Equal(s.T(), "All scopes retrieved successfully", response.Message)
	assert.NotNil(s.T(), response.Data)
}

func (s *ScopeHandlerSuite) TestListAllServiceError() {
	s.mockScopeSvc.EXPECT().FindAll(gomock.Any()).Return(nil, errors.New("database error"))

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/scopes/list", nil)

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.False(s.T(), response.Success)
	assert.Equal(s.T(), "INTERNAL_SERVER_ERROR", response.Code)
	assert.Equal(s.T(), "Failed to retrieve scopes", response.Message)
}

func (s *ScopeHandlerSuite) TestDelete() {
	req := dto.DeleteScopeRequest{
		ScopeName: "test:read",
	}

	s.mockScopeSvc.EXPECT().Delete(gomock.Any(), req.ScopeName).Return(nil)

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("DELETE", "/scopes/delete", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.True(s.T(), response.Success)
	assert.Equal(s.T(), "SCOPE_DELETED", response.Code)
	assert.Equal(s.T(), "Scope deleted successfully", response.Message)
}

func (s *ScopeHandlerSuite) TestDeleteInvalidInput() {
	req := dto.DeleteScopeRequest{
		ScopeName: "", // Empty scope name
	}

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("DELETE", "/scopes/delete", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.False(s.T(), response.Success)
	assert.Equal(s.T(), "BAD_REQUEST", response.Code)
	assert.Equal(s.T(), "Invalid request data", response.Message)
}

func (s *ScopeHandlerSuite) TestDeleteInvalidJSON() {
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("DELETE", "/scopes/delete", bytes.NewBuffer([]byte("{invalid json}")))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.False(s.T(), response.Success)
	assert.Equal(s.T(), "BAD_REQUEST", response.Code)
	assert.Equal(s.T(), "Invalid request data", response.Message)
}

func (s *ScopeHandlerSuite) TestDeleteServiceError() {
	req := dto.DeleteScopeRequest{
		ScopeName: "test:read",
	}

	s.mockScopeSvc.EXPECT().Delete(gomock.Any(), req.ScopeName).Return(errors.New("scope not found"))

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("DELETE", "/scopes/delete", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)

	var response dto.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.False(s.T(), response.Success)
	assert.Equal(s.T(), "INTERNAL_SERVER_ERROR", response.Code)
	assert.Equal(s.T(), "Failed to delete scope", response.Message)
}
