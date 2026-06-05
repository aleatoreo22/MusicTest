package repository

import (
	"music-unified-backend/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitDB abre a conexão com o banco, executa as migrations e retorna
// o *gorm.DB pronto para uso. O DSN é interpretado por cada driver;
// para SQLite é o caminho do arquivo (ex: "music.db" ou
// "file::memory:?cache=shared"), para PostgreSQL seria uma URL
// completa começando com "postgres://".
func InitDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.User{}, &models.Track{}, &models.Playlist{}); err != nil {
		return nil, err
	}

	return db, nil
}
