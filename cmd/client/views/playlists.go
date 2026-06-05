package views

import (
	"fmt"
	"music-unified-backend/cmd/client/api"
	"music-unified-backend/pkg/utils"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PlaylistView struct {
	Client       *api.Client
	Playlists    []api.Playlist
	Selected     int
	Back         bool
	Mode         int // 0: list, 1: create, 2: tracks of selected playlist
	NewName      string
	IsLoading    bool
	Error        string
	SuccessMsg   string
	Pending      bool
	width        int
	height       int

	// Tracks view
	TracksSelected int
	scrollOffset   int
	State          playbackState
	PlayingURI     string
	PlayerErr      string

	// Local playback (mpv) pra SoundCloud/YouTube na playlist
	LocalPlayer    *utils.Player
	LocalURL       string
	LocalStartSec  int       // posição (em segundos) onde o player local começou
	LocalStartedAt time.Time // quando o player local iniciou (pra PositionSec)

	// PositionSec é o tempo decorrido da faixa atual (em segundos),
	// atualizado em cada Tick. Pra Spotify Connect vem do
	// (SpotifyLastProgressMs + (now - SpotifyLastPollAt)) / 1000 — sem
	// acumulação de drift. Pra local é calculado a partir de
	// LocalStartSec + (now - LocalStartedAt).
	PositionSec int

	// Estado de sincronização com Spotify Connect. Guardamos o último
	// snapshot do /currently-playing e interpolamos a posição real em
	// qualquer momento a partir do `timestamp` (epoch ms) que o Spotify
	// devolve junto com `progress_ms`. Isso elimina o drift do
	// PositionSec++ a cada tick.
	SpotifyLastProgressMs  int       // progress_ms no momento do último poll
	SpotifyLastTimestamp   int64     // timestamp (epoch ms) do último poll
	SpotifyLastPollAt      time.Time // quando recebemos a resposta
	SpotifyItemDurationMs  int       // duração da faixa atual (vinda do poll)
	SpotifyItemID          string    // id da faixa atual (pra detectar troca)
}

func NewPlaylistView(client *api.Client) *PlaylistView {
	return &PlaylistView{
		Client:    client,
		Mode:      0,
		Playlists: []api.Playlist{},
		State:     stateStopped,
	}
}

func (v *PlaylistView) Update(msg tea.Msg, client *api.Client) (*PlaylistView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
	case tea.KeyMsg:
		switch v.Mode {
		case 0:
			return v.updateListMode(msg)
		case 1:
			return v.updateCreateMode(msg)
		case 2:
			return v.updateTracksMode(msg)
		}
	case playlistsLoadedMsg:
		v.IsLoading = false
		if msg.err != nil {
			v.Error = msg.err.Error()
			return v, nil
		}
		v.Playlists = msg.playlists
		v.Error = ""
	case playResultMsg:
		v.Pending = false
		if msg.err != nil {
			v.PlayerErr = msg.err.Error()
			return v, nil
		}
		v.PlayerErr = ""
		v.State = msg.newState
		if msg.uri != "" {
			v.PlayingURI = msg.uri
			// Limpa estado de sync com Spotify Connect — novo poll
			// vai repopular timestamp/progress_ms.
			v.SpotifyLastProgressMs = 0
			v.SpotifyLastTimestamp = 0
			v.SpotifyLastPollAt = time.Time{}
			v.SpotifyItemDurationMs = 0
			v.SpotifyItemID = ""
			// Pra Spotify Connect, busca a posição atual imediatamente
			// + um poll rápido em 1s (caso o Spotify ainda esteja
			// propagando) + repoll a cada 5s. Pra local, a posição é
			// calculada em `localResolveMsg` (acima).
			return v, tea.Batch(
				waitDurationCmd(msg.durationSec, msg.uri),
				prefetchNextLocalCmd(v),
				positionTickCmd(),
				pollSpotifyPositionCmd(v.Client),
				initialSpotifyPositionCmd(),
			)
		}
		// Resume (uri vazio) — re-inicia o tick de posição
		if v.State == statePlaying && v.PlayingURI != "" {
			return v, tea.Batch(
				positionTickCmd(),
				pollSpotifyPositionCmd(v.Client),
				initialSpotifyPositionCmd(),
			)
		}
	case simpleResultMsg:
		v.Pending = false
		if msg.err != nil {
			v.PlayerErr = msg.err.Error()
		}
	case localResolveMsg:
		v.Pending = false
		if msg.err != nil {
			v.PlayerErr = "Erro ao resolver stream: " + msg.err.Error()
			return v, nil
		}
		v.stopLocal()
		player, err := utils.Play(msg.stream, msg.startSec)
		if err != nil {
			v.PlayerErr = err.Error()
			v.State = stateStopped
			return v, nil
		}
		v.LocalPlayer = player
		v.LocalURL = msg.stream
		v.LocalStartSec = msg.startSec
		v.LocalStartedAt = time.Now()
		v.State = statePlaying
		v.PlayingURI = playingURIFor(msg.track)
		v.PositionSec = msg.startSec
		v.PlayerErr = ""
		trackKey := v.PlayingURI
		// Dispara em paralelo: watcher de fim de faixa + pre-warm da próxima
		// faixa (se for YouTube/SoundCloud) pra eliminar a latência de
		// resolução de URL na transição automática. O tick de 1s repinta a
		// UI pra mostrar o tempo decorrido na faixa atual.
		return v, tea.Batch(
			waitLocalTrackCmd(player, trackKey),
			prefetchNextLocalCmd(v),
			positionTickCmd(),
		)
	case positionTickMsg:
		// Tick periódico de 1s pra repintar a UI com posição atualizada.
		// Se a faixa ainda está tocando, re-agenda o próximo tick. Se
		// não (parada/trocada), deixa o tick morrer — o próximo play
		// vai re-agendar.
		if v.State == statePlaying && v.PlayingURI != "" {
			if v.LocalPlayer != nil {
				// Player local: posição é calculada pelo clock local
				v.PositionSec = v.currentLocalPositionSec()
				return v, positionTickCmd()
			}
			if isConnectURI(v.PlayingURI) {
				// Spotify Connect: interpola entre polls usando o
				// `timestamp` que o Spotify devolve junto com progress_ms.
				// Isso elimina o drift que vinha do PositionSec++ e
				// mantém a posição precisa mesmo se um poll atrasar.
				v.PositionSec = v.currentSpotifyPositionSec()
				return v, positionTickCmd()
			}
		}
	case pollSpotifyPositionMsg:
		// Resposta do /currently-playing chegou. Se for da faixa atual,
		// atualiza posição (corrige drift do tick de 1s). Re-pola a
		// cada 5s enquanto estiver tocando. NÃO re-agenda o tick de
		// 1s — o tick handler já se auto-sustenta.
		if msg.err != nil {
			return v, nil
		}
		if msg.uri != v.PlayingURI {
			return v, nil
		}
		// Salva o snapshot pra interpolação entre polls. O `timestamp`
		// é a chave: a posição real em qualquer "agora" é
		// progressMs + (now - timestamp).
		v.SpotifyLastProgressMs = msg.progressMs
		v.SpotifyLastTimestamp = msg.timestamp
		v.SpotifyLastPollAt = time.Now()
		v.SpotifyItemDurationMs = msg.itemDurationMs
		v.SpotifyItemID = msg.itemID
		v.PositionSec = msg.progressMs / 1000
		if msg.isPlaying && v.State == statePlaying {
			return v, repollSpotifyPositionCmd()
		}
		// Pausado/parado: reflete o estado e atualiza a posição
		if !msg.isPlaying && v.State == statePlaying {
			v.State = statePaused
		}
		return v, nil
	case repollSpotifyPositionMsg:
		// Timer de 5s expirou — re-consulta /currently-playing se a
		// faixa Connect ainda está tocando. Tick de 1s continua por
		// conta própria (a partir do positionTickCmd inicial).
		if v.State == statePlaying && v.PlayingURI != "" && isConnectURI(v.PlayingURI) {
			return v, pollSpotifyPositionCmd(v.Client)
		}
	case initialSpotifyPositionMsg:
		// Poll rápido 1s depois do play. Pode pegar a posição real
		// caso o Spotify ainda não tivesse propagado a faixa no
		// poll inicial (que dispara junto com o play).
		if v.State == statePlaying && v.PlayingURI != "" && isConnectURI(v.PlayingURI) {
			return v, pollSpotifyPositionCmd(v.Client)
		}
	case seekResultMsg:
		// Resposta do POST /player/seek. Se Spotify aceitou, re-polla
		// imediatamente pra confirmar a nova posição (o handler do
		// poll vai atualizar SpotifyLastProgressMs/SpotifyLastTimestamp,
		// e o tick passa a interpolar a partir da nova posição).
		v.Pending = false
		if msg.err != nil {
			v.PlayerErr = "Erro ao seek: " + msg.err.Error()
			return v, nil
		}
		v.PlayerErr = ""
		if v.State == statePlaying && v.PlayingURI != "" && isConnectURI(v.PlayingURI) {
			return v, pollSpotifyPositionCmd(v.Client)
		}
	case trackEndedMsg:
		// Mensagem obsoleta: o usuário já trocou de faixa manualmente
		// (PlayingURI já é outra, ou foi limpa no estado stopped).
		if msg.trackKey != v.PlayingURI {
			return v, nil
		}
		// O timer é a fonte de verdade pra "faixa acabou": quando ele
		// dispara, o tempo decorrido desde o play já é >= duração
		// esperada da faixa. Avança pra próxima independente do state
		// atual (incluindo paused — se o usuário tinha pausado e o
		// timer bateu, a faixa é considerada encerrada).
		tracks := v.currentTracks()
		nextIdx := v.TracksSelected + 1
		if nextIdx >= len(tracks) {
			// Última faixa: para a reprodução e limpa o estado.
			// O Spotify Connect continua tocando (não temos como
			// saber se é a última da fila externa), mas a UI
			// marca como stopped e o user pode decidir.
			v.State = stateStopped
			v.PlayingURI = ""
			return v, nil
		}
		return v, v.playTrackAt(nextIdx)
	}
	return v, nil
}

