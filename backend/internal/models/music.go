package models

import "time"

// TrackSource define a origem da música
type TrackSource string

const (
	SourceSpotify    TrackSource = "spotify"
	SourceYTMusic    TrackSource = "ytmusic"
	SourceSoundCloud TrackSource = "soundcloud"
	SourceYouTube    TrackSource = "youtube"
	SourceLocal      TrackSource = "local"
)

// Track representa uma música unificada
type Track struct {
	ID         uint        `json:"id" gorm:"primaryKey"`
	Title      string      `json:"title" gorm:"not null"`
	Artist     string      `json:"artist"`
	Duration   int         `json:"duration"` // em segundos
	Thumbnail  string      `json:"thumbnail"`
	PreviewURL string      `json:"preview_url,omitempty"`
	Source     TrackSource `json:"source" gorm:"not null"`
	SourceID   string      `json:"source_id"` // ID no Spotify/YT/SC
	LocalPath  string      `json:"local_path,omitempty"` // Path se for local
	CreatedAt  time.Time   `json:"created_at"`
}

// Playlist representa uma lista de músicas de um usuário
type Playlist struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	Name      string    `json:"name" gorm:"not null"`
	Tracks    []Track   `json:"tracks" gorm:"many2many:playlist_tracks;"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreatePlaylistRequest define o corpo para criar uma playlist
type CreatePlaylistRequest struct {
	Name string `json:"name" binding:"required"`
}

// AddTrackRequest define o corpo para adicionar uma música
type AddTrackRequest struct {
	Title     string      `json:"title" binding:"required"`
	Artist    string      `json:"artist"`
	Duration  int         `json:"duration"`
	Thumbnail string      `json:"thumbnail"`
	Source    TrackSource `json:"source" binding:"required"`
	SourceID  string      `json:"source_id"`
	LocalPath string      `json:"local_path"`
}
