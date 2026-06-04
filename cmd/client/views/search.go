package views

import (
	"fmt"
	"music-unified-backend/cmd/client/api"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SearchView struct {
	Client     *api.Client
	Query      string
	Results    []api.Track
	Selected   int
	Back       bool
	Mode       int // 0: input, 1: results, 2: playlist picker, 3: new playlist name
	State      playbackState
	PlayingURI string
	Pending    bool
	PlayerMsg  string
	PlayerErr  string
	width      int

	// Playlist picker
	Playlists          []api.Playlist
	PlaylistSelected   int
	PlaylistLoading    bool
	NewPlaylistName    string
	TrackToAdd         api.Track
	AddResultPending   bool
}

func NewSearchView(client *api.Client) *SearchView {
	return &SearchView{
		Client: client,
		Mode:   0,
		State:  stateStopped,
	}
}

type addTrackResultMsg struct {
	playlistName string
	err          error
}

type createPlaylistResultMsg struct {
	name string
	err  error
}

func addTrackCmd(client *api.Client, playlistID uint, track api.Track, name string) tea.Cmd {
	return func() tea.Msg {
		err := client.AddTrackToPlaylist(playlistID, track)
		return addTrackResultMsg{playlistName: name, err: err}
	}
}

func createPlaylistAndAddCmd(client *api.Client, name string, track api.Track) tea.Cmd {
	return func() tea.Msg {
		pl, err := client.CreatePlaylist(name)
		if err != nil {
			return createPlaylistResultMsg{name: name, err: err}
		}
		if err := client.AddTrackToPlaylist(pl.ID, track); err != nil {
			return addTrackResultMsg{playlistName: pl.Name, err: err}
		}
		return addTrackResultMsg{playlistName: pl.Name}
	}
}

func (v *SearchView) Update(msg tea.Msg, client *api.Client) (*SearchView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
	case tea.KeyMsg:
		switch v.Mode {
		case 0:
			return v.updateInputMode(msg)
		case 1:
			return v.updateResultsMode(msg)
		case 2:
			return v.updatePlaylistPickerMode(msg)
		case 3:
			return v.updateNewPlaylistMode(msg)
		}
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
	case playlistsLoadedMsg:
		v.PlaylistLoading = false
		if msg.err != nil {
			v.PlayerErr = "Erro ao carregar playlists: " + msg.err.Error()
			v.Mode = 1
			return v, nil
		}
		v.Playlists = msg.playlists
		v.PlaylistSelected = 0
	case addTrackResultMsg:
		v.AddResultPending = false
		if msg.err != nil {
			v.PlayerErr = "Erro ao adicionar: " + msg.err.Error()
			return v, nil
		}
		v.PlayerMsg = "✓ Adicionada em \"" + msg.playlistName + "\""
		v.Mode = 1
		v.NewPlaylistName = ""
	case createPlaylistResultMsg:
		v.AddResultPending = false
		if msg.err != nil {
			v.PlayerErr = "Erro ao criar playlist: " + msg.err.Error()
			return v, nil
		}
	}
	return v, nil
}

func (v *SearchView) updateInputMode(msg tea.KeyMsg) (*SearchView, tea.Cmd) {
	switch msg.String() {
	case "esc":
		v.Back = true
		v.Mode = 0
		v.Query = ""
	case "enter":
		if v.Query != "" {
			results, err := v.Client.SearchSpotify(v.Query)
			if err != nil {
				v.PlayerErr = err.Error()
				v.Mode = 1
			} else {
				v.Results = results
				v.Mode = 1
				v.Selected = 0
				v.PlayerErr = ""
				v.PlayerMsg = ""
			}
		}
	case "backspace":
		if len(v.Query) > 0 {
			v.Query = v.Query[:len(v.Query)-1]
		}
	default:
		if len(msg.String()) == 1 {
			v.Query += msg.String()
		}
	}
	return v, nil
}

