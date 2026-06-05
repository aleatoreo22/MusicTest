package handlers

import (
	"music-unified-backend/internal/providers/soundcloud"

	"github.com/gin-gonic/gin"
)

type SoundCloudHandler struct {
	provider *soundcloud.Provider
}

func NewSoundCloudHandler(provider *soundcloud.Provider) *SoundCloudHandler {
	return &SoundCloudHandler{provider: provider}
}

func (h *SoundCloudHandler) Search(c *gin.Context) {
	handleProviderSearch(c, h.provider)
}

func (h *SoundCloudHandler) Stream(c *gin.Context) {
	handleProviderStream(c, h.provider, "hls")
}
