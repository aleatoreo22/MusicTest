package repository

import (
	"music-unified-backend/internal/models"
	"gorm.io/gorm"
)

type PlaylistRepository interface {
	Create(playlist *models.Playlist) error
	FindAllByUserID(userID uint) ([]models.Playlist, error)
	FindByID(id uint, userID uint) (*models.Playlist, error)
	AddTrack(playlistID uint, track *models.Track) error
}

type gormPlaylistRepository struct {
	db *gorm.DB
}

func NewGormPlaylistRepository(db *gorm.DB) PlaylistRepository {
	return &gormPlaylistRepository{db: db}
}

func (r *gormPlaylistRepository) Create(playlist *models.Playlist) error {
	return r.db.Create(playlist).Error
}

func (r *gormPlaylistRepository) FindAllByUserID(userID uint) ([]models.Playlist, error) {
	var playlists []models.Playlist
	err := r.db.Where("user_id = ?", userID).Preload("Tracks").Find(&playlists).Error
	return playlists, err
}

func (r *gormPlaylistRepository) FindByID(id uint, userID uint) (*models.Playlist, error) {
	var playlist models.Playlist
	err := r.db.Where("id = ? AND user_id = ?", id, userID).Preload("Tracks").First(&playlist).Error
	if err != nil {
		return nil, err
	}
	return &playlist, nil
}

func (r *gormPlaylistRepository) AddTrack(playlistID uint, track *models.Track) error {
	// Primeiro, garante que a track existe ou cria uma nova
	// Para simplificar, sempre tentamos encontrar por Source e SourceID ou LocalPath
	var existingTrack models.Track
	query := r.db.Where("source = ?", track.Source)
	if track.Source == models.SourceLocal {
		query = query.Where("local_path = ?", track.LocalPath)
	} else {
		query = query.Where("source_id = ?", track.SourceID)
	}

	err := query.First(&existingTrack).Error
	if err != nil {
		// Se não existe, cria
		if err := r.db.Create(track).Error; err != nil {
			return err
		}
	} else {
		// Se existe, usa a ID existente
		track.ID = existingTrack.ID
	}

	// Adiciona a música à playlist (tabela many-to-many)
	var playlist models.Playlist
	playlist.ID = playlistID
	return r.db.Model(&playlist).Association("Tracks").Append(track)
}
