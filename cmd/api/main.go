package main

import (
	"log"
	"music-unified-backend/internal/api"
	"music-unified-backend/internal/providers/soundcloud"
	"music-unified-backend/internal/providers/spotify"
	"music-unified-backend/internal/providers/youtube"
	"music-unified-backend/internal/repository"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	// 1. Inicializa o Banco de Dados
	db, err := repository.InitDB()
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}

	// 2. Inicializa os Provedores
	spotifyProv := spotify.NewProvider()
	soundcloudProv := soundcloud.NewProvider()
	youtubeProv := youtube.NewProvider()

	// 3. Inicializa os Repositórios
	userRepo := repository.NewGormUserRepository(db)
	playlistRepo := repository.NewGormPlaylistRepository(db)
	musicRepo := repository.NewGormMusicRepository(db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 4. Inicializa o Roteador passando as dependências
	router := api.SetupRouter(userRepo, playlistRepo, musicRepo, spotifyProv, soundcloudProv, youtubeProv)

	log.Printf("Servidor iniciando na porta %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
