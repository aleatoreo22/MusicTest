package handlers

import (
	"errors"
	"music-unified-backend/internal/providers/spotify"
	"music-unified-backend/internal/repository"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	spotifyapi "github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

type PlayerHandler struct {
	userRepo repository.UserRepository
	provider *spotify.Provider
}

func NewPlayerHandler(repo repository.UserRepository, provider *spotify.Provider) *PlayerHandler {
	return &PlayerHandler{userRepo: repo, provider: provider}
}

type PlayRequest struct {
	URI      string `json:"uri,omitempty"`
	DeviceID string `json:"device_id,omitempty"`
}

func (h *PlayerHandler) client(c *gin.Context) (*spotifyapi.Client, error) {
	userID, _ := c.Get("userID")
	user, err := h.userRepo.FindByID(userID.(uint))
	if err != nil {
		return nil, err
	}
	if user.SpotifyAccessToken == "" {
		return nil, errors.New("spotify não vinculado")
	}
	token := &oauth2.Token{
		AccessToken:  user.SpotifyAccessToken,
		RefreshToken: user.SpotifyRefreshToken,
		Expiry:       user.SpotifyTokenExpiry,
		TokenType:    "Bearer",
	}
	return h.provider.GetClient(c.Request.Context(), token), nil
}

func (h *PlayerHandler) Play(c *gin.Context) {
	var req PlayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cli, err := h.client(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	opt := &spotifyapi.PlayOptions{}
	if req.URI != "" {
		opt.URIs = []spotifyapi.URI{spotifyapi.URI(req.URI)}
	}
	if req.DeviceID != "" {
		id := spotifyapi.ID(req.DeviceID)
		opt.DeviceID = &id
	}

	if err := cli.PlayOpt(c.Request.Context(), opt); err != nil {
		status, msg := classifySpotifyErr(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "uri": req.URI})
}

func (h *PlayerHandler) Pause(c *gin.Context) {
	cli, err := h.client(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if err := cli.Pause(c.Request.Context()); err != nil {
		status, msg := classifySpotifyErr(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *PlayerHandler) Next(c *gin.Context) {
	cli, err := h.client(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if err := cli.Next(c.Request.Context()); err != nil {
		status, msg := classifySpotifyErr(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *PlayerHandler) Previous(c *gin.Context) {
	cli, err := h.client(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if err := cli.Previous(c.Request.Context()); err != nil {
		status, msg := classifySpotifyErr(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *PlayerHandler) Devices(c *gin.Context) {
	cli, err := h.client(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	devs, err := cli.PlayerDevices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, devs)
}

func (h *PlayerHandler) CurrentlyPlaying(c *gin.Context) {
	cli, err := h.client(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	cp, err := cli.PlayerCurrentlyPlaying(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if cp == nil {
		c.JSON(http.StatusOK, gin.H{"is_playing": false})
		return
	}
	c.JSON(http.StatusOK, cp)
}

func classifySpotifyErr(err error) (int, string) {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "404"), strings.Contains(msg, "No active device"), strings.Contains(msg, "device not found"):
		return http.StatusNotFound, "Nenhum device do Spotify ativo. Abra o Spotify em algum device (celular, desktop ou web player) e tente novamente."
	case strings.Contains(msg, "403"), strings.Contains(msg, "Premium required"):
		return http.StatusForbidden, "Spotify Premium é necessário para controlar a reprodução remotamente."
	case strings.Contains(msg, "401"):
		return http.StatusUnauthorized, "Token do Spotify expirou. Revincule sua conta."
	}
	return http.StatusInternalServerError, msg
}
