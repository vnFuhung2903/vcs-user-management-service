package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/vnFuhung2903/vcs-user-management-service/entities"

	"gorm.io/gorm"
)

type IUserRepository interface {
	FindById(userId string) (*entities.User, error)
	FindAll() ([]*entities.User, error)
	Create(username, hash, email string, scopes []*entities.UserScope) (*entities.User, error)
	UpdateScope(user *entities.User, scopes []*entities.UserScope) error
	Delete(userId string) error
	BeginTransaction(ctx context.Context) (*gorm.DB, error)
	WithTransaction(tx *gorm.DB) IUserRepository
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindById(userId string) (*entities.User, error) {
	var user entities.User
	res := r.db.Preload("Scopes").First(&user, entities.User{ID: userId})
	if res.Error != nil {
		return nil, res.Error
	}
	return &user, nil
}

func (r *userRepository) FindAll() ([]*entities.User, error) {
	var users []*entities.User
	res := r.db.Preload("Scopes").Find(&users)
	if res.Error != nil {
		return nil, res.Error
	}
	return users, nil
}

func (r *userRepository) Create(username, hash, email string, scopes []*entities.UserScope) (*entities.User, error) {
	newUser := &entities.User{
		ID:       uuid.New().String(),
		Username: username,
		Hash:     hash,
		Email:    email,
		Scopes:   scopes,
	}
	res := r.db.Create(newUser)
	if res.Error != nil {
		return nil, res.Error
	}
	return newUser, nil
}

func (r *userRepository) UpdateScope(user *entities.User, scopes []*entities.UserScope) error {
	err := r.db.Model(user).Association("Scopes").Replace(scopes)
	return err
}

func (r *userRepository) Delete(userId string) error {
	res := r.db.Where("id = ?", userId).Delete(&entities.User{})
	return res.Error
}

func (r *userRepository) BeginTransaction(ctx context.Context) (*gorm.DB, error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return tx, nil
}

func (r *userRepository) WithTransaction(tx *gorm.DB) IUserRepository {
	return &userRepository{db: tx}
}
