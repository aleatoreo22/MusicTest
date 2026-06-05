# Project Overview

**MusicTest** (codename `music-unified-backend`) is a unified music backend in Go that integrates **Spotify**, **SoundCloud**, and **YouTube Music** into a single API. It ships with a **Terminal UI (TUI)** client built with Bubble Tea.

Provides search, streaming, playlist management, and Spotify Connect playback control — all behind JWT authentication, with per-user Spotify OAuth2 token isolation.

---

# Tech Stack

| Category | Technology | Version |
|---|---|---|
| Language | Go | 1.26.4 |
| Web Framework | Gin (gin-gonic/gin) | v1.12.0 |
| ORM | GORM (gorm.io/gorm) | v1.31.1 |
| Database | SQLite (gorm.io/driver/sqlite + mattn/go-sqlite3) | v1.6.0 / v1.14.44 |
| Auth (JWT) | golang-jwt/jwt/v5 | v5.3.1 |
| Auth (OAuth2) | golang.org/x/oauth2 | v0.36.0 |
| Spotify SDK | zmb3/spotify/v2 | v2.4.3 |
| TUI Framework | charmbracelet/bubbletea | v1.3.10 |
| TUI Styling | charmbracelet/lipgloss | v1.1.0 |
| Env Loading | joho/godotenv | v1.5.1 |
| JSON | bytedance/sonic | v1.15.1 |
| Testing | stretchr/testify | v1.11.1 |
| Password Hashing | golang.org/x/crypto (bcrypt) | v0.52.0 |
| Terminal Clipboard | aymanbagabas/go-osc52/v2 | v2.0.1 |

**External runtime dependencies (not Go modules):**
- `yt-dlp` (YouTube search/stream via subprocess)
- `mpv`, `ffplay`, `vlc`, or `cvlc` (local audio playback)

---

# Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    TUI Client (cmd/client)                   │
│  Bubble Tea + Lipgloss                                       │
│  ├── AuthView (login/register form)                          │
│  ├── HomeView (menu + provider listing)                      │
│  ├── SearchView (multi-source search + playback)             │
│  ├── PlaylistView (CRUD + playback)                          │
│  ├── SpotifyLinkView (OAuth linking)                         │
│  └── API Client (HTTP JSON to backend)                       │
└─────────────────────────┬───────────────────────────────────┘
                          │ HTTP REST (JSON)
