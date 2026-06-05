package views

import (
	"music-unified-backend/cmd/client/api"
	"music-unified-backend/pkg/utils"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type playbackState int

const (
	stateStopped playbackState = iota
	statePlaying
	statePaused
)

type playResultMsg struct {
	uri         string
	newState    playbackState
	durationSec int
	err         error
}

type simpleResultMsg struct {
	op  string
	err error
}

type playlistsLoadedMsg struct {
	playlists []api.Playlist
	err       error
}

// trackEndedMsg sinaliza que a faixa atual terminou (timer do Spotify ou
// Done() do player local). trackKey identifica qual faixa terminou — o view
// compara com a PlayingURI atual pra descartar mensagens obsoletas
// (ex.: timer de uma faixa que o usuário já pulou manualmente).
type trackEndedMsg struct {
	trackKey string
}

func playTrackCmd(client *api.Client, uri string, durationSec int) tea.Cmd {
	return func() tea.Msg {
		err := client.PlayTrack(uri, "")
		if err != nil {
			return playResultMsg{err: err}
		}
		return playResultMsg{uri: uri, newState: statePlaying, durationSec: durationSec}
	}
}

func resumeCmd(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		err := client.Resume()
		if err != nil {
			return playResultMsg{err: err}
		}
		return playResultMsg{newState: statePlaying}
	}
}

func pauseCmd(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		err := client.Pause()
		if err != nil {
			return simpleResultMsg{op: "pause", err: err}
		}
		return playResultMsg{newState: statePaused}
	}
}

func nextCmd(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		err := client.Next()
		if err != nil {
			return simpleResultMsg{op: "next", err: err}
		}
		return playResultMsg{newState: statePlaying}
	}
}

func previousCmd(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		err := client.Previous()
		if err != nil {
			return simpleResultMsg{op: "previous", err: err}
		}
		return playResultMsg{newState: statePlaying}
	}
}

func loadPlaylistsCmd(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		pls, err := client.GetPlaylists()
		return playlistsLoadedMsg{playlists: pls, err: err}
	}
}

// waitLocalTrackCmd bloqueia até o player local terminar (natural ou kill)
// e devolve um trackEndedMsg com a chave da faixa. Usado pra auto-avançar
// quando uma faixa de SoundCloud/YouTube termina no mpv/ffplay/vlc.
func waitLocalTrackCmd(p *utils.Player, trackKey string) tea.Cmd {
	return func() tea.Msg {
		if p == nil {
			return nil
		}
		<-p.Done()
		return trackEndedMsg{trackKey: trackKey}
	}
}

// waitDurationCmd dispara após `durationSec` segundos. Usado como aproximação
// de "faixa terminou" pra Spotify Connect (que não tem evento de fim). O
// view filtra mensagens obsoletas comparando trackKey com PlayingURI.
func waitDurationCmd(durationSec int, trackKey string) tea.Cmd {
	if durationSec <= 0 {
		return nil
	}
	return func() tea.Msg {
		time.Sleep(time.Duration(durationSec) * time.Second)
		return trackEndedMsg{trackKey: trackKey}
	}
}

// prefetchStreamCmd dispara a resolução de stream URL no servidor em
// background. Não emite mensagem — o objetivo é só aquecer o cache
// (yt-dlp streamCache / SoundCloud HEAD request) antes do play real.
func prefetchStreamCmd(client *api.Client, source, sourceID string) tea.Cmd {
	return func() tea.Msg {
		client.PrefetchStreamURL(source, sourceID)
		return nil
	}
}

// prefetchNextLocalCmd identifica a próxima faixa da playlist (no estado
// atual do view) e dispara pre-warm se ela for de uma fonte com URL
// resolvida (YouTube/SoundCloud). Spotify Connect é instantâneo e não
// precisa de pre-warm. O view é lido em snapshot — se o usuário trocar
// de faixa antes do pre-fetch chegar, a request ainda assim é válida:
// o cache fica populado pra um possível retorno.
func prefetchNextLocalCmd(v *PlaylistView) tea.Cmd {
	tracks := v.currentTracks()
	idx := v.TracksSelected + 1
	if idx >= len(tracks) {
		return nil
	}
	next := tracks[idx]
	if isConnectSource(next.Source) || next.SourceID == "" {
		return nil
	}
	return prefetchStreamCmd(v.Client, next.Source, next.SourceID)
}

