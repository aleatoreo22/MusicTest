package main

import (
	"music-unified-backend/cmd/client/api"
	"music-unified-backend/cmd/client/views"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	state        int
	apiClient    *api.Client
	authView     *views.AuthView
	homeView     *views.HomeView
	searchView   *views.SearchView
	playlistView *views.PlaylistView
	spotifyView  *views.SpotifyLinkView
}

const (
	authState = iota
	homeState
	searchState
	playlistState
	spotifyState
)

func NewModel() *Model {
	apiClient := api.NewClient("http://localhost:8080")
	return &Model{
		state:        authState,
		apiClient:    apiClient,
		authView:     views.NewAuthView(),
		homeView:     views.NewHomeView(),
		searchView:   views.NewSearchView(apiClient),
		playlistView: views.NewPlaylistView(apiClient),
		spotifyView:  views.NewSpotifyLinkView(apiClient),
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	switch m.state {
	case authState:
		var cmd tea.Cmd
		m.authView, cmd = m.authView.Update(msg, m.apiClient)
		if m.authView.IsAuthenticated {
			m.state = homeState
		}
		return m, cmd
	case homeState:
		var cmd tea.Cmd
		m.homeView, cmd = m.homeView.Update(msg)
		if m.homeView.SwitchTo != "" {
			switch m.homeView.SwitchTo {
			case "search":
				m.state = searchState
				m.searchView = views.NewSearchView(m.apiClient) // Passa client com token
			case "playlist":
				m.state = playlistState
				// Carrega playlists quando abre
				m.playlistView = views.NewPlaylistView(m.apiClient)
				if playlists, err := m.apiClient.GetPlaylists(); err == nil {
					m.playlistView.Playlists = playlists
				}
			case "spotify":
				m.state = spotifyState
				m.spotifyView = views.NewSpotifyLinkView(m.apiClient)
			case "auth":
				m.state = authState
				m.authView = views.NewAuthView()
			}
			m.homeView.SwitchTo = ""
		}
		return m, cmd
	case searchState:
		var cmd tea.Cmd
		m.searchView, cmd = m.searchView.Update(msg, m.apiClient)
		if m.searchView.Back {
			m.state = homeState
			m.searchView.Back = false
		}
		return m, cmd
	case playlistState:
		var cmd tea.Cmd
		m.playlistView, cmd = m.playlistView.Update(msg, m.apiClient)
		if m.playlistView.Back {
			m.state = homeState
			m.playlistView.Back = false
		}
		return m, cmd
	case spotifyState:
		var cmd tea.Cmd
		m.spotifyView, cmd = m.spotifyView.Update(msg)
		if m.spotifyView.Back {
			m.state = homeState
			m.spotifyView.Back = false
		}
		return m, cmd
	}

	return m, nil
}

func (m *Model) View() string {
	switch m.state {
	case authState:
		return m.authView.View()
	case homeState:
		return m.homeView.View()
	case searchState:
		return m.searchView.View()
	case playlistState:
		return m.playlistView.View()
	case spotifyState:
		return m.spotifyView.View()
	}
	return ""
}
