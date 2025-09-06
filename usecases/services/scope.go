package services

import (
	"context"

	"github.com/vnFuhung2903/vcs-user-management-service/entities"
	"github.com/vnFuhung2903/vcs-user-management-service/pkg/logger"
	"github.com/vnFuhung2903/vcs-user-management-service/usecases/repositories"
	"go.uber.org/zap"
)

type IScopeService interface {
	Create(ctx context.Context, scopeName string) (*entities.UserScope, error)
	FindOne(ctx context.Context, scopeName string) (*entities.UserScope, error)
	FindMany(ctx context.Context, scopeNames []string) ([]*entities.UserScope, error)
	FindAll(ctx context.Context) ([]*entities.UserScope, error)
	Delete(ctx context.Context, scopeName string) error
}

type scopeService struct {
	scopeRepo repositories.IScopeRepository
	logger    logger.ILogger
}

func NewScopeService(scopeRepo repositories.IScopeRepository, logger logger.ILogger) IScopeService {
	return &scopeService{
		scopeRepo: scopeRepo,
		logger:    logger,
	}
}

func (s *scopeService) Create(ctx context.Context, scopeName string) (*entities.UserScope, error) {
	scope, err := s.scopeRepo.Create(scopeName)
	if err != nil {
		s.logger.Error("failed to create scope", zap.Error(err))
		return nil, err
	}

	s.logger.Info("new scope created successfully")
	return scope, nil
}

func (s *scopeService) FindOne(ctx context.Context, scopeName string) (*entities.UserScope, error) {
	scope, err := s.scopeRepo.FindByName(scopeName)
	if err != nil {
		s.logger.Error("failed to find scope", zap.String("name", scopeName), zap.Error(err))
		return nil, err
	}

	s.logger.Info("scope found successfully")
	return scope, nil
}

func (s *scopeService) FindMany(ctx context.Context, scopeNames []string) ([]*entities.UserScope, error) {
	tx, err := s.scopeRepo.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error("failed to create transaction", zap.Error(err))
		return nil, err
	}

	txRepo := s.scopeRepo.WithTransaction(tx)
	scopes := make([]*entities.UserScope, 0, len(scopeNames))
	for _, scopeName := range scopeNames {
		scope, err := txRepo.FindByName(scopeName)
		if err != nil {
			s.logger.Error("failed to find scope", zap.String("name", scopeName), zap.Error(err))
			tx.Rollback()
			return nil, err
		}
		scopes = append(scopes, scope)
	}
	if err := tx.Commit().Error; err != nil {
		s.logger.Error("failed to commit transaction", zap.Error(err))
		return nil, err
	}

	s.logger.Info("all scopes found successfully")
	return scopes, nil
}

func (s *scopeService) FindAll(ctx context.Context) ([]*entities.UserScope, error) {
	scopes, err := s.scopeRepo.FindAll()
	if err != nil {
		s.logger.Error("failed to find all scopes", zap.Error(err))
		return nil, err
	}

	s.logger.Info("all scopes retrieved successfully")
	return scopes, nil
}

func (s *scopeService) Delete(ctx context.Context, scopeName string) error {
	err := s.scopeRepo.Delete(scopeName)
	if err != nil {
		s.logger.Error("failed to delete scope", zap.Error(err))
		return err
	}

	s.logger.Info("scope deleted successfully", zap.String("name", scopeName))
	return nil
}
