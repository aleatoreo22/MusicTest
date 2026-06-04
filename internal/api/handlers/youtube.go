package handlers

import (
	"net/http"
	"strconv"

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
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query 'q' é obrigatória"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	tracks, err := h.provider.Search(c.Request.Context(), query, limit)
	if err != nil {
		if err == youtube.ErrNotAvailable {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tracks)
}

func (h *YouTubeHandler) Stream(c *gin.Context) {
	videoID := c.Param("id")
	if videoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id é obrigatório"})
		return
	}

	url, expires, err := h.provider.Stream(c.Request.Context(), videoID)
	if err != nil {
		if err == youtube.ErrNotAvailable {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"url":        url,
		"expires_at": expires.Unix(),
		"format":     "audio",
	})
}
