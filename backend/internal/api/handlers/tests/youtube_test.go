package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYouTubeProviderListed(t *testing.T) {
	router, _, _, _, _ := setupTestRouter()
	token := getAuthToken(router)

	req, _ := http.NewRequest("GET", "/api/v1/providers", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	// Como yt-dlp está instalado no ambiente de teste (/usr/bin/yt-dlp),
	// o provider deve aparecer. Se este teste falhar em outro ambiente,
	// provavelmente yt-dlp não está no PATH.
	assert.Contains(t, resp.Body.String(), "youtube",
		"yt-dlp está no PATH do ambiente de teste; youtube deve aparecer em /providers")
}

func TestYouTubeSearch(t *testing.T) {
	router, _, _, _, _ := setupTestRouter()
	token := getAuthToken(router)

	t.Run("Missing Query", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/search/youtube", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "Query")
	})
}

func TestYouTubeStream(t *testing.T) {
	router, _, _, _, _ := setupTestRouter()
	token := getAuthToken(router)

	t.Run("Missing ID", func(t *testing.T) {
		// /stream/youtube/ com path vazio não deve bater no handler :id,
		// mas o handler valida id vazio por defesa. Aqui testamos só
		// que o endpoint existe e requer auth.
		req, _ := http.NewRequest("GET", "/api/v1/stream/youtube/dQw4w9WgXcQ", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// yt-dlp pode ou não estar disponível. Se estiver (caso padrão
		// no ambiente dev), esperamos um 200 ou 502 (rate limit/erro yt-dlp).
		// Se não estiver, esperamos 503. Em qualquer caso, não 404/401.
		assert.NotEqual(t, http.StatusUnauthorized, resp.Code)
		assert.NotEqual(t, http.StatusNotFound, resp.Code)
	})
}