func (v *SearchView) updateResultsMode(msg tea.KeyMsg) (*SearchView, tea.Cmd) {
	switch msg.String() {
	case "esc":
		v.Back = true
		v.Mode = 0
		v.Query = ""
	case "up", "k":
		if v.Selected > 0 {
			v.Selected--
		}
	case "down", "j":
		if v.Selected < len(v.Results)-1 {
			v.Selected++
		}
	case "ctrl+p":
		if v.Pending {
			return v, nil
		}
		switch v.State {
		case statePaused:
			v.PlayerErr = ""
			v.Pending = true
			return v, resumeCmd(v.Client)
		default:
			if len(v.Results) == 0 {
				return v, nil
			}
			track := v.Results[v.Selected]
			if track.SourceID == "" {
				v.PlayerErr = "Faixa sem ID do Spotify"
				return v, nil
			}
			uri := "spotify:track:" + track.SourceID
			v.PlayerErr = ""
			v.Pending = true
				return v, playTrackCmd(v.Client, uri)
		}
	case "ctrl+s":
		if v.State != statePlaying || v.Pending {
			return v, nil
		}
		v.PlayerErr = ""
		v.Pending = true
		return v, pauseCmd(v.Client)
	case "ctrl+n":
		if v.State == stateStopped || v.Pending {
			return v, nil
		}
		v.PlayerErr = ""
		v.Pending = true
		return v, nextCmd(v.Client)
	case "ctrl+b":
		if v.State == stateStopped || v.Pending {
			return v, nil
		}
		v.PlayerErr = ""
		v.Pending = true
		return v, previousCmd(v.Client)
	case "ctrl+a":
		if len(v.Results) == 0 || v.AddResultPending {
			return v, nil
		}
		v.TrackToAdd = v.Results[v.Selected]
		v.PlaylistLoading = true
		v.PlayerErr = ""
		v.PlayerMsg = ""
		v.Mode = 2
		return v, loadPlaylistsCmd(v.Client)
	}
	return v, nil
}

func (v *SearchView) updatePlaylistPickerMode(msg tea.KeyMsg) (*SearchView, tea.Cmd) {
	switch msg.String() {
	case "esc":
		v.Mode = 1
		v.PlayerErr = ""
	case "up", "k":
		if v.PlaylistSelected > 0 {
			v.PlaylistSelected--
		}
	case "down", "j":
		// +1 para a opção "+ Nova playlist"
		max := len(v.Playlists)
		if v.PlaylistSelected < max {
			v.PlaylistSelected++
		}
	case "enter":
		if v.AddResultPending {
			return v, nil
		}
		// index 0 = "+ Nova playlist"
		if v.PlaylistSelected == 0 {
			v.Mode = 3
			v.NewPlaylistName = ""
			return v, nil
		}
		// index >= 1 = playlist real
		idx := v.PlaylistSelected - 1
		if idx >= len(v.Playlists) {
			return v, nil
		}
		pl := v.Playlists[idx]
		v.AddResultPending = true
		return v, addTrackCmd(v.Client, pl.ID, v.TrackToAdd, pl.Name)
	}
	return v, nil
}

func (v *SearchView) updateNewPlaylistMode(msg tea.KeyMsg) (*SearchView, tea.Cmd) {
	switch msg.String() {
	case "esc":
		v.Mode = 2
		v.NewPlaylistName = ""
	case "enter":
		name := v.NewPlaylistName
		if name == "" || v.AddResultPending {
			return v, nil
		}
		v.AddResultPending = true
		return v, createPlaylistAndAddCmd(v.Client, name, v.TrackToAdd)
	case "backspace":
		if len(v.NewPlaylistName) > 0 {
			v.NewPlaylistName = v.NewPlaylistName[:len(v.NewPlaylistName)-1]
		}
	default:
		if len(msg.String()) == 1 {
			v.NewPlaylistName += msg.String()
		}
	}
	return v, nil
}

