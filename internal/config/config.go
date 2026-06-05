// Package config centraliza o carregamento e validação de variáveis
// de ambiente. Substitui os múltiplos os.Getenv() espalhados por
// providers e handlers, expondo uma struct imutável e validada.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config agrupa toda a configuração de runtime carregada na inicialização.
type Config struct {
	Env      string
	LogLevel string
	Port     string

	DatabaseURL string
	JWTSecret   string

	Spotify    SpotifyConfig
	SoundCloud SoundCloudConfig
	YouTube    YouTubeConfig
}

type SpotifyConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func (s SpotifyConfig) IsConfigured() bool {
	return s.ClientID != "" && s.ClientSecret != "" &&
		!isPlaceholder(s.ClientID) && !isPlaceholder(s.ClientSecret)
}

type SoundCloudConfig struct {
	ClientID     string
	ClientSecret string
}

func (s SoundCloudConfig) IsConfigured() bool {
	return s.ClientID != "" && s.ClientSecret != "" &&
		!isPlaceholder(s.ClientID)
}

type YouTubeConfig struct {
	CookiesFile string
}

// Load lê as variáveis de ambiente e retorna uma Config validada.
// Falha com erro descritivo se algum campo obrigatório estiver ausente.
func Load() (*Config, error) {
	cfg := &Config{
		Env:         getenv("ENV", "development"),
		LogLevel:    getenv("LOG_LEVEL", "info"),
		Port:        getenv("PORT", "8080"),
		DatabaseURL: getenv("DATABASE_URL", "sqlite:music.db"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		Spotify: SpotifyConfig{
			ClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
			ClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
			RedirectURL:  getenv("SPOTIFY_REDIRECT_URL", "http://localhost:8080/api/v1/auth/spotify/callback"),
		},
		SoundCloud: SoundCloudConfig{
			ClientID:     os.Getenv("SOUNDCLOUD_CLIENT_ID"),
			ClientSecret: os.Getenv("SOUNDCLOUD_CLIENT_SECRET"),
		},
		YouTube: YouTubeConfig{
			CookiesFile: os.Getenv("YOUTUBE_COOKIES_FILE"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate verifica invariantes obrigatórias. Em produção, JWT_SECRET
// é mandatório; em dev/teste aceitamos o fallback explícito via flag.
func (c *Config) validate() error {
	if c.JWTSecret == "" {
		if c.Env == "production" {
			return fmt.Errorf("config: JWT_SECRET é obrigatório em produção")
		}
		c.JWTSecret = "dev-only-insecure-secret-do-not-use-in-prod"
	}
	if len(c.JWTSecret) < 16 {
		return fmt.Errorf("config: JWT_SECRET deve ter ao menos 16 caracteres (recebido %d)", len(c.JWTSecret))
	}
	if _, err := strconv.Atoi(c.Port); err != nil {
		return fmt.Errorf("config: PORT inválida %q", c.Port)
	}
	return nil
}

// IsProduction reporta se estamos em ambiente de produção.
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// isPlaceholder detecta os valores "sua_X_aqui" deixados no .env.example.
// Centraliza a heurística que estava duplicada em spotify.Enabled() e
// soundcloud.Enabled().
func isPlaceholder(s string) bool {
	return strings.HasPrefix(s, "seu_") || strings.HasPrefix(s, "sua_") || strings.HasSuffix(s, "_aqui")
}
