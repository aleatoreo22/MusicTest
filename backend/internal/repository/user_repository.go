package repository

import "music-unified-backend/internal/models"

// UserRepository define o contrato para persistência de usuários
type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	Update(user *models.User) error
}
