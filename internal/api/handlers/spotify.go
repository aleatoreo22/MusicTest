package handlers

import (
	"fmt"
	"music-unified-backend/internal/providers/spotify"
	"music-unified-backend/internal/repository"
	"music-unified-backend/pkg/httputil"

	"github.com/gin-gonic/gin"
)

type SpotifyHandler struct {
	repo     repository.UserRepository
	provider *spotify.Provider
}

func NewSpotifyHandler(repo repository.UserRepository, provider *spotify.Provider) *SpotifyHandler {
	return &SpotifyHandler{repo: repo, provider: provider}
}

// Login gera a URL de autorização do Spotify com o userID como state.
// O state é usado no callback para identificar qual usuário vinculou.
func (h *SpotifyHandler) Login(c *gin.Context) {
	userID, ok := httputil.MustUserID(c)
	if !ok {
		return
	}
	url := h.provider.GetAuthURL(fmt.Sprintf("%d", userID))
	httputil.JSON(c, 200, gin.H{"url": url})
}

// Callback recebe o code do Spotify, troca por tokens e persiste
// no registro do usuário identificado pelo state.
func (h *SpotifyHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	if code == "" {
		httputil.Error(c, httputil.BadRequest("Código não fornecido"))
		return
	}

	token, err := h.provider.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		httputil.Error(c, httputil.Internal("Falha ao trocar código por token"))
		return
	}

	var userID uint
	fmt.Sscanf(state, "%d", &userID)

	user, err := h.repo.FindByID(userID)
	if err != nil {
		httputil.Error(c, httputil.NotFound("Usuário não encontrado"))
		return
	}

	user.SpotifyAccessToken = token.AccessToken
	user.SpotifyRefreshToken = token.RefreshToken
	user.SpotifyTokenExpiry = token.Expiry

	if err := h.repo.Update(user); err != nil {
		httputil.Error(c, httputil.Internal("Falha ao salvar tokens"))
		return
	}

	httputil.JSON(c, 200, gin.H{"message": "Conta do Spotify vinculada com sucesso!"})
}

// GetStatus retorna se o usuário atual tem o Spotify vinculado.
func (h *SpotifyHandler) GetStatus(c *gin.Context) {
	userID, ok := httputil.MustUserID(c)
	if !ok {
		return
	}
	user, err := h.repo.FindByID(userID)
	if err != nil {
		httputil.Error(c, httputil.NotFound("Usuário não encontrado"))
		return
	}
	httputil.JSON(c, 200, gin.H{"linked": user.SpotifyAccessToken != ""})
}

// Search executa a busca usando o token do próprio usuário (não o
// token global do .env). Retorna 401 se não houver vínculo.
func (h *SpotifyHandler) Search(c *gin.Context) {
	userID, ok := httputil.MustUserID(c)
	if !ok {
		return
	}
	query := c.Query("q")
	if query == "" {
		httputil.Error(c, httputil.BadRequest("Query 'q' é obrigatória"))
		return
	}

	user, err := h.repo.FindByID(userID)
	if err != nil || user.SpotifyAccessToken == "" {
		httputil.Error(c, httputil.Unauthorized("Spotify não vinculado ou usuário não encontrado"))
		return
	}

	tracks, err := h.provider.SearchFromToken(c.Request.Context(), userToToken(user), query)
	if err != nil {
		httputil.Error(c, httputil.Internal("Erro ao buscar no Spotify"))
		return
	}
	httputil.JSON(c, 200, tracks)
}
