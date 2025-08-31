package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/vnFuhung2903/vcs-user-management-service/entities"
)

type ScopeRepoSuite struct {
	suite.Suite
	db   *gorm.DB
	repo IScopeRepository
}

func (suite *ScopeRepoSuite) SetupTest() {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(suite.T(), err)
	err = gormDB.AutoMigrate(&entities.User{})
	assert.NoError(suite.T(), err)
	suite.db = gormDB
	suite.repo = NewScopeRepository(gormDB)
}

func (suite *ScopeRepoSuite) TearDownTest() {
	sqlDB, err := suite.db.DB()
	assert.NoError(suite.T(), err)
	sqlDB.Close()
}

func TestScopeRepoSuite(t *testing.T) {
	suite.Run(t, new(ScopeRepoSuite))
}

func (suite *ScopeRepoSuite) TestCreateAndFindById() {
	scope, err := suite.repo.Create("test")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), scope)

	found, err := suite.repo.FindById(scope.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test", found.Name)

	found, err = suite.repo.FindByName(scope.Name)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uint(1), found.ID)
}

func (suite *ScopeRepoSuite) TestCreateDuplicateName() {
	_, err := suite.repo.Create("test")
	assert.NoError(suite.T(), err)

	_, err = suite.repo.Create("test")
	assert.Error(suite.T(), err)
}

func (suite *ScopeRepoSuite) TestFindNotFound() {
	_, err := suite.repo.FindById(1)
	assert.Error(suite.T(), err)

	_, err = suite.repo.FindByName("test")
	assert.Error(suite.T(), err)
}

func (suite *ScopeRepoSuite) TestDelete() {
	scope, _ := suite.repo.Create("test")
	err := suite.repo.Delete(scope.Name)
	assert.NoError(suite.T(), err)

	_, err = suite.repo.FindById(scope.ID)
	assert.Error(suite.T(), err)
}

func (suite *ScopeRepoSuite) TestDeleteNonExistent() {
	err := suite.repo.Delete("not-exist")
	assert.NoError(suite.T(), err)
}

func (suite *ScopeRepoSuite) TestBeginTransactionError() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()

	_, err := suite.repo.BeginTransaction(context.Background())
	assert.Error(suite.T(), err)
}

func (suite *ScopeRepoSuite) TestBeginAndWithTransaction_Rollback() {
	tx, err := suite.repo.BeginTransaction(suite.T().Context())
	assert.NoError(suite.T(), err)

	txRepo := suite.repo.WithTransaction(tx)
	_, err = txRepo.Create("test")
	assert.NoError(suite.T(), err)

	tx.Rollback()
}