┌─────────────────────────▼───────────────────────────────────┐
│                   Backend Server (cmd/api)                    │
│  Gin Web Framework + GORM + SQLite                           │
│                                                              │
│  Handlers Layer  ──►  Providers Layer  ──► External APIs     │
│  (Gin handlers)       (Spotify/SoundCloud/YT)                 │
│       │                                                      │
│       └──► Repository Layer  ──► GORM ──► SQLite DB          │
│                                                              │
│  Middleware: JWT Bearer Token validation                      │
└─────────────────────────────────────────────────────────────┘
```

**Key architectural decisions:**
- **Repository pattern**: Interfaces defined in `internal/repository/`, GORM implementations as unexported structs
- **Handler dependency injection**: Handlers receive repo/provider interfaces via constructor, no global state
- **Per-user OAuth isolation**: Each `User` record stores its own `SpotifyAccessToken`/`RefreshToken`/`Expiry`
- **Provider pattern**: Each music source implements `Enabled()` check + search/stream methods, self-contained
- **TUI state machine**: Client uses integer state enum (`authState`, `homeState`, `searchState`, etc.) with per-state views
- **No service layer**: Business logic lives directly in handlers. `internal/service/` directory exists but is unused

---

# Directory Structure

```
.
├── cmd/
│   ├── api/
│   │   └── main.go                    # Backend entry point
│   └── client/
│       ├── main.go                    # TUI client entry point
│       ├── models.go                  # Main Bubble Tea model (state machine)
│       ├── .env.example               # Client environment template
│       ├── api/
│       │   ├── client.go             # HTTP API client for backend
│       │   └── models.go             # Client-side data models
│       └── views/
│           ├── auth.go               # Login/Register form
│           ├── home.go               # Main menu with provider listing
│           ├── search.go             # Multi-source search + playback
│           ├── playlists.go          # Playlist CRUD + playback
│           ├── spotify.go            # Spotify OAuth linking
│           ├── playback.go           # Shared playback command helpers
│           └── styles.go             # Lipgloss style definitions
├── internal/
│   ├── api/
│   │   ├── routes.go                # Route definitions + DI wiring
│   │   ├── handlers/
│   │   │   ├── auth.go              # Register, Login, GetProfile
│   │   │   ├── spotify.go           # OAuth flow + search
│   │   │   ├── soundcloud.go        # Search + stream resolution
│   │   │   ├── youtube.go           # Search + stream resolution
│   │   │   ├── player.go            # Spotify Connect control
│   │   │   ├── playlist.go          # Playlist CRUD
│   │   │   ├── providers.go         # Provider listing
│   │   │   ├── stream.go            # Local file streaming
│   │   │   └── tests/               # Integration tests
│   │   │       ├── setup_test.go
│   │   │       ├── auth_test.go
│   │   │       ├── playlist_test.go
│   │   │       ├── spotify_test.go
│   │   │       └── youtube_test.go
│   │   └── middleware/
│   │       └── auth.go              # JWT Bearer auth middleware
│   ├── models/
│   │   ├── user.go                  # User, LoginRequest, RegisterRequest
│   │   └── music.go                 # Track, Playlist, request DTOs
│   ├── providers/
│   │   ├── spotify/
│   │   │   └── provider.go          # Spotify OAuth2 + search
│   │   ├── soundcloud/
│   │   │   └── provider.go          # SoundCloud client_credentials + search/stream
│   │   └── youtube/
│   │       └── provider.go          # YouTube via yt-dlp subprocess
│   ├── repository/
│   │   ├── db.go                    # DB init + auto-migrate
│   │   ├── user_repository.go       # UserRepository interface
│   │   ├── gorm_user_repository.go  # GORM UserRepository impl
│   │   ├── music_repository.go      # MusicRepository interface + impl
│   │   └── playlist_repository.go   # PlaylistRepository interface + impl
│   └── service/                     # (empty, reserved)
├── pkg/
│   └── utils/
│       ├── auth.go                  # JWT + bcrypt utilities
│       ├── browser.go               # Cross-platform browser opener
│       ├── hyperlink.go             # Terminal OSC-8 hyperlinks
│       └── player.go                # Local audio player (mpv/ffplay/vlc)
├── configs/                         # (empty, reserved)
├── .env                             # Local environment config
├── .env.example                     # Environment template
├── Makefile                         # Build automation
├── go.mod / go.sum                  # Go module dependencies
└── *.md                             # Documentation files
```

---

# Development Commands

| Command | Description |
|---|---|
| `make run` | Start backend server on port 8080 |
| `make build` | Compile backend binary to `music-api` |
| `make client` | Run TUI client |
| `make client-build` | Compile TUI client to `music-client` |
| `make test` | Run all tests |
| `make test-v` | Run tests verbosely |
| `make cover` | Show test coverage |
| `make tidy` | Clean go.mod dependencies |
| `make clean` | Remove binaries + `music.db` + `test.db` |
| `make help` | List all targets |

**Manual start:**
```bash
cp .env.example .env   # edit with credentials
make run               # terminal 1: backend
make client            # terminal 2: TUI client
```

---

# Testing

**Framework:** `stretchr/testify` (assertions) + standard `testing` package

**Test location:** `internal/api/handlers/tests/`

**Test pattern:**
- Package `handlers_test` (external test package)
- `setup_test.go` creates an in-memory SQLite DB, auto-migrates models, instantiates repos/providers, and calls `api.SetupRouter(...)` with `gin.TestMode`
- Each test file creates a fresh router via `setupTestRouter()`
- Tests are **integration tests** hitting HTTP endpoints via `httptest.NewRecorder()`
- Auth token helper: `getAuthToken(router)` registers + logs in a test user

**Coverage areas:**
- `auth_test.go` — register (success + duplicate), login (success + wrong password)
- `playlist_test.go` — create playlist, add track, list
- `spotify_test.go` — get auth URL, search without linked account (401)
- `youtube_test.go` — provider listed, missing query (400), missing stream ID (400)

**Run:** `make test` or `make test-v`

**NOTE:** Tests that call real external APIs (Spotify, SoundCloud, YouTube) may fail without credentials. The test setup uses real provider constructors which check env vars.

---

# Deployment

**No Dockerfile or CI/CD configuration exists.** Deployment is manual:

1. Build: `make build` → `music-api` binary
2. Set `ENV=production` in `.env`
3. Copy binary + `.env` to server
4. Run binary (consider systemd/supervisor for process management)
5. Ensure `yt-dlp` is installed on the server if YouTube provider is needed

The project assumes SQLite (file-based), so no external database server is needed.

---

# Coding Conventions

### General
- **Language**: English for code (IDs, JSON fields, comments), Portuguese for user-facing strings (error messages, TUI labels, docs)
- **Comments**: Code comments are in Portuguese
- **No comments in production code**: Per current codebase style, minimal comments; most are doc comments on exported types

### Naming
- **Packages**: lowercase, single word (`handlers`, `models`, `repository`, `providers`, `utils`)
- **Interfaces**: `XxxRepository` (e.g., `UserRepository`, `PlaylistRepository`)
- **Implementations**: `gormXxxRepository` (unexported), constructor `NewGormXxxRepository`
- **Handlers**: `XxxHandler` struct with private fields, constructor `NewXxxHandler`
- **Providers**: `XxxProvider` → but package is `spotify`, `soundcloud`, `youtube` with type `Provider`
- **Models**: Request DTOs suffixed with `Request` (`LoginRequest`, `RegisterRequest`, `AddTrackRequest`)
- **Constants**: CamelCase with package prefix for sources (`SourceSpotify`, `SourceSoundCloud`)

### Project structure conventions
- Backend entry in `cmd/api/main.go`, client entry in `cmd/client/main.go`
- Internal code in `internal/` (handlers, models, providers, repository)
- Shared utilities in `pkg/utils/`
- Interface + implementation pattern for repositories (interface in one file, GORM impl in separate file)
- Constructor functions always return the interface type, not the concrete struct

### Error handling
- Errors returned to user as JSON `{"error": "message"}` (Portuguese messages)
- Specific sentinel errors for provider states (`ErrNotConfigured`, `ErrBlocked`, `ErrNotAvailable`)
- HTTP status codes mapped to error semantics (409 for duplicate, 503 for unavailable, 502 for upstream errors)

### TUI pattern
- State machine with integer constants for view states
- Each view is a struct with `Update(tea.Msg, ...) (*View, tea.Cmd)` and `View() string`
- Views receive the API client as parameter to `Update()`
- Back navigation via `Back bool` field
- Playback commands in shared `playback.go` file

### Testing conventions
- External test package (`handlers_test`)
- `setupTestRouter()` returns all dependencies for test flexibility
- Sub-tests with `t.Run()`
- `httptest.NewRecorder()` for HTTP integration tests

---

# Important Files

| File | Purpose |
|---|---|
| `cmd/api/main.go` | Backend entry point — loads env, init DB, creates providers/repos, starts Gin |
| `internal/api/routes.go` | Route wiring — creates all handlers, configures public/protected routes |
| `internal/api/middleware/auth.go` | JWT Bearer token validation middleware |
| `internal/repository/db.go` | SQLite init + GORM auto-migrate (`User`, `Track`, `Playlist`) |
| `internal/models/user.go` | User model with Spotify token fields (hidden from JSON) |
| `internal/models/music.go` | Track and Playlist models, TrackSource constants, request DTOs |
| `pkg/utils/auth.go` | JWT generate/validate (HS256, 24h), bcrypt hash/compare (cost 14) |
| `pkg/utils/player.go` | Local audio player — mpv/ffplay/vlc subprocess management |
| `internal/providers/spotify/provider.go` | Spotify OAuth2 authenticator + search via zmb3 SDK |
| `internal/providers/soundcloud/provider.go` | SoundCloud client_credentials OAuth + search/stream (HLS) |
| `internal/providers/youtube/provider.go` | YouTube via yt-dlp subprocess (search + stream), 5min cache |
| `.env` | Runtime configuration (DB, JWT secret, OAuth credentials) |
| `.env.example` | Configuration template |
| `Makefile` | All build/run/test commands |
| `SPOTIFY_AUTH.md` | Architecture doc for per-user Spotify OAuth design |

### Routes Summary

**Public (no auth):**
- `GET /health` — health check
- `POST /api/v1/auth/register` — create account
- `POST /api/v1/auth/login` — authenticate
- `GET /api/v1/auth/spotify/callback` — Spotify OAuth callback

**Protected (JWT Bearer token):**
- `GET /api/v1/providers` — list enabled providers
- `GET /api/v1/me/profile` — current user profile
- `GET /api/v1/auth/spotify` — get Spotify auth URL
- `GET /api/v1/auth/spotify/status` — check Spotify linking
- `GET /api/v1/search/spotify?q=` — search Spotify
- `GET /api/v1/search/soundcloud?q=&limit=` — search SoundCloud
- `GET /api/v1/search/youtube?q=&limit=` — search YouTube
- `GET /api/v1/stream/soundcloud/:id` — resolve SoundCloud stream URL
- `GET /api/v1/stream/youtube/:id` — resolve YouTube stream URL
- `POST/GET /api/v1/player/*` — Spotify Connect control
- `POST/GET /api/v1/playlists/*` — playlist CRUD
- `GET /api/v1/tracks/:id/stream` — local file streaming (with Range support)

---

# Notes For Future AI Sessions

### Sensitive dependencies
- **JWT_SECRET** in `.env` — if lost/changed, all existing tokens become invalid. Default fallback `"super-secret-key-fallback"` exists for dev only.
- **Spotify credentials** — hardcoded in current `.env` (CLIENT_ID/SECRET). For production, use environment variables or a secrets manager.
- **SQLite** — no connection pooling needed, but file path `music.db` is hardcoded in `db.go`. The `DATABASE_URL` env var exists in `.env.example` but is NOT read by `InitDB()`. If changing to PostgreSQL, update `InitDB()` to read `DATABASE_URL`.

### Known limitations & gotchas
- **SoundCloud stream URLs expire** (~5 min). Client must re-resolve before playing.
- **YouTube stream URLs expire** (~4 h estimated). Same concern.
- **YouTube provider** requires `yt-dlp` installed on the system, not vendored.
- **Spotify player control** requires Spotify Premium (API restriction, enforced by Spotify).
- **No token refresh** for SoundCloud (client_credentials grants are short-lived, libraries auto-renew via `token()` method).
- **Spotify refresh tokens** are stored but NOT automatically refreshed in the player handler. If a token expires mid-session, the user gets a 401 and must re-link.
- **`internal/service/` is empty** — all business logic lives in handlers. Consider migrating to a service layer if logic grows.
- **TUI client hardcodes `http://localhost:8080`** as API base URL. The `cmd/client/.env.example` has `API_URL` but it's not read by the client code.
- **DB path hardcoded** in `repository/db.go` as `"music.db"` regardless of `DATABASE_URL` env var.
- **No HTTPS support** — add TLS config for production.
- **No rate limiting**, **no request logging** beyond Gin default, **no CORS configuration** (Gin default allows all).

### Extending the project
- **Adding a new provider**: Create `internal/providers/<name>/provider.go` with `Enabled()`, `Search()`, `Stream()` methods following existing patterns. Add handler in `internal/api/handlers/`, wire route in `routes.go`, register in `providers.go`.
- **Adding a new entity**: Add GORM model in `internal/models/`, create repository interface+impl in `internal/repository/`, add auto-migrate in `db.go`.
- **Adding a new handler**: Follow existing pattern — struct with private fields, constructor, methods taking `*gin.Context`.

### Configuration reference (`.env`)
```
DATABASE_URL=sqlite:music.db       # Currently unused, db.go uses hardcoded path
PORT=8080
JWT_SECRET=...                     # 32-byte base64 recommended
SPOTIFY_CLIENT_ID=...
SPOTIFY_CLIENT_SECRET=...
SPOTIFY_REDIRECT_URL=...
SOUNDCLOUD_CLIENT_ID=...
SOUNDCLOUD_CLIENT_SECRET=...
LOG_LEVEL=info                     # debug, info, warn, error
ENV=development                    # development, staging, production
YOUTUBE_COOKIES_FILE=              # Optional: path to YouTube cookies.txt
```

### Unused / deprecated artifacts
- `configs/` — empty directory, reserved but unused
- `internal/service/` — empty, reserved but unused
- `SourceYTMusic` constant exists but no provider implements it
- `PreviewURL` field on `Track` model is populated (from Spotify) but unused by the client
