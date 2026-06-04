package handlers

import (
	"fmt"
	"music-unified-backend/internal/providers/spotify"
	"music-unified-backend/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type SpotifyHandler struct {
	repo     repository.UserRepository
	provider *spotify.Provider
}

func NewSpotifyHandler(repo repository.UserRepository, provider *spotify.Provider) *SpotifyHandler {
	return &SpotifyHandler{
		repo:     repo,
		provider: provider,
	}
}

func (h *SpotifyHandler) Login(c *gin.Context) {
	userID, _ := c.Get("userID")
	// Usamos o ID do usuário como 'state' para simplificar o vínculo no callback
	state := fmt.Sprintf("%v", userID)
	url := h.provider.GetAuthURL(state)
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *SpotifyHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Código não fornecido"})
		return
	}

	token, err := h.provider.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao trocar código por token"})
		return
	}

	// Recuperar ID do usuário do state
	var userID uint
	fmt.Sscanf(state, "%d", &userID)

	user, err := h.repo.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	// Atualizar tokens do usuário
	user.SpotifyAccessToken = token.AccessToken
	user.SpotifyRefreshToken = token.RefreshToken
	user.SpotifyTokenExpiry = token.Expiry

	if err := h.repo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao salvar tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conta do Spotify vinculada com sucesso!"})
}

func (h *SpotifyHandler) GetStatus(c *gin.Context) {
	userID, _ := c.Get("userID")

	user, err := h.repo.FindByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	isLinked := user.SpotifyAccessToken != ""
	c.JSON(http.StatusOK, gin.H{
		"linked": isLinked,
	})
}

func (h *SpotifyHandler) Search(c *gin.Context) {
	userID, _ := c.Get("userID")
	query := c.Query("q")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query 'q' é obrigatória"})
		return
	}

	user, err := h.repo.FindByID(userID.(uint))
	if err != nil || user.SpotifyAccessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Spotify não vinculado ou usuário não encontrado"})
		return
	}

	token := &oauth2.Token{
		AccessToken:  user.SpotifyAccessToken,
		RefreshToken: user.SpotifyRefreshToken,
		Expiry:       user.SpotifyTokenExpiry,
	}

	client := h.provider.GetClient(c.Request.Context(), token)
	tracks, err := h.provider.Search(c.Request.Context(), client, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar no Spotify"})
		return
	}

	c.JSON(http.StatusOK, tracks)
}
