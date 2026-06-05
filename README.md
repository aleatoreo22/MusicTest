# MusicUnified

Monorepo: backend Go (REST API) + cliente Flutter (Android/Linux).

```
.
├── backend/         # Go: API REST (Gin + GORM + SQLite)
│   ├── cmd/api/     # entrypoint do servidor
│   ├── cmd/client/  # cliente TUI legado (Bubble Tea) — coexistindo
│   ├── internal/    # handlers, models, providers, repositories, services
│   └── pkg/         # utils (auth, player, browser, ...)
├── mobile/          # Flutter: cliente Android/Linux
│   ├── lib/         # código Dart
│   ├── android/     # config nativa Android
│   └── linux/       # config nativa Linux
├── REFACTORING.md   # (em backend/) histórico do refactor do Go
└── .gitignore       # ignora artefatos de build (binários, db, .env)
```

## Quick start

### Backend

```bash
cd backend
cp .env.example .env   # editar com credenciais (JWT_SECRET, Spotify, etc)
make run               # servidor em :8080
```

### Mobile

```bash
cd mobile
flutter pub get
flutter run            # device/emulator conectado
```

Ver [`mobile/README.md`](mobile/README.md) para mais detalhes.

## Documentação

- [`backend/PROJECT_CONTEXT.md`](backend/PROJECT_CONTEXT.md) — arquitetura e overview do backend
- [`backend/REFACTORING.md`](backend/REFACTORING.md) — guia passo a passo do refactor aplicado
- [`backend/SPOTIFY_AUTH.md`](backend/SPOTIFY_AUTH.md) — design do OAuth Spotify por usuário
- [`mobile/README.md`](mobile/README.md) — guia do app Flutter
