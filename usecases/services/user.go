package services

import (
	"context"

	"github.com/vnFuhung2903/vcs-user-management-service/entities"
	"github.com/vnFuhung2903/vcs-user-management-service/interfaces"
	"github.com/vnFuhung2903/vcs-user-management-service/pkg/logger"
	"github.com/vnFuhung2903/vcs-user-management-service/usecases/repositories"
	"go.uber.org/zap"
)

type IUserService interface {
	UpdateRole(ctx context.Context, userId string, role entities.UserRole) error
	UpdateScope(ctx context.Context, userId string, scope string, isAdded bool) error
	Delete(ctx context.Context, userId string) error
}

type userService struct {
	scopeRepo   repositories.IScopeRepository
	userRepo    repositories.IUserRepository
	redisClient interfaces.IRedisClient
	logger      logger.ILogger
}

func NewUserService(scopeRepo repositories.IScopeRepository, userRepo repositories.IUserRepository, redisClient interfaces.IRedisClient, logger logger.ILogger) IUserService {
	return &userService{
		scopeRepo:   scopeRepo,
		userRepo:    userRepo,
		redisClient: redisClient,
		logger:      logger,
	}
}

func (s *userService) UpdateRole(ctx context.Context, userId string, role entities.UserRole) error {
	user, err := s.userRepo.FindById(userId)
	if err != nil {
		s.logger.Error("failed to find user by id", zap.Error(err))
		return err
	}
	if err := s.userRepo.UpdateRole(user, role); err != nil {
		s.logger.Error("failed to update user's role", zap.Error(err))
		return err
	}

	s.logger.Info("user's role updated successfully")
	return nil
}

func (s *userService) UpdateScope(ctx context.Context, userId string, scope string, isAdded bool) error {
	user, err := s.userRepo.FindById(userId)
	if err != nil {
		s.logger.Error("failed to find user by id", zap.Error(err))
		return err
	}

	scopeFound, err := s.scopeRepo.FindByName(scope)
	if err != nil {
		s.logger.Error("failed to find scope by name", zap.Error(err))
	}

	scopeList := make([]entities.UserScope, len(user.Scopes))
	for _, s := range user.Scopes {
		if s.ID == scopeFound.ID {
			continue
		}
		scopeList = append(scopeList, s)
	}
	if isAdded {
		scopeList = append(scopeList, *scopeFound)
	}

	if err := s.userRepo.UpdateScope(user, scopeList); err != nil {
		s.logger.Error("failed to update user's scopes", zap.Error(err))
		return err
	}

	if err := s.redisClient.Del(ctx, "refresh:"+user.ID); err != nil {
		s.logger.Error("failed to delete refresh token in redis", zap.Error(err))
		return err
	}

	s.logger.Info("user's scopes updated successfully")
	return nil
}

func (s *userService) Delete(ctx context.Context, userId string) error {
	if err := s.userRepo.Delete(userId); err != nil {
		s.logger.Error("failed to delete user", zap.Error(err))
		return err
	}

	if err := s.redisClient.Del(ctx, "refresh:"+userId); err != nil {
		s.logger.Error("failed to delete refresh token in redis", zap.Error(err))
		return err
	}

	s.logger.Info("user deleted successfully")
	return nil
}
