package repository

import (
	"music-unified-backend/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitDB inicializa a conexão com o banco de dados e executa migrations
func InitDB() (*gorm.DB, error) {
	// Para SQLite, o DSN é o nome do arquivo
	db, err := gorm.Open(sqlite.Open("music.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Executa migrations automáticas
	err = db.AutoMigrate(&models.User{}, &models.Track{}, &models.Playlist{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