func (v *SearchView) View() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5")).
		Render("🔍 Search Spotify")

	switch v.Mode {
	case 0:
		return v.viewInput(title)
	case 1:
		return v.viewResults(title)
	case 2:
		return v.viewPlaylistPicker(title)
	case 3:
		return v.viewNewPlaylist(title)
	}
	return ""
}

func (v *SearchView) viewInput(title string) string {
	input := lipgloss.NewStyle().
		Background(lipgloss.Color("4")).
		Foreground(lipgloss.Color("15")).
		Padding(0, 1).
		Render("Search: " + v.Query + "_")
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		input,
		"",
		"Press Enter to search | Esc to go back",
	)
}

func (v *SearchView) viewResults(title string) string {
	var results string
	if v.PlayerErr != "" {
		results = errorStyle.Render("❌ " + v.PlayerErr)
	} else if len(v.Results) == 0 {
		results = "No results found"
	} else {
		for i, track := range v.Results {
			style := lipgloss.NewStyle()
			if i == v.Selected {
				style = lipgloss.NewStyle().
					Background(lipgloss.Color("4")).
					Foreground(lipgloss.Color("15")).
					Padding(0, 1)
			}
			uri := "spotify:track:" + track.SourceID
			marker := "  "
			if v.State != stateStopped && uri == v.PlayingURI {
				switch v.State {
				case statePlaying:
					marker = "▶ "
				case statePaused:
					marker = "⏸ "
				}
			}
			results += style.Render(fmt.Sprintf("%s%s - %s\n", marker, track.Title, track.Artist))
		}
	}

	var status string
	switch {
	case v.PlayerErr != "":
		status = errorStyle.Render("⚠ " + v.PlayerErr)
	case v.PlayerMsg != "":
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(v.PlayerMsg)
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
		results,
		"",
		status,
		"",
		"[Ctrl+P] Play  [Ctrl+S] Pause  [Ctrl+N] Next  [Ctrl+B] Previous  [Ctrl+A] Add to playlist  ↑/↓: Nav  Esc: Back",
	)
}

func (v *SearchView) viewPlaylistPicker(title string) string {
	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render(fmt.Sprintf("Adicionar \"%s - %s\" em:", v.TrackToAdd.Title, v.TrackToAdd.Artist))

	if v.PlaylistLoading {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			subtitle,
			"",
			dimStyle.Render("⏳ Carregando playlists..."),
		)
	}

	items := []string{"➕ Nova playlist..."}
	for _, pl := range v.Playlists {
		items = append(items, fmt.Sprintf("%s (%d tracks)", pl.Name, len(pl.Tracks)))
	}

	var list string
	for i, item := range items {
		style := lipgloss.NewStyle()
		if i == v.PlaylistSelected {
			style = lipgloss.NewStyle().
				Background(lipgloss.Color("4")).
				Foreground(lipgloss.Color("15")).
				Padding(0, 1)
		}
		list += style.Render(item) + "\n"
	}

	status := ""
	if v.PlayerErr != "" {
		status = errorStyle.Render("⚠ " + v.PlayerErr)
	} else if v.AddResultPending {
		status = dimStyle.Render("⏳ Salvando...")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		subtitle,
		"",
		list,
		"",
		status,
		"",
		"↑/↓: Nav | Enter: Select | Esc: Back",
	)
}

func (v *SearchView) viewNewPlaylist(title string) string {
	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render(fmt.Sprintf("Criar playlist e adicionar \"%s\":", v.TrackToAdd.Title))

	input := lipgloss.NewStyle().
		Background(lipgloss.Color("4")).
		Foreground(lipgloss.Color("15")).
		Padding(0, 1).
		Render("Nome: " + v.NewPlaylistName + "_")

	status := ""
	if v.PlayerErr != "" {
		status = errorStyle.Render("⚠ " + v.PlayerErr)
	} else if v.AddResultPending {
		status = dimStyle.Render("⏳ Criando e adicionando...")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		subtitle,
		"",
		input,
		"",
		status,
		"",
		"Enter: Create & Add | Esc: Back",
	)
}
