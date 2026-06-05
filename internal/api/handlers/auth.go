package handlers

import (
	"music-unified-backend/internal/models"
	"music-unified-backend/internal/service"
	"music-unified-backend/pkg/httputil"

	"github.com/gin-gonic/gin"
)

// AuthHandler é um wrapper fino sobre AuthService. Sua única
// responsabilidade é parsear HTTP, delegar ao service, e
// formatar a resposta.
type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, httputil.BadRequest(err.Error()))
		return
	}
	res, apiErr := h.auth.Register(req)
	if apiErr != nil {
		httputil.Error(c, apiErr)
		return
	}
	httputil.JSON(c, 201, gin.H{
		"message": "Usuário criado com sucesso",
		"token":   res.Token,
		"type":    "Bearer",
		"user_id": res.User.ID,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, httputil.BadRequest(err.Error()))
		return
	}
	res, apiErr := h.auth.Login(req)
	if apiErr != nil {
		httputil.Error(c, apiErr)
		return
	}
	httputil.JSON(c, 200, gin.H{"token": res.Token, "type": "Bearer"})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, ok := httputil.MustUserID(c)
	if !ok {
		return
	}
	user, apiErr := h.auth.GetProfile(userID)
	if apiErr != nil {
		httputil.Error(c, apiErr)
		return
	}
	httputil.JSON(c, 200, user)
}
