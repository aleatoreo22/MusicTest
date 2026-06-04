package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HomeView struct {
	Selected int
	SwitchTo string
}

func NewHomeView() *HomeView {
	return &HomeView{Selected: 0}
}

func (v *HomeView) Update(msg tea.Msg) (*HomeView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			v.Selected = (v.Selected - 1 + 4) % 4
		case "down", "j":
			v.Selected = (v.Selected + 1) % 4
		case "enter":
			switch v.Selected {
			case 0:
				v.SwitchTo = "search"
			case 1:
				v.SwitchTo = "playlist"
			case 2:
				v.SwitchTo = "spotify"
			case 3:
				v.SwitchTo = "auth"
			}
		}
	}
	return v, nil
}

func (v *HomeView) View() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5")).
		Render("🎵 Music Client - Home")

	items := []string{"🔍 Search Spotify", "📋 My Playlists", "🔗 Link Spotify", "🚪 Logout"}
	var menu string
	for i, item := range items {
		style := lipgloss.NewStyle()
		if i == v.Selected {
			style = lipgloss.NewStyle().
				Background(lipgloss.Color("4")).
				Foreground(lipgloss.Color("15")).
				Padding(0, 1)
		}
		menu += style.Render(item) + "\n"
	}

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("↑/↓ or j/k: Navigate | Enter: Select | Ctrl+C: Quit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		menu,
		help,
	)
}
