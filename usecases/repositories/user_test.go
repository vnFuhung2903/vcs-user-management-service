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

type UserRepoSuite struct {
	suite.Suite
	db   *gorm.DB
	repo IUserRepository
}

func (suite *UserRepoSuite) SetupTest() {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(suite.T(), err)
	err = gormDB.AutoMigrate(&entities.User{})
	assert.NoError(suite.T(), err)
	suite.db = gormDB
	suite.repo = NewUserRepository(gormDB)
}

func (suite *UserRepoSuite) TearDownTest() {
	sqlDB, err := suite.db.DB()
	assert.NoError(suite.T(), err)
	sqlDB.Close()
}

func TestUserRepoSuite(t *testing.T) {
	suite.Run(t, new(UserRepoSuite))
}

func (suite *UserRepoSuite) TestCreateAndFindById() {
	user, err := suite.repo.Create("test", "pass", "test@example.com", []*entities.UserScope{
		{Name: "read"},
		{Name: "write"},
	})
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)

	found, err := suite.repo.FindById(user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test", found.Username)
}

func (suite *UserRepoSuite) TestCreateDuplicateEmail() {
	_, err := suite.repo.Create("test", "pass", "test@example.com", []*entities.UserScope{})
	assert.NoError(suite.T(), err)

	_, err = suite.repo.Create("testnil", "pass", "test@example.com", []*entities.UserScope{})
	assert.Error(suite.T(), err)
}

func (suite *UserRepoSuite) TestFindByIdNotFound() {
	_, err := suite.repo.FindById("non-existent-id")
	assert.Error(suite.T(), err)
}

func (suite *UserRepoSuite) TestUpdateScope() {
	user, _ := suite.repo.Create("test", "pass", "test@example.com", []*entities.UserScope{
		{Name: "read"},
	})
	err := suite.repo.UpdateScope(user, []*entities.UserScope{
		{Name: "admin"},
	})
	assert.NoError(suite.T(), err)
}

func (suite *UserRepoSuite) TestDelete() {
	user, _ := suite.repo.Create("test", "pass", "test@example.com", []*entities.UserScope{
		{Name: "read"},
	})
	err := suite.repo.Delete(user.ID)
	assert.NoError(suite.T(), err)

	_, err = suite.repo.FindById("test")
	assert.Error(suite.T(), err)
}

func (suite *UserRepoSuite) TestDeleteNonExistent() {
	err := suite.repo.Delete("not-exist")
	assert.NoError(suite.T(), err)
}

func (suite *UserRepoSuite) TestBeginTransactionError() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()

	_, err := suite.repo.BeginTransaction(context.Background())
	assert.Error(suite.T(), err)
}

func (suite *UserRepoSuite) TestBeginAndWithTransaction_Rollback() {
	tx, err := suite.repo.BeginTransaction(suite.T().Context())
	assert.NoError(suite.T(), err)

	txRepo := suite.repo.WithTransaction(tx)
	_, err = txRepo.Create("test", "pass", "test@example.com", []*entities.UserScope{
		{Name: "read"},
	})
	assert.NoError(suite.T(), err)

	tx.Rollback()
}

func (suite *UserRepoSuite) TestFindAll() {
	user1, err := suite.repo.Create("user1", "pass1", "user1@example.com", []*entities.UserScope{
		{Name: "read"},
	})
	assert.NoError(suite.T(), err)

	user2, err := suite.repo.Create("user2", "pass2", "user2@example.com", []*entities.UserScope{
		{Name: "write"},
	})
	assert.NoError(suite.T(), err)

	user3, err := suite.repo.Create("user3", "pass3", "user3@example.com", []*entities.UserScope{
		{Name: "admin"},
	})
	assert.NoError(suite.T(), err)

	users, err := suite.repo.FindAll()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), users, 3)

	usernames := make([]string, len(users))
	for i, user := range users {
		usernames[i] = user.Username
		assert.NotNil(suite.T(), user.Scopes)
	}
	assert.Contains(suite.T(), usernames, user1.Username)
	assert.Contains(suite.T(), usernames, user2.Username)
	assert.Contains(suite.T(), usernames, user3.Username)
}

func (suite *UserRepoSuite) TestFindAllDatabaseError() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()

	users, err := suite.repo.FindAll()
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), users)
}
