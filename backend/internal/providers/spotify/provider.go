package spotify

import (
	"context"
	"music-unified-backend/internal/config"
	"music-unified-backend/internal/models"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

// Provider implementa a interface comum de providers para o Spotify.
// Encapsula o OAuth2 e a busca via SDK zmb3/spotify.
type Provider struct {
	auth *spotifyauth.Authenticator
	cfg  config.SpotifyConfig
}

const name = "spotify"

func NewProvider(cfg config.SpotifyConfig) *Provider {
	return &Provider{
		cfg: cfg,
		auth: spotifyauth.New(
			spotifyauth.WithRedirectURL(cfg.RedirectURL),
			spotifyauth.WithScopes(
				spotifyauth.ScopeUserReadPrivate,
				spotifyauth.ScopeUserReadPlaybackState,
				spotifyauth.ScopeUserModifyPlaybackState,
			),
			spotifyauth.WithClientID(cfg.ClientID),
			spotifyauth.WithClientSecret(cfg.ClientSecret),
		),
	}
}

func (p *Provider) Name() string { return name }

func (p *Provider) Enabled() bool { return p.cfg.IsConfigured() }

func (p *Provider) GetAuthURL(state string) string {
	return p.auth.AuthURL(state)
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
	if results.Tracks == nil {
		return nil, nil
	}

	tracks := make([]models.Track, 0, len(results.Tracks.Tracks))
	for _, item := range results.Tracks.Tracks {
		artist := "Unknown"
		if len(item.Artists) > 0 {
			artist = item.Artists[0].Name
		}
		thumbnail := ""
		if len(item.Album.Images) > 0 {
			thumbnail = item.Album.Images[0].URL
		}
		tracks = append(tracks, models.Track{
			Title:      item.Name,
			Artist:     artist,
			Duration:   int(item.Duration / 1000),
			Thumbnail:  thumbnail,
			PreviewURL: item.PreviewURL,
			Source:     models.SourceSpotify,
			SourceID:   string(item.ID),
		})
	}
	return tracks, nil
}

// SearchFromToken monta o client a partir do token do usuário e
// executa a busca. Usado pelo handler quando ele já tem o token
// persistido (em vez de receber o client pronto).
func (p *Provider) SearchFromToken(ctx context.Context, token *oauth2.Token, query string) ([]models.Track, error) {
	return p.Search(ctx, p.GetClient(ctx, token), query)
}
