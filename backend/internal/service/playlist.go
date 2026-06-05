package service

import (
	"music-unified-backend/internal/models"
	"music-unified-backend/internal/repository"
	"music-unified-backend/pkg/httputil"
)

// PlaylistService encapsula as regras de playlist: validação de
// ownership, montagem da Track a partir do request, e delegação
// ao repository.
type PlaylistService struct {
	playlists repository.PlaylistRepository
}

func NewPlaylistService(playlists repository.PlaylistRepository) *PlaylistService {
	return &PlaylistService{playlists: playlists}
}

func (s *PlaylistService) Create(userID uint, req models.CreatePlaylistRequest) (*models.Playlist, *httputil.APIError) {
	playlist := models.Playlist{Name: req.Name, UserID: userID}
	if err := s.playlists.Create(&playlist); err != nil {
		return nil, httputil.Internal("Erro ao criar playlist")
	}
	return &playlist, nil
}

func (s *PlaylistService) List(userID uint) ([]models.Playlist, *httputil.APIError) {
	playlists, err := s.playlists.FindAllByUserID(userID)
	if err != nil {
		return nil, httputil.Internal("Erro ao listar playlists")
	}
	return playlists, nil
}

func (s *PlaylistService) AddTrack(userID uint, playlistID uint, req models.AddTrackRequest) *httputil.APIError {
	if _, err := s.playlists.FindByID(playlistID, userID); err != nil {
		return httputil.NotFound("Playlist não encontrada ou acesso negado")
	}
	track := models.Track{
		Title:     req.Title,
		Artist:    req.Artist,
		Duration:  req.Duration,
		Thumbnail: req.Thumbnail,
		Source:    req.Source,
		SourceID:  req.SourceID,
		LocalPath: req.LocalPath,
	}
	if err := s.playlists.AddTrack(playlistID, &track); err != nil {
		return httputil.Internal("Erro ao adicionar música")
	}
	return nil
}
