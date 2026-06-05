# 🎵 Music Client - Terminal UI

Cliente TUI (Terminal User Interface) construído com **Bubble Tea** e **Lipgloss** para interagir com o backend de música unificada.

## 📋 Features

✨ **Autenticação**
- Login com email e senha
- Registro de novos usuários
- Gerenciamento de tokens JWT

🔍 **Busca Spotify**
- Buscar músicas em tempo real
- Visualizar resultados com artista e duração
- Navegação intuitiva

📚 **Gerenciamento de Playlists**
- Criar novas playlists
- Listar suas playlists
- Adicionar músicas às playlists
- Visualizar tracks em cada playlist

🎨 **Interface**
- TUI responsiva com Bubble Tea
- Cores destacadas com Lipgloss
- Navegação com arrow keys (↑↓) ou vim keys (j/k)

## 🚀 Instalação

### 1. Instalar dependências
```bash
cd MusicTest
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go mod tidy
```

### 2. Compilar o cliente
```bash
make client-build
# ou para rodar direto
make client
```

## 📖 Uso

### Terminal 1 - Iniciar o backend
```bash
make run
```

O backend rodará em `http://localhost:8080`

### Terminal 2 - Iniciar o cliente
```bash
make client
```

Ou se já compilou:
```bash
./music-client
```

## ⌨️ Controles

### Auth Screen
| Tecla | Ação |
|-------|------|
| `Tab` | Navegar entre campos |
| `Shift+Tab` | Navegar para trás |
| `Ctrl+L` | Mudar para Login |
| `Ctrl+R` | Mudar para Register |
| `Enter` | Enviar |
| `Ctrl+C` | Sair |

### Home Menu
| Tecla | Ação |
|-------|------|
| `↑` / `k` | Subir |
| `↓` / `j` | Descer |
| `Enter` | Selecionar |
| `Ctrl+C` | Sair |

### Search
| Tecla | Ação |
|-------|------|
| `Enter` | Buscar (quando digitando) |
| `↑` / `k` | Navegar resultados |
| `↓` / `j` | Navegar resultados |
| `Esc` | Voltar ao menu |

### Playlists
| Tecla | Ação |
|-------|------|
| `↑` / `k` | Navegar listas |
| `↓` / `j` | Navegar listas |
| `n` | Nova playlist |
| `Enter` | Criar (se criando) |
| `Esc` | Sair/Cancelar |

## 🏗️ Estrutura

```
cmd/client/
├── main.go              # Entry point
├── models.go            # Bubble Tea main model
├── api/
│   ├── models.go        # Data models
│   └── client.go        # HTTP client
└── views/
    ├── auth.go          # Login/Register
    ├── home.go          # Main menu
    ├── search.go        # Search interface
    └── playlists.go     # Playlists management
```

## 🔧 Configuração

### Mudar URL do backend
Edite `MusicTest/cmd/client/models.go`:
```go
apiClient := api.NewClient("http://localhost:8080") // Mudar URL aqui
```

## 🛠️ Desenvolvimento

### Build
```bash
make client-build
```

### Run (modo desenvolvimento)
```bash
make client
```

### Limpar
```bash
make clean
```

## 📝 Exemplos de Uso

### 1. Registrar novo usuário
1. Execute `make client`
2. Pressione `Ctrl+R` para ir para Register
3. Digite seu email e senha
4. Pressione `Tab` para navegar
5. Pressione `Enter` no botão [Register]

### 2. Buscar música
1. Faça login com suas credenciais
2. Navegue até "🔍 Search Spotify"
3. Digite o nome da música (ex: "Bohemian Rhapsody")
4. Pressione `Enter` para buscar
5. Use ↑↓ para navegar resultados

### 3. Criar e gerenciar playlists
1. Após login, vá para "📋 My Playlists"
2. Pressione `n` para criar nova playlist
3. Digite um nome
4. Pressione `Enter` para criar
5. Selecione uma playlist e adicione músicas

## 🐛 Troubleshooting

### "Connection refused"
- Verifique se o backend está rodando: `make run`
- Confirme a porta 8080 está correta

### "Erro de autenticação"
- Verifique se suas credenciais estão corretas
- Se for novo usuário, registre-se primeiro com `Ctrl+R`

### Interface quebrada
- Tente redimensionar seu terminal
- Verifique compatibilidade do terminal (use iTerm2, Terminal.app, ou similar em macOS)

## 📚 Dependências

- `github.com/charmbracelet/bubbletea` - Framework TUI
- `github.com/charmbracelet/lipgloss` - Styling

## 📄 Licença

Parte do projeto Music Unified Backend

## 💡 Próximas Melhorias

- [ ] Suporte para adicionar tracks direto da busca
- [ ] Playlist colaborativas
- [ ] Sincronização com Spotify/YouTube Music
- [ ] Cache local
- [ ] Suporte para Dark/Light theme
- [ ] Histórico de reprodução
