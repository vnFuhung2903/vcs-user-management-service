package repositories

import (
	"github.com/google/uuid"
	"github.com/vnFuhung2903/vcs-user-management-service/entities"

	"gorm.io/gorm"
)

type IScopeRepository interface {
	FindById(scopeId string) (*entities.UserScope, error)
	FindByName(name string) (*entities.UserScope, error)
	Create(name string) (*entities.UserScope, error)
	Delete(name string) error
}

type scopeRepository struct {
	db *gorm.DB
}

func NewScopeRepository(db *gorm.DB) IScopeRepository {
	return &scopeRepository{db: db}
}

func (r *scopeRepository) FindById(Id string) (*entities.UserScope, error) {
	var scope entities.UserScope
	res := r.db.First(&scope, entities.UserScope{ID: Id})
	if res.Error != nil {
		return nil, res.Error
	}
	return &scope, nil
}

func (r *scopeRepository) FindByName(name string) (*entities.UserScope, error) {
	var scope entities.UserScope
	res := r.db.First(&scope, entities.UserScope{Name: name})
	if res.Error != nil {
		return nil, res.Error
	}
	return &scope, nil
}

func (r *scopeRepository) Create(name string) (*entities.UserScope, error) {
	newScope := &entities.UserScope{
		ID:   uuid.New().String(),
		Name: name,
	}
	res := r.db.Create(newScope)
	if res.Error != nil {
		return nil, res.Error
	}
	return newScope, nil
}

func (r *scopeRepository) Delete(name string) error {
	res := r.db.Where("name = ?", name).Delete(&entities.UserScope{})
	return res.Error
}
