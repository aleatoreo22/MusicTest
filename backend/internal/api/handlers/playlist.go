package handlers

import (
	"music-unified-backend/internal/models"
	"music-unified-backend/internal/repository"
	"music-unified-backend/pkg/httputil"
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
	userID, ok := httputil.MustUserID(c)
	if !ok {
		return
	}
	var req models.CreatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, httputil.BadRequest(err.Error()))
		return
	}
	playlist := models.Playlist{Name: req.Name, UserID: userID}
	if err := h.playlistRepo.Create(&playlist); err != nil {
		httputil.Error(c, httputil.Internal("Erro ao criar playlist"))
		return
	}
	httputil.JSON(c, 201, playlist)
}

func (h *PlaylistHandler) List(c *gin.Context) {
	userID, ok := httputil.MustUserID(c)
	if !ok {
		return
	}
	playlists, err := h.playlistRepo.FindAllByUserID(userID)
	if err != nil {
		httputil.Error(c, httputil.Internal("Erro ao listar playlists"))
		return
	}
	httputil.JSON(c, 200, playlists)
}

func (h *PlaylistHandler) AddTrack(c *gin.Context) {
	userID, ok := httputil.MustUserID(c)
	if !ok {
		return
	}
	playlistID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		httputil.Error(c, httputil.BadRequest("ID inválido"))
		return
	}
	var req models.AddTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, httputil.BadRequest(err.Error()))
		return
	}
	if _, err := h.playlistRepo.FindByID(uint(playlistID), userID); err != nil {
		httputil.Error(c, httputil.NotFound("Playlist não encontrada ou acesso negado"))
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
		httputil.Error(c, httputil.Internal("Erro ao adicionar música"))
		return
	}
	httputil.JSON(c, 200, gin.H{"message": "Música adicionada com sucesso"})
}
