package spotify

import (
	"context"
	"music-unified-backend/internal/models"
	"os"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

type Provider struct {
	auth *spotifyauth.Authenticator
}

func NewProvider() *Provider {
	redirectURL := os.Getenv("SPOTIFY_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/api/v1/auth/spotify/callback"
	}

	return &Provider{
		auth: spotifyauth.New(
			spotifyauth.WithRedirectURL(redirectURL),
			spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate, spotifyauth.ScopeUserReadPlaybackState, spotifyauth.ScopeUserModifyPlaybackState),
			spotifyauth.WithClientID(os.Getenv("SPOTIFY_CLIENT_ID")),
			spotifyauth.WithClientSecret(os.Getenv("SPOTIFY_CLIENT_SECRET")),
		),
	}
}

func (p *Provider) Enabled() bool {
	id := os.Getenv("SPOTIFY_CLIENT_ID")
	secret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	return id != "" && secret != "" &&
		id != "seu_client_id_aqui" && secret != "seu_client_secret_aqui"
}

func (p *Provider) GetAuthURL(state string) string {
	test := p.auth.AuthURL(state)
	println(test)
	return test
}

func (p *Provider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.auth.Exchange(ctx, code)
}

func (p *Provider) GetClient(ctx context.Context, token *oauth2.Token) *spotify.Client {
	return spotify.New(p.auth.Client(ctx, token))
}

func (p *Provider) Search(ctx context.Context, client *spotify.Client, query string) ([]models.Track, error) {
	results, err := client.Search(ctx, query, spotify.SearchTypeTrack)
	if err != nil {
		return nil, err
	}

	var tracks []models.Track
	if results.Tracks != nil {
		for _, item := range results.Tracks.Tracks {
			durationSec := item.Duration / 1000

			artistName := "Unknown"
			if len(item.Artists) > 0 {
				artistName = item.Artists[0].Name
			}

			thumbnail := ""
			if len(item.Album.Images) > 0 {
				thumbnail = item.Album.Images[0].URL
			}

			tracks = append(tracks, models.Track{
				Title:      item.Name,
				Artist:     artistName,
				Duration:   int(durationSec),
				Thumbnail:  thumbnail,
				PreviewURL: item.PreviewURL,
				Source:     models.SourceSpotify,
				SourceID:   string(item.ID),
			})
		}
	}

	return tracks, nil
}