func (v *PlaylistView) updateListMode(msg tea.KeyMsg) (*PlaylistView, tea.Cmd) {
	switch msg.String() {
	case "esc":
		v.Back = true
	case "ctrl+n":
		v.Mode = 1
		v.NewName = ""
		v.Error = ""
	case "up", "k":
		if v.Selected > 0 {
			v.Selected--
		}
	case "down", "j":
		if v.Selected < len(v.Playlists)-1 {
			v.Selected++
		}
	case "enter":
		if v.Selected >= 0 && v.Selected < len(v.Playlists) {
			v.Mode = 2
			v.TracksSelected = 0
			v.scrollOffset = 0
			v.PlayerErr = ""
		}
	}
	return v, nil
}

func (v *PlaylistView) updateCreateMode(msg tea.KeyMsg) (*PlaylistView, tea.Cmd) {
	switch msg.String() {
	case "esc":
		v.Mode = 0
		v.NewName = ""
		v.Error = ""
	case "enter":
		if v.NewName == "" {
			return v, nil
		}
		name := v.NewName
		_, err := v.Client.CreatePlaylist(name)
		if err != nil {
			v.Error = err.Error()
			return v, nil
		}
		v.Mode = 0
		v.NewName = ""
		v.SuccessMsg = "✓ Playlist \"" + name + "\" criada"
		v.Error = ""
		if pls, err := v.Client.GetPlaylists(); err == nil {
			v.Playlists = pls
		}
	case "backspace":
		if len(v.NewName) > 0 {
			v.NewName = v.NewName[:len(v.NewName)-1]
		}
	default:
		if len(msg.String()) == 1 {
			v.NewName += msg.String()
		}
	}
	return v, nil
}

