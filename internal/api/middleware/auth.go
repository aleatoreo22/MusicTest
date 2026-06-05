package middleware

import (
	"music-unified-backend/pkg/httputil"
	"music-unified-backend/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware verifica se o token JWT é válido e injeta o ID do
// usuário no contexto do request para uso pelos handlers.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			httputil.Error(c, httputil.Unauthorized("Cabeçalho de autorização é obrigatório"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
			httputil.Error(c, httputil.Unauthorized("Formato de autorização inválido. Use 'Bearer <token>'"))
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			httputil.Error(c, httputil.Unauthorized("Token inválido ou expirado"))
			c.Abort()
			return
		}

		c.Set(httputil.ContextKeyUserID, claims.UserID)
		c.Next()
	}
}
