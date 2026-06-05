// Package service contém a lógica de negócio do backend, isolada
// dos handlers HTTP. Cada service recebe suas dependências via
// construtor (DI) e expõe métodos que retornam erros tipados
// (*httputil.APIError) que os handlers convertem em respostas.
package service

import (
	"music-unified-backend/internal/models"
	"music-unified-backend/internal/repository"
	"music-unified-backend/pkg/httputil"
	"music-unified-backend/pkg/utils"
)

// AuthService encapsula registro, login e leitura de perfil.
// Todos os métodos retornam *APIError pronto para o handler
// responder, isolando-o das regras de negócio.
type AuthService struct {
	users repository.UserRepository
}

func NewAuthService(users repository.UserRepository) *AuthService {
	return &AuthService{users: users}
}

// RegisterResult contém o user criado e o token JWT gerado.
type RegisterResult struct {
	User  *models.User
	Token string
}

func (s *AuthService) Register(req models.RegisterRequest) (*RegisterResult, *httputil.APIError) {
	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, httputil.Internal("Erro ao processar senha")
	}
	user := &models.User{Email: req.Email, Password: hashed}
	if err := s.users.Create(user); err != nil {
		return nil, httputil.Conflict("Usuário já existe ou erro no banco")
	}
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		return nil, httputil.Internal("Erro ao gerar token")
	}
	return &RegisterResult{User: user, Token: token}, nil
}

// LoginResult contém apenas o token (não devolvemos o user no login).
type LoginResult struct {
	Token string
}

func (s *AuthService) Login(req models.LoginRequest) (*LoginResult, *httputil.APIError) {
	user, err := s.users.FindByEmail(req.Email)
	if err != nil || !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, httputil.Unauthorized("Credenciais inválidas")
	}
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		return nil, httputil.Internal("Erro ao gerar token")
	}
	return &LoginResult{Token: token}, nil
}

func (s *AuthService) GetProfile(userID uint) (*models.User, *httputil.APIError) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, httputil.NotFound("Usuário não encontrado")
	}
	return user, nil
}
