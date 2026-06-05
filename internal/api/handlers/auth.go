package handlers

import (
	"music-unified-backend/internal/models"
	"music-unified-backend/internal/repository"
	"music-unified-backend/pkg/httputil"
	"music-unified-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userRepo repository.UserRepository
}

func NewAuthHandler(repo repository.UserRepository) *AuthHandler {
	return &AuthHandler{userRepo: repo}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, httputil.BadRequest(err.Error()))
		return
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		httputil.Error(c, httputil.Internal("Erro ao processar senha"))
		return
	}

	user := &models.User{Email: req.Email, Password: hashed}
	if err := h.userRepo.Create(user); err != nil {
		httputil.Error(c, httputil.Conflict("Usuário já existe ou erro no banco"))
		return
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		httputil.Error(c, httputil.Internal("Erro ao gerar token"))
		return
	}

	httputil.JSON(c, 201, gin.H{
		"message": "Usuário criado com sucesso",
		"token":   token,
		"type":    "Bearer",
		"user_id": user.ID,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, httputil.BadRequest(err.Error()))
		return
	}

	user, err := h.userRepo.FindByEmail(req.Email)
	if err != nil || !utils.CheckPasswordHash(req.Password, user.Password) {
		httputil.Error(c, httputil.Unauthorized("Credenciais inválidas"))
		return
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		httputil.Error(c, httputil.Internal("Erro ao gerar token"))
		return
	}

	httputil.JSON(c, 200, gin.H{"token": token, "type": "Bearer"})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, ok := httputil.MustUserID(c)
	if !ok {
		return
	}

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		httputil.Error(c, httputil.NotFound("Usuário não encontrado"))
		return
	}

	httputil.JSON(c, 200, user)
}
