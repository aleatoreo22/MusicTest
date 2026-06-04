package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getAuthToken(router http.Handler) string {
	// Registro
	regBody := map[string]string{"email": "tester@example.com", "password": "password123"}
	rb, _ := json.Marshal(regBody)
	reqReg, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(rb))
	router.ServeHTTP(httptest.NewRecorder(), reqReg)

	// Login
	reqLog, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(rb))
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, reqLog)

	var data map[string]string
	json.Unmarshal(resp.Body.Bytes(), &data)
	return data["token"]
}

func TestPlaylistOperations(t *testing.T) {
	router, _, _, _, _ := setupTestRouter()
	token := getAuthToken(router)

	t.Run("Create Playlist", func(t *testing.T) {
		body := map[string]string{"name": "Test Playlist"}
		jb, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/v1/playlists", bytes.NewBuffer(jb))
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusCreated, resp.Code)
	})

	t.Run("Add Track and List", func(t *testing.T) {
		// Criar playlist primeiro
		body := map[string]string{"name": "Music List"}
		jb, _ := json.Marshal(body)
		reqP, _ := http.NewRequest("POST", "/api/v1/playlists", bytes.NewBuffer(jb))
		reqP.Header.Set("Authorization", "Bearer "+token)
		respP := httptest.NewRecorder()
		router.ServeHTTP(respP, reqP)
		
		var playlist map[string]interface{}
		json.Unmarshal(respP.Body.Bytes(), &playlist)
		playlistID := int(playlist["id"].(float64))

		// Adicionar música local
		trackBody := map[string]interface{}{
			"title": "Test Local",
			"source": "local",
			"local_path": "/tmp/test.mp3",
		}
		tb, _ := json.Marshal(trackBody)
		reqT, _ := http.NewRequest("POST", "/api/v1/playlists/1/tracks", bytes.NewBuffer(tb))
		reqT.Header.Set("Authorization", "Bearer "+token)
		respT := httptest.NewRecorder()
		router.ServeHTTP(respT, reqT)

		assert.Equal(t, http.StatusOK, respT.Code)

		// Listar e verificar
		reqL, _ := http.NewRequest("GET", "/api/v1/playlists", nil)
		reqL.Header.Set("Authorization", "Bearer "+token)
		respL := httptest.NewRecorder()
		router.ServeHTTP(respL, reqL)

		assert.Equal(t, http.StatusOK, respL.Code)
		assert.Contains(t, respL.Body.String(), "Test Local")
		_ = playlistID
	})
}