func (v *PlaylistView) updateTracksMode(msg tea.KeyMsg) (*PlaylistView, tea.Cmd) {
	tracks := v.currentTracks()
	switch msg.String() {
	case "esc":
		v.Mode = 0
		v.PlayerErr = ""
	case "up", "k":
		if v.TracksSelected > 0 {
			v.TracksSelected--
			v.adjustScroll()
		}
	case "down", "j":
		if v.TracksSelected < len(tracks)-1 {
			v.TracksSelected++
			v.adjustScroll()
		}
	case "ctrl+p":
		if v.Pending {
			return v, nil
		}
		if len(tracks) == 0 {
			return v, nil
		}
		track := tracks[v.TracksSelected]
		// Se o track selecionado já é o que está tocando e está pausado, resume
		if v.PlayingURI == playingURIFor(track) && v.State == statePaused {
			if isConnectSource(track.Source) {
				v.PlayerErr = ""
				v.Pending = true
				return v, resumeCmd(v.Client)
			}
			// Local: re-resolve a URL e retoma da posição salva
			startSec := v.LocalStartSec
			if v.LocalPlayer != nil {
				startSec = v.currentLocalPositionSec()
			} else {
				startSec = v.PositionSec
			}
			v.PlayerErr = ""
			v.Pending = true
			return v, resolveAndPlayLocalCmd(v.Client, track, startSec)
		}
		return v, v.playTrackAt(v.TracksSelected)
	case "ctrl+s":
		if v.State != statePlaying || v.Pending {
			return v, nil
		}
		if len(tracks) > 0 && isConnectSource(tracks[v.TracksSelected].Source) {
			v.PlayerErr = ""
			v.Pending = true
			return v, pauseCmd(v.Client)
		}
		// Local: salva a posição atual e mata o player (resume usa --start)
		if v.LocalPlayer != nil {
			v.LocalStartSec = v.currentLocalPositionSec()
		}
		v.PositionSec = v.LocalStartSec
		v.stopLocal()
		v.State = statePaused
	case "ctrl+n":
		// Próxima faixa da playlist (qualquer serviço)
		if v.Pending || len(tracks) == 0 {
			return v, nil
		}
		nextIdx := v.TracksSelected + 1
		if nextIdx >= len(tracks) {
			return v, nil
		}
		return v, v.playTrackAt(nextIdx)
	case "ctrl+b":
		// Faixa anterior da playlist (qualquer serviço)
		if v.Pending || len(tracks) == 0 {
			return v, nil
		}
		prevIdx := v.TracksSelected - 1
		if prevIdx < 0 {
			return v, nil
		}
		return v, v.playTrackAt(prevIdx)
	case "ctrl+left", "ctrl+right":
		// Seek no Spotify Connect: -10s / +10s. Ignora se não há faixa
		// tocando via Connect, se está pending (play/seek anterior não
		// acabou) ou se estamos com player local (cujo seek é tratado
		// pelo mpv/ffplay via --start, não via Connect).
		if v.Pending {
			return v, nil
		}
		if v.PlayingURI == "" || !isConnectURI(v.PlayingURI) {
			return v, nil
		}
		if v.State != statePlaying && v.State != statePaused {
			return v, nil
		}
		delta := 10
		if msg.String() == "ctrl+left" {
			delta = -10
		}
		current := v.currentSpotifyPositionSec()
		// Limites: 0 .. duração (se conhecida). Sem duração conhecida
		// (poll inicial não chegou), só evita negativo.
		max := current
		if v.SpotifyItemDurationMs > 0 {
			max = v.SpotifyItemDurationMs / 1000
		} else {
			for _, t := range tracks {
				if playingURIFor(t) == v.PlayingURI {
					max = t.Duration
					break
				}
			}
		}
		newSec := current + delta
		if newSec < 0 {
			newSec = 0
		}
		if max > 0 && newSec > max {
			newSec = max
		}
		v.Pending = true
		v.PlayerErr = ""
		return v, seekCmd(v.Client, newSec*1000, delta)
	}
	return v, nil
}

