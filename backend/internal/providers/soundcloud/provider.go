package soundcloud

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"music-unified-backend/internal/config"
	"music-unified-backend/internal/models"
)

const (
	tokenURL  = "https://secure.soundcloud.com/oauth/token"
	apiBase   = "https://api.soundcloud.com"
	defaultUA = "Mozilla/5.0 (compatible; MusicUnifiedBackend/1.0)"
	name      = "soundcloud"
)

// Provider implementa a interface comum de providers para o SoundCloud.
// Usa o fluxo client_credentials para obter um token de aplicação e
// busca/resolve URLs HLS para streaming off-platform.
type Provider struct {
	cfg  config.SoundCloudConfig
	http *http.Client

	mu          sync.Mutex
	accessToken string
	tokenExpiry time.Time
}

var (
	ErrNotConfigured = errors.New("soundcloud: SOUNDCLOUD_CLIENT_ID/SECRET não configurados")
	ErrBlocked       = errors.New("soundcloud: faixa bloqueada/indisponível para streaming off-platform")
)

func NewProvider(cfg config.SoundCloudConfig) *Provider {
	return &Provider{
		cfg:  cfg,
		http: &http.Client{Timeout: 15 * time.Second},
	}
}

func (p *Provider) Name() string { return name }

func (p *Provider) Enabled() bool { return p.cfg.IsConfigured() }

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

func (p *Provider) token(ctx context.Context) (string, error) {
	if !p.Enabled() {
		return "", ErrNotConfigured
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.accessToken != "" && time.Until(p.tokenExpiry) > 30*time.Second {
		return p.accessToken, nil
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	auth := p.cfg.ClientID + ":" + p.cfg.ClientSecret
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", defaultUA)

	resp, err := p.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("soundcloud token: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("soundcloud token status %d: %s", resp.StatusCode, string(body))
	}
	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", err
	}
	if tr.AccessToken == "" {
		return "", errors.New("soundcloud: token vazio na resposta")
	}
	p.accessToken = tr.AccessToken
	if tr.ExpiresIn > 0 {
		p.tokenExpiry = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	} else {
		p.tokenExpiry = time.Now().Add(50 * time.Minute)
	}
	return p.accessToken, nil
}

func (p *Provider) authed(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
	tok, err := p.token(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, apiBase+endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "OAuth "+tok)
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", defaultUA)
	return p.http.Do(req)
}

type scTrack struct {
	ID         int64  `json:"id"`
	URN        string `json:"urn"`
	Title      string `json:"title"`
	Duration   int    `json:"duration"` // ms
	ArtworkURL string `json:"artwork_url"`
	User       struct {
		Username string `json:"username"`
	} `json:"user"`
	Streamable   bool   `json:"streamable"`
	Policy       string `json:"policy"`
	Access       string `json:"access"`
	EmbeddableBy string `json:"embeddable_by"`
}

type scSearchResponse struct {
	Collection []scTrack `json:"collection"`
}

// Search implementa a interface providers.Provider. limit fora do
// intervalo [1,50] é normalizado para 20.
func (p *Provider) Search(ctx context.Context, query string, limit int) ([]models.Track, error) {
	if !p.Enabled() {
		return nil, ErrNotConfigured
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	q := url.Values{}
	q.Set("q", query)
	q.Set("limit", strconv.Itoa(limit))
	endpoint := "/search/tracks?" + q.Encode()

	resp, err := p.authed(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("soundcloud search status %d: %s", resp.StatusCode, string(body))
	}
	var sr scSearchResponse
	if err := json.Unmarshal(body, &sr); err != nil {
		return nil, err
	}
	out := make([]models.Track, 0, len(sr.Collection))
	for _, t := range sr.Collection {
		if t.Policy == "BLOCK" || t.Access == "blocked" || !t.Streamable {
			continue
		}
		thumbnail := strings.Replace(t.ArtworkURL, "-large", "-t500x500", 1)
		out = append(out, models.Track{
			Title:     t.Title,
			Artist:    t.User.Username,
			Duration:  t.Duration / 1000,
			Thumbnail: thumbnail,
			Source:    models.SourceSoundCloud,
			SourceID:  strconv.FormatInt(t.ID, 10),
		})
	}
	return out, nil
}

type scStreamsResponse struct {
	HlsAAC160  string `json:"hls_aac_160_url"`
	HlsAAC96   string `json:"hls_aac_96_url"`
	HlsMP3     string `json:"hls_mp3_128_url"`
	PreviewMP3 string `json:"preview_mp3_128_url"`
}

func (p *Provider) Streams(ctx context.Context, trackID string) (*scStreamsResponse, error) {
	if !p.Enabled() {
		return nil, ErrNotConfigured
	}
	resp, err := p.authed(ctx, http.MethodGet, "/tracks/soundcloud:tracks:"+trackID+"/streams", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 404 {
		return nil, ErrBlocked
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("soundcloud streams status %d: %s", resp.StatusCode, string(body))
	}
	var sr scStreamsResponse
	if err := json.Unmarshal(body, &sr); err != nil {
		return nil, err
	}
	return &sr, nil
}

// Stream resolve a URL final reproduzível (HLS) para uma faixa. Faz
// o segundo passo autenticado para extrair o Location do 302.
func (p *Provider) Stream(ctx context.Context, trackID string) (string, time.Time, error) {
	streams, err := p.Streams(ctx, trackID)
	if err != nil {
		return "", time.Time{}, err
	}
	candidate := firstNonEmpty(streams.HlsAAC160, streams.HlsAAC96, streams.HlsMP3)
	if candidate == "" {
		return "", time.Time{}, errors.New("soundcloud: nenhum stream disponível para esta faixa")
	}
	tok, err := p.token(ctx)
	if err != nil {
		return "", time.Time{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, candidate, nil)
	if err != nil {
		return "", time.Time{}, err
	}
	req.Header.Set("Authorization", "OAuth "+tok)
	req.Header.Set("User-Agent", defaultUA)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()
	if loc := resp.Header.Get("Location"); loc != "" {
		expires := time.Now().Add(5 * time.Minute)
		if t, err := http.ParseTime(resp.Header.Get("Expires")); err == nil {
			expires = t
		}
		return loc, expires, nil
	}
	return "", time.Time{}, fmt.Errorf("soundcloud: sem header Location na resposta (status %d)", resp.StatusCode)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
