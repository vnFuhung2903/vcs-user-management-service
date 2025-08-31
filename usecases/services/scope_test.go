package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	Logger "gorm.io/gorm/logger"

	"github.com/vnFuhung2903/vcs-user-management-service/entities"
	"github.com/vnFuhung2903/vcs-user-management-service/mocks/interfaces"
	"github.com/vnFuhung2903/vcs-user-management-service/mocks/logger"
	"github.com/vnFuhung2903/vcs-user-management-service/mocks/repositories"
)

type ScopeServiceSuite struct {
	suite.Suite
	ctrl         *gomock.Controller
	scopeService IScopeService
	mockRepo     *repositories.MockIScopeRepository
	mockRedis    *interfaces.MockIRedisClient
	logger       *logger.MockILogger
	ctx          context.Context
}

func (s *ScopeServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = repositories.NewMockIScopeRepository(s.ctrl)
	s.mockRedis = interfaces.NewMockIRedisClient(s.ctrl)
	s.logger = logger.NewMockILogger(s.ctrl)
	s.scopeService = NewScopeService(s.mockRepo, s.logger)
	s.ctx = context.Background()
}

func (s *ScopeServiceSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestScopeServiceSuite(t *testing.T) {
	suite.Run(t, new(ScopeServiceSuite))
}

func (s *ScopeServiceSuite) TestCreate() {
	name := "test"
	expected := &entities.UserScope{
		ID:   uint(1),
		Name: name,
	}

	s.mockRepo.EXPECT().Create(name).Return(expected, nil)
	s.logger.EXPECT().Info("new scope created successfully").Times(1)

	result, err := s.scopeService.Create(s.ctx, name)
	s.NoError(err)
	s.Equal(expected, result)
}

func (s *ScopeServiceSuite) TestCreateError() {
	name := "test"

	s.mockRepo.EXPECT().Create(name).Return(nil, errors.New("db error"))
	s.logger.EXPECT().Error("failed to create scope", gomock.Any()).Times(1)

	result, err := s.scopeService.Create(s.ctx, name)
	s.ErrorContains(err, "db error")
	s.Nil(result)
}

func (s *ScopeServiceSuite) TestFindOne() {
	name := "test"
	expected := &entities.UserScope{
		ID:   uint(1),
		Name: name,
	}

	s.mockRepo.EXPECT().FindByName(name).Return(expected, nil)
	s.logger.EXPECT().Info("scope found successfully").Times(1)

	result, err := s.scopeService.FindOne(s.ctx, name)
	s.NoError(err)
	s.Equal(expected, result)
}

func (s *ScopeServiceSuite) TestFindOneError() {
	name := "test"

	s.mockRepo.EXPECT().FindByName(name).Return(nil, errors.New("db error"))
	s.logger.EXPECT().Error("failed to find scope", gomock.Any()).Times(1)

	result, err := s.scopeService.FindOne(s.ctx, name)
	s.ErrorContains(err, "db error")
	s.Nil(result)
}

func (s *ScopeServiceSuite) TestFindMany() {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: Logger.Default.LogMode(Logger.Silent),
	})
	assert.NoError(s.T(), err)

	tx := gormDB.Begin()
	assert.NoError(s.T(), tx.Error)

	names := []string{"read", "write"}
	expectedScopes := []*entities.UserScope{
		{ID: uint(1), Name: "read"},
		{ID: uint(2), Name: "write"},
	}

	mockTxRepo := repositories.NewMockIScopeRepository(s.ctrl)

	s.mockRepo.EXPECT().BeginTransaction(s.ctx).Return(tx, nil)
	s.mockRepo.EXPECT().WithTransaction(tx).Return(mockTxRepo)

	mockTxRepo.EXPECT().FindByName("read").Return(expectedScopes[0], nil)
	mockTxRepo.EXPECT().FindByName("write").Return(expectedScopes[1], nil)

	s.logger.EXPECT().Info("all scopes found successfully").Times(1)

	result, err := s.scopeService.FindMany(s.ctx, names)
	s.NoError(err)
	s.Equal(expectedScopes, result)
}

func (s *ScopeServiceSuite) TestFindManyError() {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: Logger.Default.LogMode(Logger.Silent),
	})
	assert.NoError(s.T(), err)

	tx := gormDB.Begin()
	assert.NoError(s.T(), tx.Error)

	names := []string{"read", "write"}

	mockTxRepo := repositories.NewMockIScopeRepository(s.ctrl)

	s.mockRepo.EXPECT().BeginTransaction(s.ctx).Return(tx, nil)
	s.mockRepo.EXPECT().WithTransaction(tx).Return(mockTxRepo)

	mockTxRepo.EXPECT().FindByName("read").Return(&entities.UserScope{ID: 1, Name: "read"}, nil)
	mockTxRepo.EXPECT().FindByName("write").Return(nil, errors.New("scope not found"))

	s.logger.EXPECT().Error("failed to find scope", gomock.Any()).Times(1)

	result, err := s.scopeService.FindMany(s.ctx, names)
	s.ErrorContains(err, "scope not found")
	s.Nil(result)
}

func (s *ScopeServiceSuite) TestFindManyBeginTransactionError() {
	names := []string{"read", "write"}

	s.mockRepo.EXPECT().BeginTransaction(s.ctx).Return(nil, errors.New("transaction error"))
	s.logger.EXPECT().Error("failed to create transaction", gomock.Any()).Times(1)

	result, err := s.scopeService.FindMany(s.ctx, names)
	s.ErrorContains(err, "transaction error")
	s.Nil(result)
}
