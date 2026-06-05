# music_unified

Cliente Flutter para o [MusicUnified backend](../backend). Espelha o
cliente TUI escrito em Go (Bubble Tea) usando a mesma API REST.

## Plataformas suportadas

- **Android** (APK) — alvo principal
- **Linux desktop** — funciona, requer `libsecret-1` (ver deps)
- iOS / macOS / Windows / Web — não configurados ainda

## Dependências do sistema

### Linux (para `flutter build linux`)

```bash
sudo dnf install -y libsecret-devel   # Fedora
sudo apt install -y libsecret-1-dev   # Debian/Ubuntu
```

### Android (para `flutter build apk`)

1. Instalar Android Studio (recomendado) OU
2. Baixar [command line tools][cmdline] e exportar `ANDROID_HOME`
3. Aceitar licenças: `flutter doctor --android-licenses`

[cmdline]: https://developer.android.com/studio#command-line-tools-only

## Build

```bash
# Dependências
flutter pub get

# Análise estática
flutter analyze

# Testes
flutter test

# APK debug (Android)
flutter build apk --debug

# Linux (requer libsecret-1-dev)
flutter build linux --debug
```

## Configuração

A URL do backend é definida em `lib/main.dart` (`_defaultBaseUrl`):

| Ambiente | URL |
|---|---|
| Emulador Android (host = dev machine) | `http://10.0.2.2:8080` |
| Linux desktop | `http://127.0.0.1:8080` |
| Dispositivo físico Android | `http://<IP-LAN-DO-DEV>:8080` |

> O backend Go deve estar rodando (`cd backend && make run`).

## Estrutura

```
lib/
├── api/
│   ├── client.dart      # espelho Dart do client.go Go
│   ├── models.dart      # DTOs (User, Track, Playlist, ...)
│   └── auth_store.dart  # persistência do JWT (Keychain/Keystore)
├── screens/
│   ├── login_screen.dart
│   └── home_screen.dart
└── main.dart            # entrypoint + _AuthGate
```

## Status

- [x] Auth (login/registro, persistência JWT)
- [x] Listagem de providers habilitados
- [ ] Busca multi-provider
- [ ] Playlists (criar, listar, add track)
- [ ] Player (Spotify Connect + playback local)
- [ ] Vincular Spotify (OAuth)
- [ ] Background playback com MediaSession
