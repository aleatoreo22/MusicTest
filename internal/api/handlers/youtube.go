package handlers

import (
	"music-unified-backend/internal/providers/youtube"

	"github.com/gin-gonic/gin"
)

type YouTubeHandler struct {
	provider *youtube.Provider
}

func NewYouTubeHandler(provider *youtube.Provider) *YouTubeHandler {
	return &YouTubeHandler{provider: provider}
}

func (h *YouTubeHandler) Search(c *gin.Context) {
	handleProviderSearch(c, h.provider)
}

func (h *YouTubeHandler) Stream(c *gin.Context) {
	handleProviderStream(c, h.provider, "audio")
}
