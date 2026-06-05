package handlers

import (
	"music-unified-backend/internal/models"
	"music-unified-backend/internal/repository"
	"music-unified-backend/pkg/httputil"
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

// StreamTrack serve um arquivo de áudio local com suporte a Range
// requests (seeking) e Content-Type automático via c.File do Gin.
func (h *StreamHandler) StreamTrack(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		httputil.Error(c, httputil.BadRequest("ID inválido"))
		return
	}

	track, err := h.musicRepo.FindTrackByID(uint(id))
	if err != nil {
		httputil.Error(c, httputil.NotFound("Música não encontrada"))
		return
	}

	if track.Source != models.SourceLocal {
		httputil.Error(c, httputil.BadRequest("Esta música não é local. Streaming direto via API suportado apenas para arquivos locais."))
		return
	}

	if track.LocalPath == "" {
		httputil.Error(c, httputil.Internal("Caminho do arquivo não configurado"))
		return
	}

	if _, err := os.Stat(track.LocalPath); os.IsNotExist(err) {
		httputil.Error(c, httputil.NotFound("Arquivo de áudio não encontrado no disco"))
		return
	}

	c.File(track.LocalPath)
}
