package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	router, _, _, _, _ := setupTestRouter()

	t.Run("Success", func(t *testing.T) {
		body := map[string]string{
			"email":    "newuser@example.com",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusCreated, resp.Code)

		var data map[string]string
		json.Unmarshal(resp.Body.Bytes(), &data)
		assert.NotEmpty(t, data["token"], "register deve retornar um JWT")
	})

	t.Run("Duplicate Email", func(t *testing.T) {
		body := map[string]string{
			"email":    "duplicate@example.com",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(body)
		
		// Primeiro registro
		req1, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(httptest.NewRecorder(), req1)

		// Segundo registro com mesmo email
		req2, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req2)

		assert.Equal(t, http.StatusConflict, resp.Code)
	})
}

func TestLogin(t *testing.T) {
	router, _, _, _, _ := setupTestRouter()

	// Registrar usuário para o teste
	registerBody := map[string]string{
		"email":    "login@example.com",
		"password": "password123",
	}
	rb, _ := json.Marshal(registerBody)
	reqReg, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(rb))
	router.ServeHTTP(httptest.NewRecorder(), reqReg)

	t.Run("Success", func(t *testing.T) {
		loginBody := map[string]string{
			"email":    "login@example.com",
			"password": "password123",
		}
		lb, _ := json.Marshal(loginBody)
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(lb))
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		
		var data map[string]string
		json.Unmarshal(resp.Body.Bytes(), &data)
		assert.NotEmpty(t, data["token"])
	})

	t.Run("Wrong Password", func(t *testing.T) {
		loginBody := map[string]string{
			"email":    "login@example.com",
			"password": "wrongpassword",
		}
		lb, _ := json.Marshal(loginBody)
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(lb))
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})
}
