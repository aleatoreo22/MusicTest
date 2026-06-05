# Unified Music Backend (Go)

Este é o backend de uma aplicação unificada de música que integra **Spotify**, **YouTube Music** e **SoundCloud**, além de suporte a **arquivos locais**.

## 🚀 Tecnologias
- **Linguagem:** Go (Golang)
- **Framework Web:** Gin Gonic
- **Banco de Dados:** SQLite (com suporte para PostgreSQL via GORM)
- **Autenticação:** JWT (JSON Web Token)
- **Integração:** Spotify OAuth2 & Search API

## 🛠️ Configuração do Spotify
Para usar o Spotify, você deve criar um aplicativo no [Spotify for Developers](https://developer.spotify.com/dashboard) e configurar:
1. **Redirect URI:** `http://localhost:8080/api/v1/auth/spotify/callback`
2. **Variáveis no .env:**
   ```text
   SPOTIFY_CLIENT_ID=seu_id
   SPOTIFY_CLIENT_SECRET=sua_secret
   SPOTIFY_REDIRECT_URL=http://localhost:8080/api/v1/auth/spotify/callback
   ```

**Nota:** Cada usuário vincula sua própria conta Spotify através do cliente. Tokens são armazenados por usuário.

## 📖 API Reference

### Público
- `GET /health`: Status do sistema.
- `POST /api/v1/auth/register`: Cadastro de usuário.
- `POST /api/v1/auth/login`: Login e obtenção de Token JWT.
- `GET /api/v1/auth/spotify/callback`: Endpoint de retorno do Spotify.

### Protegido (Requer Header `Authorization: Bearer <token>`)

#### Usuário & Auth Spotify
- `GET /api/v1/me/profile`: Perfil do usuário.
- `GET /api/v1/auth/spotify`: URL para vincular Spotify.
- `GET /api/v1/auth/spotify/status`: Status do Spotify.
- `GET /api/v1/auth/spotify/callback`: Callback Spotify.

#### Busca
- `GET /api/v1/search/spotify?q=musica`: Busca músicas no catálogo do Spotify.

#### Playlists
- `POST /api/v1/playlists`: Criar playlist (`{"name": "..."}`).
- `GET /api/v1/playlists`: Listar minhas playlists.
- `POST /api/v1/playlists/:id/tracks`: Adicionar música (local ou streaming).

## 🛠️ Comandos (Makefile)
- `make run`: Inicia o servidor.
- `make test`: Roda os testes.
- `make tidy`: Organiza dependências.
- `make client`: Inicia o cliente TUI.
- `make client-build`: Compila o cliente em binário.

## 🎯 Como Usar

### Terminal 1 - Backend
```bash
make run
```

### Terminal 2 - Cliente TUI
```bash
make client
```

Para mais detalhes sobre o cliente, veja `cmd/client/README.md`.
