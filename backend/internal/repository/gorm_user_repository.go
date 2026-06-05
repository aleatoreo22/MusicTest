package repository

import (
	"music-unified-backend/internal/models"
	"gorm.io/gorm"
)

type gormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository cria uma nova instância do repositório usando GORM
func NewGormUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *gormUserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}
