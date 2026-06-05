# 🎵 Solução: Token Spotify por Usuário

## 🐛 Problema Identificado
Anteriormente, o token do Spotify era armazenado em variáveis de ambiente (`.env`), sendo compartilhado globalmente para toda a API.

**Problemas com essa abordagem:**
- ❌ Todos os usuários compartilhavam o mesmo token Spotify
- ❌ Cada usuário via as mesmas playlists/favoritos
- ❌ Não era possível ter múltiplos usuários com contas Spotify diferentes

## ✅ Solução Implementada

### 1. **Arquitetura - Cada usuário tem seu próprio token**

```
User Model (backend)
├── ID
├── Email
├── Password
├── SpotifyAccessToken  ← Token por usuário
├── SpotifyRefreshToken ← Refresh token por usuário
└── SpotifyTokenExpiry  ← Expiração por usuário
```

### 2. **Novo Fluxo de Autenticação Spotify**

```
1. Usuário faz login com email/senha
   └─> Recebe JWT token da API

2. Na home, seleciona "🔗 Link Spotify"
   └─> Frontend solicita URL de autenticação Spotify
   └─> Backend retorna URL com userID como state

3. Usuário é redirecionado para Spotify
   └─> Autoriza acesso da app
   └─> Spotify redireciona para callback com code

4. Backend troca code por token Spotify
   └─> Salva accessToken, refreshToken e expiry NO USUÁRIO

5. Agora o usuário pode buscar músicas no Spotify
   └─> Backend usa token do usuário para cada requisição
```

### 3. **Novos Endpoints Adicionados**

#### GET `/api/v1/auth/spotify/status`
Verifica se usuário tem Spotify vinculado
```bash
curl -H "Authorization: Bearer {jwt_token}" \
  http://localhost:8080/api/v1/auth/spotify/status
# Response: { "linked": true }
```

#### GET `/api/v1/auth/spotify`
Retorna URL para autorização do usuário
```bash
curl -H "Authorization: Bearer {jwt_token}" \
  http://localhost:8080/api/v1/auth/spotify
# Response: { "url": "https://accounts.spotify.com/authorize?..." }
```

### 4. **Fluxo no Cliente TUI**

```
Home Menu
├─ 🔍 Search Spotify → Busca (requer Spotify vinculado)
├─ 📋 My Playlists → Gerencia playlists
├─ 🔗 Link Spotify → Vincula conta Spotify (NOVO)
└─ 🚪 Logout
```

## 🔧 Implementação Backend

### Modelo User Atualizado
```go
type User struct {
    ID                   uint
    Email                string
    Password             string
    SpotifyAccessToken   string    // ← Token do usuário
    SpotifyRefreshToken  string    // ← Refresh token
    SpotifyTokenExpiry   time.Time // ← Expiração
}
```

### Handler Spotify Atualizado
```go
// GetStatus - Verifica se Spotify está vinculado
func (h *SpotifyHandler) GetStatus(c *gin.Context) {
    userID := c.Get("userID")
    user := h.repo.FindByID(userID)
    isLinked := user.SpotifyAccessToken != ""
    c.JSON(http.StatusOK, gin.H{"linked": isLinked})
}

// Search - Usa token do usuário
func (h *SpotifyHandler) Search(c *gin.Context) {
    userID := c.Get("userID")
    user := h.repo.FindByID(userID)
    
    // Rejeita se não tem Spotify vinculado
    if user.SpotifyAccessToken == "" {
        return 401 "Spotify não vinculado"
    }
    
    // Usa token do usuário
    token := &oauth2.Token{
        AccessToken:  user.SpotifyAccessToken,
        RefreshToken: user.SpotifyRefreshToken,
        Expiry:       user.SpotifyTokenExpiry,
    }
    
    client := h.provider.GetClient(ctx, token)
    results := h.provider.Search(ctx, client, query)
}
```

## 🖥️ Implementação Frontend (TUI)

### Nova View: SpotifyLinkView
- Verifica status de vinculação
- Solicita URL de autenticação
- Exibe instruções ao usuário
- Permite vincular conta

### Menu Atualizado
```
🎵 Music Client - Home
├─ 🔍 Search Spotify
├─ 📋 My Playlists  
├─ 🔗 Link Spotify   ← NOVO
└─ 🚪 Logout
```

## 📋 Fluxo de Uso Completo

### Primeiro login e vinculação:
```
1. Executar: make client
2. Registrar: Ctrl+R → Email/Senha
3. Login: Ctrl+L → Email/Senha
4. Navegar: ↓ → "🔗 Link Spotify" → Enter
5. Autorizar: Visite a URL exibida no Spotify
6. Retornar: Pressione Enter no cliente
7. Pronto! Agora pode buscar: ↑ → "🔍 Search Spotify"
```

### Próximos logins:
```
1. Executar: make client
2. Login: Ctrl+L → Email/Senha
3. Usar: Já tem Spotify vinculado, busque à vontade!
```

## 🔒 Segurança

- ✅ Token Spotify armazenado no banco de dados
- ✅ Cada usuário tem seu próprio token
- ✅ Token não é exposto ao frontend
- ✅ Apenas backend usa token para requisições Spotify
- ✅ JWT token da API garante autenticação do usuário

## 📚 Arquivos Modificados

### Backend
1. `internal/api/handlers/spotify.go` - AddedGetStatus method
2. `internal/api/routes.go` - Added `/auth/spotify/status` route

### Frontend (TUI)
1. `cmd/client/views/spotify.go` - New SpotifyLinkView
2. `cmd/client/views/home.go` - Updated menu items
3. `cmd/client/models.go` - Added spotifyState
4. `cmd/client/api/client.go` - Added GetSpotifyStatus, GetSpotifyAuthURL

## 🚀 Próximas Melhorias

- [ ] Renovação automática de token quando expirar
- [ ] Logout que limpa token Spotify
- [ ] Suporte para múltiplas contas Spotify
- [ ] Sincronizar favoritos do Spotify com playlists locais
- [ ] Atualizar playlists do Spotify automaticamente
