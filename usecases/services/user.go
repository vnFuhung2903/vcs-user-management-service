package services

import (
	"context"
	"net/mail"

	"github.com/vnFuhung2903/vcs-user-management-service/entities"
	"github.com/vnFuhung2903/vcs-user-management-service/interfaces"
	"github.com/vnFuhung2903/vcs-user-management-service/pkg/logger"
	"github.com/vnFuhung2903/vcs-user-management-service/usecases/repositories"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type IUserService interface {
	Create(username, password, email string, scopes []*entities.UserScope) (*entities.User, error)
	UpdateScope(ctx context.Context, userId string, scope *entities.UserScope, isAdded bool) error
	Delete(ctx context.Context, userId string) error
}

type userService struct {
	userRepo    repositories.IUserRepository
	redisClient interfaces.IRedisClient
	logger      logger.ILogger
}

func NewUserService(userRepo repositories.IUserRepository, redisClient interfaces.IRedisClient, logger logger.ILogger) IUserService {
	return &userService{
		userRepo:    userRepo,
		redisClient: redisClient,
		logger:      logger,
	}
}

func (s *userService) Create(username, password, email string, scopes []*entities.UserScope) (*entities.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err))
		return nil, err
	}

	mail, err := mail.ParseAddress(email)
	if err != nil {
		s.logger.Error("failed to parse email", zap.Error(err))
		return nil, err
	}

	user, err := s.userRepo.Create(username, string(hash), mail.Address, scopes)
	if err != nil {
		s.logger.Error("failed to create user", zap.Error(err))
		return nil, err
	}

	s.logger.Info("new user registered successfully")
	return user, nil
}

func (s *userService) UpdateScope(ctx context.Context, userId string, scope *entities.UserScope, isAdded bool) error {
	user, err := s.userRepo.FindById(userId)
	if err != nil {
		s.logger.Error("failed to find user by id", zap.Error(err))
		return err
	}

	scopeList := make([]*entities.UserScope, len(user.Scopes))
	for _, s := range user.Scopes {
		if s.ID == scope.ID {
			continue
		}
		scopeList = append(scopeList, s)
	}
	if isAdded {
		scopeList = append(scopeList, scope)
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
