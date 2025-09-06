package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/vnFuhung2903/vcs-user-management-service/entities"
	"github.com/vnFuhung2903/vcs-user-management-service/mocks/interfaces"
	"github.com/vnFuhung2903/vcs-user-management-service/mocks/logger"
	"github.com/vnFuhung2903/vcs-user-management-service/mocks/repositories"
)

type UserServiceSuite struct {
	suite.Suite
	ctrl        *gomock.Controller
	userService IUserService
	mockRepo    *repositories.MockIUserRepository
	mockRedis   *interfaces.MockIRedisClient
	logger      *logger.MockILogger
	ctx         context.Context
}

func (s *UserServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = repositories.NewMockIUserRepository(s.ctrl)
	s.mockRedis = interfaces.NewMockIRedisClient(s.ctrl)
	s.logger = logger.NewMockILogger(s.ctrl)
	s.userService = NewUserService(s.mockRepo, s.mockRedis, s.logger)
	s.ctx = context.Background()
}

func (s *UserServiceSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceSuite))
}

func (s *UserServiceSuite) TestCreate() {
	username := "testuser"
	password := "password123"
	email := "test@example.com"
	scopes := []*entities.UserScope{}

	expected := &entities.User{
		ID:       "test-id",
		Username: username,
		Email:    email,
		Scopes:   scopes,
	}

	s.mockRepo.EXPECT().Create(username, gomock.Any(), email, scopes).Return(expected, nil)
	s.logger.EXPECT().Info("new user registered successfully").Times(1)

	result, err := s.userService.Create(username, password, email, scopes)
	s.NoError(err)
	s.Equal(expected, result)
}

func (s *UserServiceSuite) TestCreateInvalidEmail() {
	username := "testuser"
	password := "password123"
	email := "invalid-email"
	scopes := []*entities.UserScope{}

	s.logger.EXPECT().Error("failed to parse email", gomock.Any()).Times(1)

	result, err := s.userService.Create(username, password, email, scopes)
	s.Error(err)
	s.Nil(result)
}

func (s *UserServiceSuite) TestCreateError() {
	username := "testuser"
	password := "password123"
	email := "test@example.com"
	scopes := []*entities.UserScope{}

	s.mockRepo.EXPECT().Create(username, gomock.Any(), email, scopes).Return(nil, errors.New("db error"))
	s.logger.EXPECT().Error("failed to create user", gomock.Any()).Times(1)

	result, err := s.userService.Create(username, password, email, scopes)
	s.ErrorContains(err, "db error")
	s.Nil(result)
}

func (s *UserServiceSuite) TestUpdateScopeAdd() {
	userId := "test-id"
	scopes := []*entities.UserScope{
		{Name: "user:read", ID: 2},
		{Name: "user:modify", ID: 1},
	}
	newScope := &entities.UserScope{
		Name: "user:modify",
		ID:   1,
	}
	existingUser := &entities.User{
		ID:     userId,
		Scopes: scopes,
	}
	expectedScope := existingUser.Scopes

	s.mockRepo.EXPECT().FindById(userId).Return(existingUser, nil)
	s.mockRepo.EXPECT().UpdateScope(existingUser, expectedScope).Return(nil)
	s.mockRedis.EXPECT().Del(s.ctx, "refresh:"+userId).Return(nil)
	s.logger.EXPECT().Info("user's scopes updated successfully").Times(1)

	err := s.userService.UpdateScope(s.ctx, userId, newScope, true)
	s.NoError(err)
}

func (s *UserServiceSuite) TestUpdateScopeUserNotFound() {
	userId := "nonexistent-id"
	newScope := &entities.UserScope{
		Name: "user:modify",
		ID:   1,
	}

	s.mockRepo.EXPECT().FindById(userId).Return(nil, errors.New("user not found"))
	s.logger.EXPECT().Error("failed to find user by id", gomock.Any()).Times(1)

	err := s.userService.UpdateScope(s.ctx, userId, newScope, false)
	s.ErrorContains(err, "user not found")
}

