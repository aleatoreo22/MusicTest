package repository

import (
	"music-unified-backend/internal/models"
	"gorm.io/gorm"
)

type MusicRepository interface {
	FindTrackByID(id uint) (*models.Track, error)
}

type gormMusicRepository struct {
	db *gorm.DB
}

func NewGormMusicRepository(db *gorm.DB) MusicRepository {
	return &gormMusicRepository{db: db}
}

func (r *gormMusicRepository) FindTrackByID(id uint) (*models.Track, error) {
	var track models.Track
	err := r.db.First(&track, id).Error
	if err != nil {
		return nil, err
	}
	return &track, nil
}