// positionTickMsg sinaliza "passou 1 segundo — recalcula a posição da faixa
// atual". O handler atualiza PositionSec (do clock local pro player local,
// ou via pollSpotifyPositionCmd pro Connect) e re-agenda outro tick.
type positionTickMsg struct{}

// positionTickCmd dispara após 1s. Usado em loop: cada handler re-agenda
// o próximo tick se a faixa ainda está tocando. Tem precisão de ~1s, o
// que é suficiente pra UI tipo Spotify (refresh visual a cada segundo).
func positionTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return positionTickMsg{}
	})
}

// pollSpotifyPositionMsg carrega o resultado de uma consulta ao
// /api/v1/player/currently-playing. O handler compara uri com
// PlayingURI pra descartar respostas obsoletas (troca de faixa).
//
// Os campos `progressMs` e `timestamp` juntos definem a posição de
// reprodução do Spotify Connect no momento da resposta. O client usa
// timestamp pra interpolar a posição em tempo real entre polls
// (eliminando o drift que vinha do `PositionSec++` a cada tick).
type pollSpotifyPositionMsg struct {
	uri           string
	progressMs    int
	timestamp     int64
	itemID        string
	itemDurationMs int
	isPlaying     bool
	err           error
}

// repollSpotifyPositionMsg é o "sentinel" que dispara nova consulta ao
// /currently-playing. Disparado a cada 5s enquanto a faixa Connect estiver
// tocando — mantém PositionSec razoavelmente preciso entre ticks de 1s.
type repollSpotifyPositionMsg struct{}

// initialSpotifyPositionMsg é o "sentinel" do poll inicial rápido
// (1s) disparado logo depois de começar a tocar. Serve pra corrigir
// a posição estimada no primeiro segundo de reprodução, em que o
// Spotify pode ainda não ter propagado a faixa no Connect.
type initialSpotifyPositionMsg struct{}

// seekResultMsg é o resultado de um POST /player/seek. err==nil
// significa que o Spotify aceitou o seek. A UI deve re-pollar pra
// confirmar a nova posição.
type seekResultMsg struct {
	deltaSec int
	err      error
}

// pollSpotifyPositionCmd consulta o estado atual de reprodução no
// Spotify Connect. É um instantâneo — não bloqueia, é melhor-esforço.
// A UI usa o `timestamp` da resposta pra interpolar a posição entre
// polls (precisão de ms sem polling constante).
func pollSpotifyPositionCmd(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		snap, err := client.CurrentlyPlaying()
		if err != nil || snap == nil {
			return pollSpotifyPositionMsg{err: err}
		}
		return pollSpotifyPositionMsg{
			uri:            snap.URI,
			progressMs:     snap.ProgressMs,
			timestamp:      snap.Timestamp,
			itemID:         snap.ItemID,
			itemDurationMs: snap.ItemDurationMs,
			isPlaying:      snap.Playing,
		}
	}
}

// repollSpotifyPositionCmd agenda uma nova consulta em 5s.
func repollSpotifyPositionCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return repollSpotifyPositionMsg{}
	})
}

// initialSpotifyPositionCmd agenda o primeiro poll rápido (1s) logo
// após iniciar a reprodução. Serve pra corrigir a posição estimada
// enquanto o Spotify propaga a faixa no Connect.
func initialSpotifyPositionCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return initialSpotifyPositionMsg{}
	})
}

// seekCmd envia um seek (em ms) pro Spotify Connect. O `deltaSec` é
// só metadado pra mensagem de log/erro na UI; a posição absoluta é
// calculada pelo caller (currentSpotfiyPositionMs + deltaMs).
func seekCmd(client *api.Client, positionMs int, deltaSec int) tea.Cmd {
	return func() tea.Msg {
		err := client.Seek(positionMs)
		return seekResultMsg{deltaSec: deltaSec, err: err}
	}
}

// isConnectURI retorna true pra URIs que tocam via Spotify Connect
// (formato "spotify:track:<id>"). Demais (soundcloud:/youtube:) usam
// player local.
func isConnectURI(uri string) bool {
	return len(uri) > 8 && uri[:8] == "spotify:"
}
