package handlers

import (
	"music-unified-backend/internal/providers/soundcloud"
	"music-unified-backend/internal/providers/spotify"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProvidersHandler struct {
	spotify    *spotify.Provider
	soundcloud *soundcloud.Provider
}

func NewProvidersHandler(sp *spotify.Provider, sc *soundcloud.Provider) *ProvidersHandler {
	return &ProvidersHandler{spotify: sp, soundcloud: sc}
}

func (h *ProvidersHandler) List(c *gin.Context) {
	providers := []string{}
	if h.spotify != nil && h.spotify.Enabled() {
		providers = append(providers, "spotify")
	}
	if h.soundcloud != nil && h.soundcloud.Enabled() {
		providers = append(providers, "soundcloud")
	}
	c.JSON(http.StatusOK, gin.H{
		"providers": providers,
	})
}