func (s *UserServiceSuite) TestUpdateScopeRepoError() {
	userId := "test-id"
	scopes := []*entities.UserScope{}
	newScope := &entities.UserScope{
		Name: "user:modify",
		ID:   1,
	}
	existingUser := &entities.User{
		ID:     userId,
		Scopes: scopes,
	}
	expectedScope := append(existingUser.Scopes, newScope)

	s.mockRepo.EXPECT().FindById(userId).Return(existingUser, nil)
	s.mockRepo.EXPECT().UpdateScope(existingUser, expectedScope).Return(errors.New("update failed"))
	s.logger.EXPECT().Error("failed to update user's scopes", gomock.Any()).Times(1)

	err := s.userService.UpdateScope(s.ctx, userId, newScope, true)
	s.ErrorContains(err, "update failed")
}

func (s *UserServiceSuite) TestUpdateScopeRedisError() {
	userId := "test-id"
	scopes := []*entities.UserScope{}
	newScope := &entities.UserScope{
		Name: "user:modify",
		ID:   1,
	}
	existingUser := &entities.User{
		ID:     userId,
		Scopes: scopes,
	}

	expectedScope := append(existingUser.Scopes, newScope)

	s.mockRepo.EXPECT().FindById(userId).Return(existingUser, nil)
	s.mockRepo.EXPECT().UpdateScope(existingUser, expectedScope).Return(nil)
	s.mockRedis.EXPECT().Del(s.ctx, "refresh:"+userId).Return(errors.New("redis error"))
	s.logger.EXPECT().Error("failed to delete refresh token in redis", gomock.Any()).Times(1)

	err := s.userService.UpdateScope(s.ctx, userId, newScope, true)
	s.ErrorContains(err, "redis error")
}

func (s *UserServiceSuite) TestDelete() {
	userId := "test-id"

	s.mockRepo.EXPECT().Delete(userId).Return(nil)
	s.mockRedis.EXPECT().Del(s.ctx, "refresh:"+userId).Return(nil)
	s.logger.EXPECT().Info("user deleted successfully").Times(1)

	err := s.userService.Delete(s.ctx, userId)
	s.NoError(err)
}

func (s *UserServiceSuite) TestDeleteRepoError() {
	userId := "test-id"

	s.mockRepo.EXPECT().Delete(userId).Return(errors.New("delete failed"))
	s.logger.EXPECT().Error("failed to delete user", gomock.Any()).Times(1)

	err := s.userService.Delete(s.ctx, userId)
	s.ErrorContains(err, "delete failed")
}

func (s *UserServiceSuite) TestDeleteRedisError() {
	userId := "test-id"

	s.mockRepo.EXPECT().Delete(userId).Return(nil)
	s.mockRedis.EXPECT().Del(s.ctx, "refresh:"+userId).Return(errors.New("redis error"))
	s.logger.EXPECT().Error("failed to delete refresh token in redis", gomock.Any()).Times(1)

	err := s.userService.Delete(s.ctx, userId)
	s.ErrorContains(err, "redis error")
}

func (s *UserServiceSuite) TestFindAll() {
	expectedUsers := []*entities.User{
		{
			ID:       "user-1",
			Username: "user1",
			Email:    "user1@example.com",
			Scopes:   []*entities.UserScope{{ID: 1, Name: "read"}},
		},
		{
			ID:       "user-2",
			Username: "user2",
			Email:    "user2@example.com",
			Scopes:   []*entities.UserScope{{ID: 2, Name: "write"}},
		},
		{
			ID:       "user-3",
			Username: "user3",
			Email:    "user3@example.com",
			Scopes:   []*entities.UserScope{{ID: 3, Name: "admin"}},
		},
	}

	s.mockRepo.EXPECT().FindAll().Return(expectedUsers, nil)
	s.logger.EXPECT().Info("all users retrieved successfully").Times(1)

	result, err := s.userService.FindAll(s.ctx)
	s.NoError(err)
	s.Equal(expectedUsers, result)
}

func (s *UserServiceSuite) TestFindAllError() {
	s.mockRepo.EXPECT().FindAll().Return(nil, errors.New("database error"))
	s.logger.EXPECT().Error("failed to find all users", gomock.Any()).Times(1)

	result, err := s.userService.FindAll(s.ctx)
	s.ErrorContains(err, "database error")
	s.Nil(result)
}
