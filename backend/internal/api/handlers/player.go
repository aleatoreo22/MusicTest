package handlers

import (
	"music-unified-backend/internal/providers/spotify"
	"music-unified-backend/internal/repository"
	"music-unified-backend/pkg/httputil"
	"strings"

	"github.com/gin-gonic/gin"
	spotifyapi "github.com/zmb3/spotify/v2"
)

type PlayerHandler struct {
	userRepo repository.UserRepository
	provider *spotify.Provider
}

func NewPlayerHandler(repo repository.UserRepository, provider *spotify.Provider) *PlayerHandler {
	return &PlayerHandler{userRepo: repo, provider: provider}
}

type PlayRequest struct {
	URI      string `json:"uri,omitempty"`
	DeviceID string `json:"device_id,omitempty"`
}

type SeekRequest struct {
	PositionMs int    `json:"position_ms"`
	DeviceID   string `json:"device_id,omitempty"`
}

// spotifyClient resolve o client do Spotify para o usuário autenticado.
// Retorna (nil, false) se o client não pôde ser obtido — nesse caso o
// handler já respondeu com o erro apropriado.
func (h *PlayerHandler) spotifyClient(c *gin.Context) (*spotifyapi.Client, bool) {
	userID, ok := httputil.MustUserID(c)
	if !ok {
		return nil, false
	}
	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		httputil.Error(c, httputil.NotFound("Usuário não encontrado"))
		return nil, false
	}
	if user.SpotifyAccessToken == "" {
		httputil.Error(c, httputil.Unauthorized("spotify não vinculado"))
		return nil, false
	}
	return h.provider.GetClient(c.Request.Context(), userToToken(user)), true
}

func (h *PlayerHandler) Play(c *gin.Context) {
	var req PlayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, httputil.BadRequest(err.Error()))
		return
	}
	cli, ok := h.spotifyClient(c)
	if !ok {
		return
	}

	opt := &spotifyapi.PlayOptions{}
	if req.URI != "" {
		opt.URIs = []spotifyapi.URI{spotifyapi.URI(req.URI)}
	}
	if req.DeviceID != "" {
		id := spotifyapi.ID(req.DeviceID)
		opt.DeviceID = &id
	}
	if err := cli.PlayOpt(c.Request.Context(), opt); err != nil {
		httputil.Error(c, classifySpotifyErr(err))
		return
	}
	httputil.JSON(c, 200, gin.H{"ok": true, "uri": req.URI})
}

func (h *PlayerHandler) Pause(c *gin.Context) {
	cli, ok := h.spotifyClient(c)
	if !ok {
		return
	}
	if err := cli.Pause(c.Request.Context()); err != nil {
		// Cleanup do TUI no Ctrl+C: best-effort, nunca falha. Erros
		// "no active device", "premium required", token expirado são
		// todos no-op pra finalidade de "garantir que parou". Respondemos
		// 200 direto nesse caso.
		if c.Query("cleanup") == "1" {
			httputil.JSON(c, 200, gin.H{"ok": true})
			return
		}
		httputil.Error(c, classifySpotifyErr(err))
		return
	}
	httputil.JSON(c, 200, gin.H{"ok": true})
}

func (h *PlayerHandler) Next(c *gin.Context) {
	h.playerSimple(c, func(cli *spotifyapi.Client) error { return cli.Next(c.Request.Context()) })
}

func (h *PlayerHandler) Previous(c *gin.Context) {
	h.playerSimple(c, func(cli *spotifyapi.Client) error { return cli.Previous(c.Request.Context()) })
}

// playerSimple é o helper para ações do player que não precisam de
// body nem parâmetros — todas elas têm a mesma estrutura: pegar
// client, executar ação, classificar erro.
func (h *PlayerHandler) playerSimple(c *gin.Context, action func(*spotifyapi.Client) error) {
	cli, ok := h.spotifyClient(c)
	if !ok {
		return
	}
	if err := action(cli); err != nil {
		httputil.Error(c, classifySpotifyErr(err))
		return
	}
	httputil.JSON(c, 200, gin.H{"ok": true})
}

