package api

import (
	"music-unified-backend/internal/api/handlers"
	"music-unified-backend/internal/api/middleware"
	"music-unified-backend/internal/providers"
	"music-unified-backend/internal/providers/soundcloud"
	"music-unified-backend/internal/providers/spotify"
	"music-unified-backend/internal/providers/youtube"
	"music-unified-backend/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Providers agrupa os providers que implementam a interface comum
// (busca/stream sem contexto de usuário). Spotify é tratado separado
// por precisar de OAuth por usuário.
type Providers struct {
	Searchable []providers.Provider // SoundCloud, YouTube
	Spotify    *spotify.Provider
}

// SetupRouter configura todas as rotas da API. Aceita dependências
// via parâmetros explícitos (DI por construtor), sem estado global.
func SetupRouter(
	userRepo repository.UserRepository,
	playlistRepo repository.PlaylistRepository,
	musicRepo repository.MusicRepository,
	spotifyProv *spotify.Provider,
	soundcloudProv *soundcloud.Provider,
	youtubeProv *youtube.Provider,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Handlers
	authHandler := handlers.NewAuthHandler(userRepo)
	playlistHandler := handlers.NewPlaylistHandler(playlistRepo)
	spotifyHandler := handlers.NewSpotifyHandler(userRepo, spotifyProv)
	streamHandler := handlers.NewStreamHandler(musicRepo)
	playerHandler := handlers.NewPlayerHandler(userRepo, spotifyProv)

	searchable := []providers.Provider{soundcloudProv, youtubeProv}
	providersHandler := handlers.NewProvidersHandler(searchable, spotifyProv)
	soundcloudHandler := handlers.NewSoundCloudHandler(soundcloudProv)
	youtubeHandler := handlers.NewYouTubeHandler(youtubeProv)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/spotify/callback", spotifyHandler.Callback)
		}

		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/providers", providersHandler.List)
			protected.GET("/me/profile", authHandler.GetProfile)
			protected.GET("/auth/spotify", spotifyHandler.Login)
			protected.GET("/auth/spotify/status", spotifyHandler.GetStatus)
			protected.GET("/search/spotify", spotifyHandler.Search)
			protected.GET("/search/soundcloud", soundcloudHandler.Search)
			protected.GET("/search/youtube", youtubeHandler.Search)
			protected.GET("/stream/soundcloud/:id", soundcloudHandler.Stream)
			protected.GET("/stream/youtube/:id", youtubeHandler.Stream)

			player := protected.Group("/player")
			{
				player.POST("/play", playerHandler.Play)
				player.POST("/pause", playerHandler.Pause)
				player.POST("/next", playerHandler.Next)
				player.POST("/previous", playerHandler.Previous)
				player.POST("/seek", playerHandler.Seek)
				player.GET("/devices", playerHandler.Devices)
				player.GET("/currently-playing", playerHandler.CurrentlyPlaying)
			}

			playlists := protected.Group("/playlists")
			{
				playlists.POST("", playlistHandler.Create)
				playlists.GET("", playlistHandler.List)
				playlists.POST("/:id/tracks", playlistHandler.AddTrack)
			}

			protected.GET("/tracks/:id/stream", streamHandler.StreamTrack)
		}
	}

	return r
}
