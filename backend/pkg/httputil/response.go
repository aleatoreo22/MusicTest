// Package httputil fornece helpers compartilhados entre handlers HTTP
// para padronizar respostas, extração de contexto e tipagem de erros.
package httputil

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ContextKeyUserID é a chave usada pelo AuthMiddleware para injetar
// o ID do usuário autenticado no contexto do Gin.
const ContextKeyUserID = "userID"

// APIError representa um erro de domínio que carrega consigo o status
// HTTP apropriado e uma mensagem segura para o cliente. Use-o em
// services e handlers para evitar o "stringly-typed" gin.H{"error": "..."}.
type APIError struct {
	Status  int
	Message string
	Cause   error
}

func (e *APIError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *APIError) Unwrap() error { return e.Cause }

// Construtores ergonômicos para os status mais comuns.
func BadRequest(msg string) *APIError   { return &APIError{Status: http.StatusBadRequest, Message: msg} }
func Unauthorized(msg string) *APIError { return &APIError{Status: http.StatusUnauthorized, Message: msg} }
func Forbidden(msg string) *APIError    { return &APIError{Status: http.StatusForbidden, Message: msg} }
func NotFound(msg string) *APIError     { return &APIError{Status: http.StatusNotFound, Message: msg} }
func Conflict(msg string) *APIError     { return &APIError{Status: http.StatusConflict, Message: msg} }
func Internal(msg string) *APIError     { return &APIError{Status: http.StatusInternalServerError, Message: msg} }
func BadGateway(msg string) *APIError   { return &APIError{Status: http.StatusBadGateway, Message: msg} }
func Unavailable(msg string) *APIError  { return &APIError{Status: http.StatusServiceUnavailable, Message: msg} }

// Error responde ao cliente com status + mensagem. Aceita tanto *APIError
// quanto errors genéricos (esses viram 500 com a mensagem do erro).
func Error(c *gin.Context, err error) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		c.JSON(apiErr.Status, gin.H{"error": apiErr.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

// JSON responde com o status e payload fornecido.
func JSON(c *gin.Context, status int, payload any) {
	c.JSON(status, payload)
}

// UserIDFromCtx extrai o ID do usuário injetado pelo AuthMiddleware.
// Retorna (0, false) se ausente — nunca faz type assertion insegura.
func UserIDFromCtx(c *gin.Context) (uint, bool) {
	v, ok := c.Get(ContextKeyUserID)
	if !ok {
		return 0, false
	}
	id, ok := v.(uint)
	return id, ok
}

// MustUserID extrai o ID ou aborta com 401. Use em handlers protegidos
// onde a presença do userID é invariante (middleware já validou).
func MustUserID(c *gin.Context) (uint, bool) {
	id, ok := UserIDFromCtx(c)
	if !ok {
		Error(c, Unauthorized("usuário não autenticado"))
		return 0, false
	}
	return id, true
}
