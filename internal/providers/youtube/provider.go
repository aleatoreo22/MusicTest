package youtube

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"music-unified-backend/internal/models"
)

const (
	// YtSearchPrefix é usado pelo yt-dlp pra diferenciar "busca" de "url direta".
	// Ex.: ytsearch20:never gonna give you up → resolve pros 20 primeiros resultados.
	YtSearchPrefix = "ytsearch"
)

// Provider resolve search e stream de YouTube/YouTube Music via yt-dlp subprocess.
// Não usa API oficial (não existe pra áudio); yt-dlp é a abordagem padrão de fato.
type Provider struct {
	ytDlpPath  string
	cookiesFile string
	format     string

	mu    sync.Mutex
	cache *searchCache
}

type searchCacheEntry struct {
	tracks    []models.Track
	expiresAt time.Time
}

type searchCache struct {
	data map[string]searchCacheEntry
	ttl  time.Duration
}

// NewProvider detecta yt-dlp no PATH. Falha silenciosa se não estiver —
// Enabled() retornará false e o provider não aparece na listagem.
func NewProvider() *Provider {
	path, _ := exec.LookPath("yt-dlp")
	return &Provider{
		ytDlpPath:   path,
		cookiesFile: strings.TrimSpace(os.Getenv("YOUTUBE_COOKIES_FILE")),
		format:      "bestaudio[ext=m4a]/bestaudio/best",
		cache: &searchCache{
			data: make(map[string]searchCacheEntry),
			ttl:  5 * time.Minute,
		},
	}
}

// Enabled retorna true se yt-dlp está disponível. Não exige credenciais —
// YouTube é acessível sem login pra conteúdo público.
func (p *Provider) Enabled() bool {
	return p.ytDlpPath != ""
}

// Name identifica o provider (usado em logs, na listagem /providers, e como Source nos tracks).
func (p *Provider) Name() string {
	return "youtube"
}

// Search faz uma busca textual via yt-dlp e retorna até `limit` faixas.
// Usa --flat-playlist (rápido, sem extrair info de cada vídeo) e --dump-json (JSON Lines).
// Cacheia o resultado em memória por 5min pra evitar bater no YouTube em sequência.
func (p *Provider) Search(ctx context.Context, query string, limit int) ([]models.Track, error) {
	if !p.Enabled() {
		return nil, ErrNotAvailable
	}
	if query == "" {
		return nil, errors.New("youtube: query vazia")
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	cacheKey := fmt.Sprintf("%s|%d", strings.ToLower(strings.TrimSpace(query)), limit)
	if cached, ok := p.cacheGet(cacheKey); ok {
		return cached, nil
	}

	searchArg := YtSearchPrefix + strconv.Itoa(limit) + ":" + query
	args := []string{
		searchArg,
		"--flat-playlist",
		"--dump-json",
		"--skip-download",
		"--no-warnings",
		"--no-playlist",
	}
	if p.cookiesFile != "" {
		args = append(args, "--cookies", p.cookiesFile)
	}

	stdout, err := p.run(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("youtube search: %w", err)
	}

	tracks := parseSearchOutput(stdout)
	if len(tracks) == 0 {
		return nil, nil
	}
	p.cachePut(cacheKey, tracks)
	return tracks, nil
}

// Stream resolve a URL direta de áudio pra um vídeo do YouTube. A URL expira
// em algumas horas (server-side) — retornamos um expires_at conservador de 4h.
func (p *Provider) Stream(ctx context.Context, videoID string) (string, time.Time, error) {
	if !p.Enabled() {
		return "", time.Time{}, ErrNotAvailable
	}
	if videoID == "" {
		return "", time.Time{}, errors.New("youtube: videoID vazio")
	}

	watchURL := "https://www.youtube.com/watch?v=" + videoID
	args := []string{
		watchURL,
		"-f", p.format,
		"-g",
		"--no-warnings",
		"--no-playlist",
	}
	if p.cookiesFile != "" {
		args = append(args, "--cookies", p.cookiesFile)
	}

	stdout, err := p.run(ctx, args)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("youtube stream: %w", err)
	}

	// `-g` pode retornar múltiplas URLs (video separado de audio em DASH).
	// Pegamos a primeira — pra bestaudio[ext=m4a] é tipicamente uma única URL.
	urls := splitNonEmptyLines(stdout)
	if len(urls) == 0 {
		return "", time.Time{}, errors.New("youtube: yt-dlp não retornou URL de stream")
	}
	streamURL := urls[0]
	expires := time.Now().Add(4 * time.Hour)
	return streamURL, expires, nil
}

// run executa o yt-dlp e captura stdout. stderr é incluído na mensagem de erro.
func (p *Provider) run(ctx context.Context, args []string) (string, error) {
	if p.ytDlpPath == "" {
		return "", ErrNotAvailable
	}
	// Valida o cookiesFile se configurado (erro claro em vez de yt-dlp silencioso)
	if p.cookiesFile != "" {
		if _, err := os.Stat(p.cookiesFile); err != nil {
			return "", fmt.Errorf("youtube: cookies file %s inacessível: %w", filepath.Base(p.cookiesFile), err)
		}
	}

	// 30s default; search pode demorar, stream é rápido
	cctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cctx, p.ytDlpPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", errors.New(errMsg)
	}
	return stdout.String(), nil
}

// parseSearchOutput parseia o JSON Lines que yt-dlp emite com --dump-json.
// Mesmo em --flat-playlist, cada linha é um JSON completo com metadados do vídeo.
func parseSearchOutput(out string) []models.Track {
	var tracks []models.Track
	scanner := bufio.NewScanner(strings.NewReader(out))
	// Aumenta buffer — vídeos muito longos podem ter descrições grandes
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry struct {
			ID         string  `json:"id"`
			Title      string  `json:"title"`
			Channel    string  `json:"channel"`
			Uploader   string  `json:"uploader"`
			Artist     string  `json:"artist"`
			Creator    string  `json:"creator"`
			Duration   float64 `json:"duration"`
			Thumbnails []struct {
				URL string `json:"url"`
			} `json:"thumbnails"`
		}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if entry.ID == "" || entry.Title == "" {
			continue
		}
		artist := entry.Artist
		if artist == "" {
			artist = entry.Creator
		}
		if artist == "" {
			artist = entry.Uploader
		}
		if artist == "" {
			artist = entry.Channel
		}
		thumbnail := ""
		if len(entry.Thumbnails) > 0 {
			// Pega o último (tipicamente maior resolução)
			thumbnail = entry.Thumbnails[len(entry.Thumbnails)-1].URL
		}
		dur := int(entry.Duration)
		if dur < 0 {
			dur = 0
		}
		tracks = append(tracks, models.Track{
			Title:     entry.Title,
			Artist:    artist,
			Duration:  dur,
			Thumbnail: thumbnail,
			Source:    models.SourceYouTube,
			SourceID:  entry.ID,
		})
	}
	return tracks
}

func splitNonEmptyLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func (p *Provider) cacheGet(key string) ([]models.Track, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	e, ok := p.cache.data[key]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.tracks, true
}

func (p *Provider) cachePut(key string, tracks []models.Track) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cache.data[key] = searchCacheEntry{
		tracks:    tracks,
		expiresAt: time.Now().Add(p.cache.ttl),
	}
}

// ErrNotAvailable indica que yt-dlp não está instalado no sistema.
var ErrNotAvailable = errors.New("youtube: yt-dlp não encontrado no PATH")
