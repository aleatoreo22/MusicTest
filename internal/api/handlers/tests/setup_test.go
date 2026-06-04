package handlers_test

import (
	"music-unified-backend/internal/api"
	"music-unified-backend/internal/models"
	"music-unified-backend/internal/providers/soundcloud"
	"music-unified-backend/internal/providers/spotify"
	"music-unified-backend/internal/providers/youtube"
	"music-unified-backend/internal/repository"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestRouter() (*gin.Engine, repository.UserRepository, repository.PlaylistRepository, repository.MusicRepository, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	db.AutoMigrate(&models.User{}, &models.Track{}, &models.Playlist{})

	userRepo := repository.NewGormUserRepository(db)
	playlistRepo := repository.NewGormPlaylistRepository(db)
	musicRepo := repository.NewGormMusicRepository(db)

	spotifyProv := spotify.NewProvider()
	soundcloudProv := soundcloud.NewProvider()
	youtubeProv := youtube.NewProvider()

	r := api.SetupRouter(userRepo, playlistRepo, musicRepo, spotifyProv, soundcloudProv, youtubeProv)

	os.Setenv("JWT_SECRET", "test-secret")

	return r, userRepo, playlistRepo, musicRepo, db
}
