# 🚀 Quick Start - Music Backend

## ⚡ 30 segundos para começar

```bash
# 1. Setup
cp .env.example .env
# ⚠️  Editar .env com suas credenciais Spotify

# 2. Backend (Terminal 1)
make run

# 3. Cliente (Terminal 2)
make client
```

## 🎮 Controles Básicos

| Ação | Teclas |
|------|--------|
| Navegar | `↑↓` ou `j/k` |
| Confirmar | `Enter` |
| Voltar | `Esc` |
| Sair | `Ctrl+C` |
| Login | `Ctrl+L` |
| Register | `Ctrl+R` |

## 📝 Arquivos Importantes

- `.env.example` → Copie para `.env`
- `cmd/client/README.md` → Guia do cliente
- `SETUP.md` → Instalação detalhada
- `SPOTIFY_AUTH.md` → Arquitetura

## 🔑 Configuração Spotify

1. https://developer.spotify.com/dashboard
2. Create App
3. Copie `Client ID` e `Client Secret`
4. Cole em `.env`
5. Registre Redirect URI: `http://localhost:8080/api/v1/auth/spotify/callback`

## ✅ Checklist Inicial

- [ ] `.env.example` → `.env`
- [ ] Credenciais Spotify no `.env`
- [ ] `make run` (backend rodando)
- [ ] `make client` (cliente rodando)
- [ ] Registrar novo usuário
- [ ] Login
- [ ] Vincular Spotify
- [ ] Buscar música

## 🐛 Problemas Comuns

| Erro | Solução |
|------|---------|
| Connection refused | Backend não está rodando (`make run`) |
| Spotify não vinculado | Clique em "🔗 Link Spotify" |
| API error | Verifique credenciais Spotify |
| JWT error | Gere novo JWT_SECRET |

## 📚 Documentação Completa

```
.
├── README.md           ← Visão geral
├── SETUP.md            ← Instalação passo a passo
├── SPOTIFY_AUTH.md     ← Arquitetura em detalhe
├── MIGRATION_CHECKLIST.md ← Para migração
└── cmd/client/README.md ← Cliente TUI
```

## 🎵 Próximas Etapas

- Explorar Search
- Criar Playlists
- Adicionar Tracks
- Sincronizar com Spotify

---

**Pronto? Execute `make run` + `make client` e aproveite! 🚀**