func (v *PlaylistView) stopLocal() {
	if v.LocalPlayer != nil {
		_ = v.LocalPlayer.Stop()
		v.LocalPlayer = nil
	}
	v.LocalURL = ""
}

// currentLocalPositionSec retorna a posição atual (em segundos) do
// player local, baseada em LocalStartSec + elapsed desde LocalStartedAt.
// Se o player não existe (parado), devolve LocalStartSec (que contém a
// posição salva no pause).
func (v *PlaylistView) currentLocalPositionSec() int {
	if v.LocalPlayer == nil {
		return v.LocalStartSec
	}
	elapsed := int(time.Since(v.LocalStartedAt).Seconds())
	return v.LocalStartSec + elapsed
}

// currentSpotifyPositionSec retorna a posição atual (em segundos) da
// faixa do Spotify Connect. A posição real é `SpotifyLastProgressMs +
// (now - SpotifyLastTimestamp)` — o `timestamp` é epoch ms que o
// Spotify devolve junto com progress_ms, então a interpolação é
// precisa em ms sem polling constante.
//
// Se ainda não temos um poll (SpotifyLastPollAt zero), devolve 0
// (a UI mostra 0:00 / X:XX até o primeiro poll chegar).
func (v *PlaylistView) currentSpotifyPositionSec() int {
	if v.SpotifyLastPollAt.IsZero() || v.SpotifyLastTimestamp == 0 {
		return v.SpotifyLastProgressMs / 1000
	}
	elapsedMs := time.Since(v.SpotifyLastPollAt).Milliseconds()
	totalMs := v.SpotifyLastProgressMs + int(elapsedMs)
	return totalMs / 1000
}

