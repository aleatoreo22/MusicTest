package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpotifyAuth(t *testing.T) {
	router, _, _, _, _ := setupTestRouter()
	token := getAuthToken(router)

	t.Run("Get Auth URL", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/auth/spotify", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), "https://accounts.spotify.com/authorize")
	})
}

func TestSpotifySearchUnauthorized(t *testing.T) {
	router, _, _, _, _ := setupTestRouter()
	token := getAuthToken(router)

	t.Run("Search Without Link", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/search/spotify?q=test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		// Deve retornar 401 porque o usuário no banco de testes não tem token do Spotify
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Contains(t, resp.Body.String(), "Spotify não vinculado")
	})
}
