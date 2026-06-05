package main

import (
	"log"
	"music-unified-backend/internal/api"
	"music-unified-backend/internal/config"
	"music-unified-backend/internal/providers/soundcloud"
	"music-unified-backend/internal/providers/spotify"
	"music-unified-backend/internal/providers/youtube"
	"music-unified-backend/internal/repository"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config inválida: %v", err)
	}

	db, err := repository.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("erro ao conectar ao banco: %v", err)
	}

	spotifyProv := spotify.NewProvider(cfg.Spotify)
	soundcloudProv := soundcloud.NewProvider(cfg.SoundCloud)
	youtubeProv := youtube.NewProvider(cfg.YouTube)

	userRepo := repository.NewGormUserRepository(db)
	playlistRepo := repository.NewGormPlaylistRepository(db)
	musicRepo := repository.NewGormMusicRepository(db)

	router := api.SetupRouter(userRepo, playlistRepo, musicRepo, spotifyProv, soundcloudProv, youtubeProv)

	log.Printf("servidor iniciando na porta %s (env=%s)...", cfg.Port, cfg.Env)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("erro ao iniciar o servidor: %v", err)
	}
}
