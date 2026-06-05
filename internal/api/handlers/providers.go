package handlers

import (
	"music-unified-backend/internal/providers"
	"music-unified-backend/internal/providers/spotify"
	"music-unified-backend/pkg/httputil"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProvidersHandler lista os providers habilitados. Aceita providers
// que implementam a interface comum (busca/stream sem contexto de
// usuário) e, opcionalmente, o Spotify separadamente — que precisa
// de OAuth por usuário e não se enquadra na interface.
type ProvidersHandler struct {
	providers []providers.Provider
	spotify   *spotify.Provider
}

func NewProvidersHandler(p []providers.Provider, sp *spotify.Provider) *ProvidersHandler {
	return &ProvidersHandler{providers: p, spotify: sp}
}

func (h *ProvidersHandler) List(c *gin.Context) {
	names := make([]string, 0, len(h.providers)+1)
	for _, p := range h.providers {
		if p.Enabled() {
			names = append(names, p.Name())
		}
	}
	if h.spotify != nil && h.spotify.Enabled() {
		names = append(names, h.spotify.Name())
	}
	httputil.JSON(c, http.StatusOK, gin.H{"providers": names})
}
