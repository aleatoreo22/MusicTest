package handlers

import (
	"net/http"
	"strconv"

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
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query 'q' é obrigatória"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	tracks, err := h.provider.Search(c.Request.Context(), query, limit)
	if err != nil {
		if err == soundcloud.ErrNotConfigured {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tracks)
}

func (h *SoundCloudHandler) Stream(c *gin.Context) {
	trackID := c.Param("id")
	if trackID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id é obrigatório"})
		return
	}

	url, expires, err := h.provider.Stream(c.Request.Context(), trackID)
	if err != nil {
		if err == soundcloud.ErrNotConfigured {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		if err == soundcloud.ErrBlocked {
			c.JSON(http.StatusForbidden, gin.H{"error": "Faixa bloqueada/indisponível no SoundCloud"})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"url":        url,
		"expires_at": expires.Unix(),
		"format":     "hls",
	})
}