// Seek move a posição de reprodução do Spotify Connect para
// req.PositionMs (ms, relativo ao início da faixa). Se passar da
// duração, o Spotify pula pra próxima automaticamente.
func (h *PlayerHandler) Seek(c *gin.Context) {
	var req SeekRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, httputil.BadRequest(err.Error()))
		return
	}
	if req.PositionMs < 0 {
		httputil.Error(c, httputil.BadRequest("position_ms deve ser >= 0"))
		return
	}
	cli, ok := h.spotifyClient(c)
	if !ok {
		return
	}
	opt := &spotifyapi.PlayOptions{}
	if req.DeviceID != "" {
		id := spotifyapi.ID(req.DeviceID)
		opt.DeviceID = &id
	}
	if err := cli.SeekOpt(c.Request.Context(), req.PositionMs, opt); err != nil {
		httputil.Error(c, classifySpotifyErr(err))
		return
	}
	httputil.JSON(c, 200, gin.H{"ok": true, "position_ms": req.PositionMs})
}

func (h *PlayerHandler) Devices(c *gin.Context) {
	cli, ok := h.spotifyClient(c)
	if !ok {
		return
	}
	devs, err := cli.PlayerDevices(c.Request.Context())
	if err != nil {
		httputil.Error(c, httputil.Internal(err.Error()))
		return
	}
	httputil.JSON(c, 200, devs)
}

// PlaybackSnapshotFlat é a projeção do CurrentlyPlaying do Spotify SDK
// para os campos que o TUI usa. Retornamos flat (em vez do SDK completo)
// para evitar dependência de JSON tags internos do zmb3/spotify e
// dar nomes mais claros no contrato HTTP.
type PlaybackSnapshotFlat struct {
	IsPlaying      bool   `json:"is_playing"`
	ProgressMs     int    `json:"progress_ms"`
	Timestamp      int64  `json:"timestamp"`
	URI            string `json:"uri"`
	ItemID         string `json:"item_id,omitempty"`
	ItemDurationMs int    `json:"item_duration_ms,omitempty"`
}

func (h *PlayerHandler) CurrentlyPlaying(c *gin.Context) {
	cli, ok := h.spotifyClient(c)
	if !ok {
		return
	}
	cp, err := cli.PlayerCurrentlyPlaying(c.Request.Context())
	if err != nil {
		httputil.Error(c, httputil.Internal(err.Error()))
		return
	}
	if cp == nil {
		httputil.JSON(c, 200, gin.H{"is_playing": false})
		return
	}
	resp := PlaybackSnapshotFlat{
		IsPlaying:  cp.Playing,
		ProgressMs: int(cp.Progress),
		Timestamp:  cp.Timestamp,
	}
	if cp.Item != nil {
		resp.URI = string(cp.Item.URI)
		resp.ItemID = string(cp.Item.ID)
		resp.ItemDurationMs = int(cp.Item.Duration)
	}
	httputil.JSON(c, 200, resp)
}

// classifySpotifyErr converte erros do Spotify em status HTTP. Faz
// string matching nas mensagens (a API não retorna tipos estruturados
// em todos os casos). TODO: usar erros tipados do SDK quando possível.
func classifySpotifyErr(err error) *httputil.APIError {
	msg := err.Error()
	switch {
	case containsAny(msg, "404", "No active device", "device not found"):
		return httputil.NotFound("Nenhum device do Spotify ativo. Abra o Spotify em algum device (celular, desktop ou web player) e tente novamente.")
	case containsAny(msg, "403", "Premium required"):
		return httputil.Forbidden("Spotify Premium é necessário para controlar a reprodução remotamente.")
	case strings.Contains(msg, "401"):
		return httputil.Unauthorized("Token do Spotify expirou. Revincule sua conta.")
	}
	return httputil.Internal(msg)
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
