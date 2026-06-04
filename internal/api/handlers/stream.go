package handlers

import (
	"music-unified-backend/internal/models"
	"music-unified-backend/internal/repository"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

type StreamHandler struct {
	musicRepo repository.MusicRepository
}

func NewStreamHandler(repo repository.MusicRepository) *StreamHandler {
	return &StreamHandler{musicRepo: repo}
}

func (h *StreamHandler) StreamTrack(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	track, err := h.musicRepo.FindTrackByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Música não encontrada"})
		return
	}

	if track.Source != models.SourceLocal {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Esta música não é local. Streaming direto via API suportado apenas para arquivos locais."})
		return
	}

	if track.LocalPath == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Caminho do arquivo não configurado"})
		return
	}

	// Verifica se o arquivo existe no disco
	if _, err := os.Stat(track.LocalPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Arquivo de áudio não encontrado no disco", "path": track.LocalPath})
		return
	}

	// http.ServeFile lida automaticamente com Range requests (seeking) e Content-Type
	// O navegador consegue baixar pedaços do arquivo conforme necessário
	c.File(track.LocalPath)
}