// playTrackAt inicia a faixa no índice `idx` da playlist atual, fazendo
// a transição entre serviços quando necessário:
//   - Spotify → Spotify: /player/play com a nova URI (substitui a fila)
//   - Spotify → local: pausa Spotify + inicia mpv com a URL resolvida
//   - local → Spotify: mata o player local + /player/play com a URI
//   - local → local: utils.Play mata o anterior e inicia o novo
//
// Atualiza TracksSelected, PlayingURI, Pending. O watch command
// (waitLocalTrackCmd ou waitDurationCmd) é despachado depois que o
// playResultMsg/localResolveMsg confirma o início da reprodução.
func (v *PlaylistView) playTrackAt(idx int) tea.Cmd {
	tracks := v.currentTracks()
	if idx < 0 || idx >= len(tracks) {
		return nil
	}
	track := tracks[idx]
	v.TracksSelected = idx
	v.Pending = true
	v.PlayerErr = ""
	v.PlayingURI = playingURIFor(track)

	if isConnectSource(track.Source) {
		// Saindo do player local pro Spotify: mata o mpv
		v.stopLocal()
		uri, err := buildPlayableURI(track)
		if err != nil {
			v.Pending = false
			v.PlayerErr = err.Error()
			return nil
		}
		return playTrackCmd(v.Client, uri, track.Duration)
	}

	// Saindo do Spotify pro player local: pausa a reprodução no Connect
	if v.State == statePlaying || v.State == statePaused {
		_ = v.Client.Pause()
	}
	v.stopLocal()
	if track.SourceID == "" {
		v.Pending = false
		v.PlayerErr = "Faixa sem ID"
		return nil
	}
	return resolveAndPlayLocalCmd(v.Client, track, 0)
}

func (v *PlaylistView) currentTracks() []api.Track {
	if v.Selected < 0 || v.Selected >= len(v.Playlists) {
		return nil
	}
	return v.Playlists[v.Selected].Tracks
}

// buildPlayableURI gera URI no formato esperado pelo Spotify Connect
// pra faixas do Spotify. Pra outras fontes, retorna erro — playback local
// é tratado via resolveAndPlayLocalCmd.
func buildPlayableURI(track api.Track) (string, error) {
	if track.Source != "spotify" {
		return "", fmt.Errorf("playback local: use resolveAndPlayLocalCmd para %s", track.Source)
	}
	if track.SourceID == "" {
		return "", fmt.Errorf("faixa sem ID do Spotify")
	}
	return "spotify:track:" + track.SourceID, nil
}

// formatDuration converte segundos em "m:ss" ou "h:mm:ss".
func formatDuration(sec int) string {
	if sec < 0 {
		sec = 0
	}
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

// truncate corta uma string pelo visual width (respeitando ANSI), adicionando
// "…" se passar do limite. Retorna a string original se já couber.
func truncate(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	if maxWidth <= 1 {
		return strings.Repeat("…", maxWidth)
	}
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes))+1 > maxWidth {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}

