package handlers

import (
	"music-unified-backend/internal/providers/soundcloud"
	"music-unified-backend/internal/providers/spotify"
	"music-unified-backend/internal/providers/youtube"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProvidersHandler struct {
	spotify    *spotify.Provider
	soundcloud *soundcloud.Provider
	youtube    *youtube.Provider
}

func NewProvidersHandler(sp *spotify.Provider, sc *soundcloud.Provider, yt *youtube.Provider) *ProvidersHandler {
	return &ProvidersHandler{spotify: sp, soundcloud: sc, youtube: yt}
}

func (h *ProvidersHandler) List(c *gin.Context) {
	providers := []string{}
	if h.spotify != nil && h.spotify.Enabled() {
		providers = append(providers, "spotify")
	}
	if h.soundcloud != nil && h.soundcloud.Enabled() {
		providers = append(providers, "soundcloud")
	}
	if h.youtube != nil && h.youtube.Enabled() {
		providers = append(providers, "youtube")
	}
	c.JSON(http.StatusOK, gin.H{
		"providers": providers,
	})
}
