package views

import (
	"fmt"
	"music-unified-backend/cmd/client/api"
	"music-unified-backend/pkg/utils"

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

	// Tracks view
	TracksSelected int
	State          playbackState
	PlayingURI     string
	PlayerErr      string

	// Local playback (mpv) pra SoundCloud/YouTube na playlist
	LocalPlayer *utils.Player
	LocalURL    string
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
		player, err := utils.Play(msg.stream)
		if err != nil {
			v.PlayerErr = err.Error()
			v.State = stateStopped
			return v, nil
		}
		v.LocalPlayer = player
		v.LocalURL = msg.stream
		v.State = statePlaying
		v.PlayingURI = playingURIFor(msg.track)
		v.PlayerErr = ""
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
		}
	case "down", "j":
		if v.TracksSelected < len(tracks)-1 {
			v.TracksSelected++
		}
	case "ctrl+p":
		if v.Pending {
			return v, nil
		}
		if isConnectSource(tracksOrSource(tracks)) {
			switch v.State {
			case statePaused:
				v.PlayerErr = ""
				v.Pending = true
				return v, resumeCmd(v.Client)
			default:
				if len(tracks) == 0 {
					return v, nil
				}
				track := tracks[v.TracksSelected]
				uri, err := buildPlayableURI(track)
				if err != nil {
					v.PlayerErr = err.Error()
					return v, nil
				}
				v.PlayerErr = ""
				v.Pending = true
				return v, playTrackCmd(v.Client, uri)
			}
		}
		// SoundCloud/YouTube na playlist: resolve URL e toca via mpv local
		if len(tracks) == 0 {
			return v, nil
		}
		track := tracks[v.TracksSelected]
		if track.SourceID == "" {
			v.PlayerErr = "Faixa sem ID"
			return v, nil
		}
		v.PlayerErr = ""
		v.Pending = true
		return v, resolveAndPlayLocalCmd(v.Client, track)
	case "ctrl+s":
		if v.State != statePlaying || v.Pending {
			return v, nil
		}
		if isConnectSource(tracksOrSource(tracks)) {
			v.PlayerErr = ""
			v.Pending = true
			return v, pauseCmd(v.Client)
		}
		// Local: mata o player
		v.stopLocal()
		v.State = statePaused
	case "ctrl+n":
		if !isConnectSource(tracksOrSource(tracks)) {
			return v, nil
		}
		if v.State == stateStopped || v.Pending {
			return v, nil
		}
		v.PlayerErr = ""
		v.Pending = true
		return v, nextCmd(v.Client)
	case "ctrl+b":
		if !isConnectSource(tracksOrSource(tracks)) {
			return v, nil
		}
		if v.State == stateStopped || v.Pending {
			return v, nil
		}
		v.PlayerErr = ""
		v.Pending = true
		return v, previousCmd(v.Client)
	}
	return v, nil
}

// tracksOrSource retorna a source da faixa atualmente selecionada, ou ""
// se a lista está vazia. Usado pra rotear play/pause/next entre Connect e local.
func tracksOrSource(tracks []api.Track) string {
	if len(tracks) == 0 {
		return ""
	}
	return tracks[0].Source
}

func (v *PlaylistView) stopLocal() {
	if v.LocalPlayer != nil {
		_ = v.LocalPlayer.Stop()
		v.LocalPlayer = nil
	}
	v.LocalURL = ""
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

func (v *PlaylistView) viewTracks(title string) string {
	if v.Selected < 0 || v.Selected >= len(v.Playlists) {
		return title
	}
	pl := v.Playlists[v.Selected]
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("6")).
		Render(fmt.Sprintf("▶ %s", pl.Name))

	tracks := pl.Tracks
	var list string
	if len(tracks) == 0 {
		list = dimStyle.Render("(vazia)")
	} else {
		for i, track := range tracks {
			style := lipgloss.NewStyle()
			if i == v.TracksSelected {
				style = lipgloss.NewStyle().
					Background(lipgloss.Color("4")).
					Foreground(lipgloss.Color("15")).
					Padding(0, 1)
			}
			uri := ""
			if track.Source == "spotify" {
				uri = "spotify:track:" + track.SourceID
			}
			marker := "  "
			if v.State != stateStopped && uri != "" && uri == v.PlayingURI {
				switch v.State {
				case statePlaying:
					marker = "▶ "
				case statePaused:
					marker = "⏸ "
				}
			}
			sourceTag := "[" + track.Source + "]"
			list += style.Render(fmt.Sprintf("%s%s %s - %s\n", marker, sourceTag, track.Title, track.Artist))
		}
	}

	var status string
	switch {
	case v.PlayerErr != "":
		status = errorStyle.Render("⚠ " + v.PlayerErr)
	case v.Pending:
		status = dimStyle.Render("⏳ Enviando para o Spotify...")
	case v.State == statePlaying:
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("▶ Tocando no Spotify")
	case v.State == statePaused:
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("⏸ Pausado")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		header,
		"",
		list,
		"",
		status,
		"",
		"[Ctrl+P] Play  [Ctrl+S] Pause  [Ctrl+N] Next  [Ctrl+B] Previous  ↑/↓: Nav  Esc: Back",
	)
}