// sourceChip retorna um label colorido de 3 chars representando a fonte.
// Spotify = magenta, YouTube = vermelho, SoundCloud = ciano.
func sourceChip(source string) string {
	var label, color string
	switch source {
	case "spotify":
		label, color = "SPT", "5"
	case "youtube":
		label, color = "YT ", "1"
	case "soundcloud":
		label, color = "SC ", "6"
	default:
		label = strings.ToUpper(source)
		if len(label) > 3 {
			label = label[:3]
		}
		for len(label) < 3 {
			label += " "
		}
		color = "8"
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Render(label)
}

// padTo preenche à direita até `width` (medido por visual width).
func padTo(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

// rightPadTo preenche à esquerda até `width`.
func rightPadTo(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return strings.Repeat(" ", width-w) + s
}

// effectiveWidth retorna a largura útil do terminal (default 80).
func (v *PlaylistView) effectiveWidth() int {
	if v.width > 0 {
		return v.width
	}
	return 80
}

// visibleTrackCount calcula quantas faixas cabem na viewport atual
// considerando header + footer. Cada faixa ocupa 2 linhas (título + artista).
func (v *PlaylistView) visibleTrackCount() int {
	const reserved = 11 // header(2) + colHeader(2) + footer(3) + help(2) + blanks(2)
	h := v.height
	if h == 0 {
		return 100
	}
	n := (h - reserved) / 2
	if n < 3 {
		n = 3
	}
	tracks := v.currentTracks()
	if len(tracks) < n {
		n = len(tracks)
	}
	return n
}

// adjustScroll mantém a faixa selecionada visível. Chamado após mudar
// TracksSelected via navegação (j/k ou setas).
func (v *PlaylistView) adjustScroll() {
	visible := v.visibleTrackCount()
	if visible <= 0 {
		v.scrollOffset = 0
		return
	}
	if v.TracksSelected < v.scrollOffset {
		v.scrollOffset = v.TracksSelected
	}
	if v.TracksSelected >= v.scrollOffset+visible {
		v.scrollOffset = v.TracksSelected - visible + 1
	}
	if v.scrollOffset < 0 {
		v.scrollOffset = 0
	}
}

func (v *PlaylistView) View() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5")).
		Render("📋 My Playlists")

	switch v.Mode {
	case 0:
		return v.viewList(title)
	case 1:
		return v.viewCreate(title)
	case 2:
		return v.viewTracks(title)
	}
	return ""
}

func (v *PlaylistView) viewList(title string) string {
	var content string
	if v.Error != "" {
		content = errorStyle.Render("❌ " + v.Error)
	} else if len(v.Playlists) == 0 {
		content = dimStyle.Render("Nenhuma playlist. Pressione Ctrl+N para criar.")
	} else {
		for i, pl := range v.Playlists {
			style := lipgloss.NewStyle()
			if i == v.Selected {
				style = lipgloss.NewStyle().
					Background(lipgloss.Color("4")).
					Foreground(lipgloss.Color("15")).
					Padding(0, 1)
			}
			content += style.Render(fmt.Sprintf("%s (%d tracks)\n", pl.Name, len(pl.Tracks)))
		}
	}

	status := ""
	if v.SuccessMsg != "" {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(v.SuccessMsg)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		content,
		"",
		status,
		"",
		"↑/↓: Nav | Enter: Open | Ctrl+N: New | Esc: Back",
	)
}

func (v *PlaylistView) viewCreate(title string) string {
	input := lipgloss.NewStyle().
		Background(lipgloss.Color("4")).
		Foreground(lipgloss.Color("15")).
		Padding(0, 1).
		Render("Playlist name: " + v.NewName + "_")

	status := ""
	if v.Error != "" {
		status = errorStyle.Render("⚠ " + v.Error)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		input,
		"",
		status,
		"",
		"Enter: Create | Esc: Cancel",
	)
}

// Layout columns. Larguras fixas exceto o título, que é flex até o limite
// do terminal. "numCol" alinha o índice/marcador à direita; o título e
// artista ficam indentados pelo mesmo valor pra alinhar visualmente.
const (
	ptNumColW   = 3
	ptArtistW   = 22
	ptDurW      = 12 // "1:23 / 4:31" na faixa tocando; só o total (padded) nas demais
	ptSourceW   = 3
	ptColGap    = 2
	ptMaxTitleW = 60
	ptIndent    = ptNumColW + ptColGap // colunas à esquerda do título (marker + gap)
)

func (v *PlaylistView) viewTracks(title string) string {
	if v.Selected < 0 || v.Selected >= len(v.Playlists) {
		return title
	}
	pl := v.Playlists[v.Selected]
	tracks := pl.Tracks

	parts := []string{
		title,
		"",
		v.renderPTHeader(pl),
		"",
		v.renderPTColumnHeader(),
		v.renderPTColumnSeparator(),
		v.renderPTTrackList(tracks),
		"",
		v.renderPTNowPlaying(tracks),
		"",
		dimStyle.Render("↑/↓: Nav    Ctrl+P: Play    Ctrl+S: Pause    Ctrl+←/→: Seek ±10s    Ctrl+N: Next    Ctrl+B: Previous    Esc: Back"),
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderPTHeader desenha o cabeçalho da playlist: nome grande à esquerda,
// meta (qtd + duração total) à direita.
func (v *PlaylistView) renderPTHeader(pl api.Playlist) string {
	name := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5")).
		Padding(0, 0).
		Render("▶ " + pl.Name)

	n := len(pl.Tracks)
	total := 0
	for _, t := range pl.Tracks {
		total += t.Duration
	}
	meta := fmt.Sprintf("%d %s · %s", n, pluralMusic(n), formatDuration(total))
	metaStyled := dimStyle.Render(meta)

	avail := v.effectiveWidth()
	nameW := lipgloss.Width(name)
	metaW := lipgloss.Width(metaStyled)
	gap := avail - nameW - metaW
	if gap < 2 {
		gap = 2
	}
	return name + strings.Repeat(" ", gap) + metaStyled
}

func pluralMusic(n int) string {
	if n == 1 {
		return "música"
	}
	return "músicas"
}

// renderPTColumnHeader desenha o cabeçalho das colunas (#, TÍTULO, ARTISTA, ⏱, FONTE).
func (v *PlaylistView) renderPTColumnHeader() string {
	hdr := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Bold(true)
	col := func(s string, w int) string { return padTo(hdr.Render(s), w) }
	parts := []string{
		rightPadTo(hdr.Render("#"), ptNumColW),
		col("TÍTULO", v.ptTitleColWidth()),
		col("ARTISTA", ptArtistW),
		rightPadTo(hdr.Render("⏱"), ptDurW),
		hdr.Render("FONTE"),
	}
	return strings.Join(parts, strings.Repeat(" ", ptColGap))
}

// renderPTColumnSeparator desenha uma linha horizontal abaixo do cabeçalho.
func (v *PlaylistView) renderPTColumnSeparator() string {
	return dimStyle.Render(strings.Repeat("─", v.effectiveWidth()))
}

// ptTitleColWidth calcula a largura do título (flex) dado o espaço disponível.
func (v *PlaylistView) ptTitleColWidth() int {
	used := ptNumColW + ptArtistW + ptDurW + ptSourceW + (4 * ptColGap)
	avail := v.effectiveWidth() - used
	if avail < 16 {
		avail = 16
	}
	if avail > ptMaxTitleW {
		avail = ptMaxTitleW
	}
	return avail
}

// renderPTTrackList renderiza as faixas visíveis (com scroll).
func (v *PlaylistView) renderPTTrackList(tracks []api.Track) string {
	if len(tracks) == 0 {
		return dimStyle.Render("  (vazia — pressione Esc e adicione faixas pela busca)")
	}
	visible := v.visibleTrackCount()
	if v.scrollOffset > len(tracks)-1 {
		v.scrollOffset = 0
	}
	end := v.scrollOffset + visible
	if end > len(tracks) {
		end = len(tracks)
	}
	window := tracks[v.scrollOffset:end]

	var rows []string
	for i, track := range window {
		absIdx := v.scrollOffset + i
		rows = append(rows, v.renderPTRow(absIdx, track))
	}

	// Indicador de scroll no rodapé da lista
	out := strings.Join(rows, "\n")
	if v.scrollOffset > 0 || end < len(tracks) {
		indicator := fmt.Sprintf("  ⋮ %d–%d de %d", v.scrollOffset+1, end, len(tracks))
		out += "\n" + dimStyle.Render(indicator)
	}
	return out
}

// renderPTRow renderiza uma faixa individual (2 linhas: título + artista).
func (v *PlaylistView) renderPTRow(i int, track api.Track) string {
	isSelected := i == v.TracksSelected
	isPlaying := v.PlayingURI == playingURIFor(track) && v.State != stateStopped

	// Marker (#) column
	marker := v.renderPTMarker(i, isPlaying, isSelected)
	markerCol := rightPadTo(marker, ptNumColW)

	// Title
	titleW := v.ptTitleColWidth()
	titleStr := truncate(track.Title, titleW)
	titleStr = v.renderPTTitle(titleStr, isSelected, isPlaying)

	// Artist
	artistStr := truncate(track.Artist, ptArtistW)
	if track.Artist == "" {
		artistStr = "—"
	}
	artistStr = v.renderPTArtist(artistStr, isSelected)

	// Duration: pra faixa tocando, mostra "elapsed / total" (ex.: "1:23 / 4:31");
	// pras demais só o total. A faixa tocando usa uma coluna mais larga
	// pra acomodar os dois tempos.
	durStr := v.renderPTDuration(track, isPlaying)

	// Source chip
	srcStr := sourceChip(track.Source)

	// Linha 1: marker + title  ........  duration + source
	line1Left := markerCol + strings.Repeat(" ", ptColGap) + titleStr
	line1Right := durStr + strings.Repeat(" ", ptColGap) + srcStr
	avail := v.effectiveWidth()
	gap := avail - lipgloss.Width(line1Left) - lipgloss.Width(line1Right)
	if gap < 1 {
		gap = 1
	}
	line1 := line1Left + strings.Repeat(" ", gap) + line1Right

	// Linha 2: indent + artist (mesma indentação da linha 1)
	line2 := strings.Repeat(" ", ptIndent) + artistStr

	// Selection: aplica background na linha toda
	if isSelected {
		bg := lipgloss.NewStyle().Background(lipgloss.Color("4"))
		line1 = bg.Foreground(lipgloss.Color("15")).Render(truncate(line1, avail))
		line2 = bg.Foreground(lipgloss.Color("15")).Render(truncate(line2, avail))
	} else {
		// Sem seleção: garante que as linhas preencham a largura pro background
		// do título não ficar curto
		if pad := avail - lipgloss.Width(line1); pad > 0 {
			line1 = line1 + strings.Repeat(" ", pad)
		}
		if pad := avail - lipgloss.Width(line2); pad > 0 {
			line2 = line2 + strings.Repeat(" ", pad)
		}
	}

	return line1 + "\n" + line2
}

// renderPTMarker desenha o índice/indicador de play. Verde pra tocando,
// cinza pra índice normal. Selected recebe emphasis extra (bold).
func (v *PlaylistView) renderPTMarker(i int, isPlaying, isSelected bool) string {
	if isPlaying {
		var icon string
		switch v.State {
		case statePlaying:
			icon = "▶"
		case statePaused:
			icon = "⏸"
		default:
			icon = fmt.Sprintf("%d", i+1)
		}
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
		if isSelected {
			style = style.Bold(true)
		}
		return style.Render(icon)
	}
	num := dimStyle.Render(fmt.Sprintf("%d", i+1))
	if isSelected {
		num = lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%d", i+1))
	}
	return num
}

// renderPTTitle aplica estilo ao título. Tocando = verde, selecionado = branco
// em negrito sobre fundo azul, normal = branco.
func (v *PlaylistView) renderPTTitle(s string, isSelected, isPlaying bool) string {
	if isSelected {
		// bg já cuida da cor; mantém o conteúdo legível
		return lipgloss.NewStyle().Bold(true).Render(s)
	}
	if isPlaying {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true).Render(s)
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Render(s)
}

// renderPTArtist aplica estilo ao artista. Sempre dim (cinza), exceto na
// linha de seleção onde fica branco pra contraste.
func (v *PlaylistView) renderPTArtist(s string, isSelected bool) string {
	if isSelected {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(s)
	}
	return dimStyle.Render(s)
}

// renderPTDuration monta a string de duração. Pra faixa tocando, mostra
// "elapsed / total" (ex.: "1:23 / 4:31") em verde negrito. Pra demais
// faixas, só o total em dim. Ambas as formas são alinhadas à direita
// dentro da coluna de largura ptDurW (12) pra que o source chip
// permaneça alinhado verticalmente entre as linhas.
func (v *PlaylistView) renderPTDuration(track api.Track, isPlaying bool) string {
	total := formatDuration(track.Duration)
	if !isPlaying {
		return rightPadTo(dimStyle.Render(total), ptDurW)
	}
	pos := v.currentDisplayPositionSec(track)
	elapsed := formatDuration(pos)
	text := elapsed + " / " + total
	// Verde pra indicar "tempo andando"; quando pausado, amarelo
	style := lipgloss.NewStyle().Bold(true)
	if v.State == statePaused {
		style = style.Foreground(lipgloss.Color("3"))
	} else {
		style = style.Foreground(lipgloss.Color("2"))
	}
	return rightPadTo(style.Render(text), ptDurW)
}

// currentDisplayPositionSec retorna a posição (em segundos) a exibir
// na UI. Pra player local usa o clock local (LocalStartSec + elapsed);
// pra Connect usa a interpolação via timestamp (precisa, sem drift).
// Não avança se o estado é "pausado".
func (v *PlaylistView) currentDisplayPositionSec(track api.Track) int {
	if v.State == statePaused {
		return v.PositionSec
	}
	if v.LocalPlayer != nil {
		return v.currentLocalPositionSec()
	}
	if isConnectURI(v.PlayingURI) {
		return v.currentSpotifyPositionSec()
	}
	return v.PositionSec
}

// renderPTNowPlaying desenha a barra de "tocando agora" no rodapé.
func (v *PlaylistView) renderPTNowPlaying(tracks []api.Track) string {
	status := v.renderPTStatusLine(tracks)
	if status == "" {
		return ""
	}
	return status
}

// renderPTStatusLine monta a linha de status. Retorna string vazia se não há
// informação útil (playlist vazia sem play).
func (v *PlaylistView) renderPTStatusLine(tracks []api.Track) string {
	var left, right string

	switch {
	case v.PlayerErr != "":
		left = errorStyle.Render("⚠  " + v.PlayerErr)
	case v.Pending:
		left = dimStyle.Render("⏳  carregando…")
	case v.State == statePlaying:
		left = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true).Render("▶  tocando agora")
	case v.State == statePaused:
		left = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true).Render("⏸  pausado")
	default:
		// Nada tocando — não mostra rodapé pra não poluir
		return ""
	}

	// Detalhe da faixa atual
	if v.PlayingURI != "" {
		for _, t := range tracks {
			if playingURIFor(t) == v.PlayingURI {
				right = fmt.Sprintf("%s  ·  %s — %s", sourceChip(t.Source), t.Title, t.Artist)
				break
			}
		}
	}

	avail := v.effectiveWidth()
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	if rightW > 0 {
		gap := avail - leftW - rightW
		if gap < 2 {
			gap = 2
		}
		return left + strings.Repeat(" ", gap) + dimStyle.Render(right)
	}
	// Preenche até a largura pra conectar com a borda
	if pad := avail - leftW; pad > 0 {
		return left + strings.Repeat(" ", pad)
	}
	return left
}
