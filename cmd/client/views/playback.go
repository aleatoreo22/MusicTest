package views

import (
	"music-unified-backend/cmd/client/api"

	tea "github.com/charmbracelet/bubbletea"
)

type playbackState int

const (
	stateStopped playbackState = iota
	statePlaying
	statePaused
)

type playResultMsg struct {
	uri      string
	newState playbackState
	err      error
}

type simpleResultMsg struct {
	op  string
	err error
}

type playlistsLoadedMsg struct {
	playlists []api.Playlist
	err       error
}

func playTrackCmd(client *api.Client, uri string) tea.Cmd {
	return func() tea.Msg {
		err := client.PlayTrack(uri, "")
		if err != nil {
			return playResultMsg{err: err}
		}
		return playResultMsg{uri: uri, newState: statePlaying}
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
