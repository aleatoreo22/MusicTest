# 🚀 Setup Guia - Music Backend

## 📋 Pré-requisitos

- Go 1.26.4 ou superior
- Conta no [Spotify for Developers](https://developer.spotify.com/dashboard)
- Terminal/Shell

## 1️⃣ Clonar e Preparar

```bash
cd MusicTest

# Copiar arquivo de exemplo
cp .env.example .env

# Editar configurações (pelo menos o Spotify)
# nano .env  (ou editor de sua preferência)
```

## 2️⃣ Configurar Spotify OAuth

1. Acesse https://developer.spotify.com/dashboard
2. Faça login ou crie uma conta
3. Clique em "Create an App"
4. Aceite os termos e crie
5. Você receberá:
   - `Client ID`
   - `Client Secret`
6. Adicione a Redirect URI em "Edit Settings":
   ```
   http://localhost:8080/api/v1/auth/spotify/callback
   ```

## 3️⃣ Atualizar .env

Abra `MusicTest/.env` e preencha:

```env
SPOTIFY_CLIENT_ID=<seu_client_id>
SPOTIFY_CLIENT_SECRET=<seu_client_secret>
JWT_SECRET=<gere_com_openssl_rand_-base64_32>
PORT=8080
```

## 4️⃣ Instalar Dependências

```bash
go mod download
go mod tidy
```

## 5️⃣ Iniciar o Backend

```bash
# Opção 1: Com Make
make run

# Opção 2: Direto com Go
go run cmd/api/main.go
```

Você deve ver:
```
[GIN-debug] Listening and serving HTTP on :8080
```

## 6️⃣ Iniciar o Cliente (em outro terminal)

```bash
cd MusicTest

# Opção 1: Modo desenvolvimento
make client

# Opção 2: Binário compilado
make client-build
./music-client
```

## 📖 Fluxo de Uso

### Primeiro Acesso

1. **Registre uma conta**
   - Pressione `Ctrl+R`
   - Digite email e senha
   - Pressione `Enter`

2. **Faça Login**
   - Pressione `Ctrl+L`
   - Use as credenciais que registrou
   - Pressione `Enter`

3. **Vincule Spotify**
   - Navegue até "🔗 Link Spotify"
   - Pressione `Enter`
   - Visite a URL exibida
   - Autorize o acesso
   - Retorne e pressione `Enter`

4. **Busque Músicas**
   - Selecione "🔍 Search Spotify"
   - Digite o nome de uma música
   - Pressione `Enter` para buscar

### Próximos Acessos

Basta fazer login - seu Spotify já estará vinculado!

## 🛠️ Troubleshooting

### "Connection refused" no cliente

- Verifique se o backend está rodando: `make run`
- Confirme que está rodando em `http://localhost:8080`

### Erro ao buscar músicas

- Verifique se você vincou Spotify: "🔗 Link Spotify"
- Confirme que o Spotify foi autorizado

### Erro "JWT_SECRET não definido"

```bash
# Gere uma chave segura
openssl rand -base64 32

# Cole a saída no .env como JWT_SECRET
```

## 📚 Arquivos Importantes

- `.env` - Configurações (⚠️ Não commitar!)
- `.env.example` - Exemplo de configuração (✅ Commitar)
- `cmd/api/main.go` - Entry point do backend
- `cmd/client/` - Cliente TUI
- `internal/` - Lógica da aplicação

## ✅ Comandos Make Úteis

```bash
make run              # Inicia servidor
make client           # Inicia cliente
make build            # Compila backend
make client-build     # Compila cliente
make test             # Roda testes
make clean            # Remove build artifacts
make tidy             # Organiza dependências
```

## 🔒 Segurança em Produção

Nunca commitir:
- `.env` (contém secrets)
- `music.db` (dados do usuário)
- Tokens Spotify

Sempre use:
- Variáveis de ambiente para secrets
- HTTPS em produção
- JWT_SECRET forte
- Refresh tokens do Spotify

## 📝 Próximas Etapas

- [ ] Implementar renovação de token Spotify
- [ ] Sincronizar com playlists do Spotify
- [ ] Suporte para YouTube Music
- [ ] Suporte para SoundCloud
- [ ] Deploy em produção

---

Pronto! 🎵 Você deve estar com o sistema rodando. Aproveite!
