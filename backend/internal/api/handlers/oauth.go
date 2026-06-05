package handlers

import (
	"music-unified-backend/internal/models"
	"golang.org/x/oauth2"
)

// userToToken converte os tokens persistidos no User em oauth2.Token.
// Centraliza essa conversão para evitar duplicação entre handlers
// que precisam do client do Spotify (player, search).
func userToToken(user *models.User) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  user.SpotifyAccessToken,
		RefreshToken: user.SpotifyRefreshToken,
		Expiry:       user.SpotifyTokenExpiry,
		TokenType:    "Bearer",
	}
}
