package service

import (
	"context"
	"fmt"
	"music-unified-backend/internal/models"
	"music-unified-backend/internal/providers/spotify"
	"music-unified-backend/internal/repository"
	"music-unified-backend/pkg/httputil"

	"golang.org/x/oauth2"
)

// SpotifyService concentra o fluxo OAuth e a busca com token do
// usuário. Mantém o handler fino: recebe request, chama service,
// devolve resultado.
type SpotifyService struct {
	users    repository.UserRepository
	provider *spotify.Provider
}

func NewSpotifyService(users repository.UserRepository, provider *spotify.Provider) *SpotifyService {
	return &SpotifyService{users: users, provider: provider}
}

// LoginURL gera a URL de autorização com userID como state.
func (s *SpotifyService) LoginURL(userID uint) string {
	return s.provider.GetAuthURL(fmt.Sprintf("%d", userID))
}

// Callback troca o code por tokens e persiste no User.
func (s *SpotifyService) Callback(ctx context.Context, code, state string) (*models.User, *httputil.APIError) {
	token, err := s.provider.ExchangeCode(ctx, code)
	if err != nil {
		return nil, httputil.Internal("Falha ao trocar código por token")
	}
	var userID uint
	fmt.Sscanf(state, "%d", &userID)
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, httputil.NotFound("Usuário não encontrado")
	}
	user.SpotifyAccessToken = token.AccessToken
	user.SpotifyRefreshToken = token.RefreshToken
	user.SpotifyTokenExpiry = token.Expiry
	if err := s.users.Update(user); err != nil {
		return nil, httputil.Internal("Falha ao salvar tokens")
	}
	return user, nil
}

// IsLinked consulta o repository para checar se o usuário tem
// o Spotify vinculado.
func (s *SpotifyService) IsLinked(userID uint) (bool, *httputil.APIError) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return false, httputil.NotFound("Usuário não encontrado")
	}
	return user.SpotifyAccessToken != "", nil
}

// Search executa a busca com o token do próprio usuário.
func (s *SpotifyService) Search(ctx context.Context, userID uint, query string) ([]models.Track, *httputil.APIError) {
	user, err := s.users.FindByID(userID)
	if err != nil || user.SpotifyAccessToken == "" {
		return nil, httputil.Unauthorized("Spotify não vinculado ou usuário não encontrado")
	}
	tracks, err := s.provider.SearchFromToken(ctx, userToToken(user), query)
	if err != nil {
		return nil, httputil.Internal("Erro ao buscar no Spotify")
	}
	return tracks, nil
}

// userToToken é mantido em service para que os handlers não
// manipulem oauth2.Token diretamente.
func userToToken(user *models.User) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  user.SpotifyAccessToken,
		RefreshToken: user.SpotifyRefreshToken,
		Expiry:       user.SpotifyTokenExpiry,
		TokenType:    "Bearer",
	}
}
