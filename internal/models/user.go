package models

import (
	"time"
)

// User representa o usuário no sistema
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Password  string    `json:"-" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Tokens de integração Spotify
	SpotifyAccessToken  string    `json:"-"`
	SpotifyRefreshToken string    `json:"-"`
	SpotifyTokenExpiry  time.Time `json:"-"`
}

// LoginRequest define o corpo da requisição de login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest define o corpo da requisição de registro
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}
