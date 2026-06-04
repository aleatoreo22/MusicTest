package handlers

import (
	"music-unified-backend/internal/models"
	"music-unified-backend/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PlaylistHandler struct {
	playlistRepo repository.PlaylistRepository
}

func NewPlaylistHandler(repo repository.PlaylistRepository) *PlaylistHandler {
	return &PlaylistHandler{playlistRepo: repo}
}

func (h *PlaylistHandler) Create(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req models.CreatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	playlist := models.Playlist{
		Name:   req.Name,
		UserID: userID.(uint),
	}

	if err := h.playlistRepo.Create(&playlist); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar playlist"})
		return
	}

	c.JSON(http.StatusCreated, playlist)
}

func (h *PlaylistHandler) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	playlists, err := h.playlistRepo.FindAllByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar playlists"})
		return
	}

	c.JSON(http.StatusOK, playlists)
}

func (h *PlaylistHandler) AddTrack(c *gin.Context) {
	userID, _ := c.Get("userID")
	playlistIDStr := c.Param("id")
	playlistID, _ := strconv.ParseUint(playlistIDStr, 10, 32)

	var req models.AddTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verifica se a playlist pertence ao usuário
	_, err := h.playlistRepo.FindByID(uint(playlistID), userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist não encontrada ou acesso negado"})
		return
	}

	track := models.Track{
		Title:     req.Title,
		Artist:    req.Artist,
		Duration:  req.Duration,
		Thumbnail: req.Thumbnail,
		Source:    req.Source,
		SourceID:  req.SourceID,
		LocalPath: req.LocalPath,
	}

	if err := h.playlistRepo.AddTrack(uint(playlistID), &track); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao adicionar música"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Música adicionada com sucesso"})
}
