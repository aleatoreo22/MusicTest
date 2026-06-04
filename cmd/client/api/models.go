package api

import "time"

type User struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
}

type Track struct {
	ID         uint      `json:"id"`
	Title      string    `json:"title"`
	Artist     string    `json:"artist"`
	Duration   int       `json:"duration"`
	Thumbnail  string    `json:"thumbnail"`
	PreviewURL string    `json:"preview_url"`
	Source     string    `json:"source"`
	SourceID   string    `json:"source_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type Playlist struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Tracks    []Track   `json:"tracks"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
