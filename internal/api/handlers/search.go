package handlers

import (
	"music-unified-backend/internal/providers"
	"music-unified-backend/pkg/httputil"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// handleProviderSearch implementa o endpoint GET /search/<provider>
// para qualquer Provider. Lê q (obrigatório) e limit (opcional, default
// 20) do query string e mapeia o erro para o status HTTP apropriado.
func handleProviderSearch(c *gin.Context, p providers.Provider) {
	query := c.Query("q")
	if query == "" {
		httputil.Error(c, httputil.BadRequest("Query 'q' é obrigatória"))
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	tracks, err := p.Search(c.Request.Context(), query, limit)
	if err != nil {
		httputil.Error(c, providerErrorStatus(p, err))
		return
	}
	httputil.JSON(c, 200, tracks)
}

// handleProviderStream implementa o endpoint GET /stream/<provider>/:id.
// O parâmetro `format` identifica o tipo de mídia retornado
// (ex: "hls" para SoundCloud, "audio" para YouTube).
func handleProviderStream(c *gin.Context, p providers.Provider, format string) {
	id := c.Param("id")
	if id == "" {
		httputil.Error(c, httputil.BadRequest("id é obrigatório"))
		return
	}

	url, expires, err := p.Stream(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, providerErrorStatus(p, err))
		return
	}
	if expires.IsZero() {
		expires = time.Now().Add(4 * time.Hour)
	}
	httputil.JSON(c, 200, gin.H{
		"url":        url,
		"expires_at": expires.Unix(),
		"format":     format,
	})
}

// providerErrorStatus classifica erros de provider em status HTTP.
// Se o provider não está habilitado, retornamos 503 (configuração).
// Outros erros do provider viram 502 (gateway/upstream).
func providerErrorStatus(p providers.Provider, err error) *httputil.APIError {
	if !p.Enabled() {
		return httputil.Unavailable(err.Error())
	}
	return httputil.BadGateway(err.Error())
}
