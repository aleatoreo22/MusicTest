package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HomeView struct {
	Selected  int
	SwitchTo  string
	Providers []string
	providers []string
}

func NewHomeView() *HomeView {
	return &HomeView{
		Selected: 0,
		Providers: []string{},
	}
}

func (v *HomeView) SetProviders(p []string) {
	v.providers = p
	v.Providers = p
}

func (v *HomeView) hasProvider(name string) bool {
	for _, p := range v.providers {
		if p == name {
			return true
		}
	}
	return false
}

func (v *HomeView) buildItems() []string {
	items := []string{"🔍 Search"}
	if v.hasProvider("spotify") {
		items = append(items, "🔗 Link Spotify")
	}
	items = append(items, "📋 My Playlists")
	items = append(items, "🚪 Logout")
	return items
}

func (v *HomeView) itemAction(i int) string {
	if len(v.providers) == 0 {
		return ""
	}
	items := v.buildItems()
	if i < 0 || i >= len(items) {
		return ""
	}
	switch items[i] {
	case "🔍 Search":
		return "search"
	case "📋 My Playlists":
		return "playlist"
	case "🔗 Link Spotify":
		return "spotify"
	case "🚪 Logout":
		return "auth"
	}
	return ""
}

func (v *HomeView) Update(msg tea.Msg) (*HomeView, tea.Cmd) {
	if len(v.providers) == 0 {
		return v, nil
	}
	items := v.buildItems()
	n := len(items)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			v.Selected = (v.Selected - 1 + n) % n
		case "down", "j":
			v.Selected = (v.Selected + 1) % n
		case "enter":
			v.SwitchTo = v.itemAction(v.Selected)
		}
	}
	return v, nil
}

func (v *HomeView) View() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5")).
		Render("🎵 Music Client - Home")

	items := v.buildItems()
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

	footer := fmt.Sprintf("Serviços disponíveis: %v", v.providers)
	help := dimStyle.Render("↑/↓ or j/k: Navigate | Enter: Select | Ctrl+C: Quit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		menu,
		footer,
		"",
		help,
	)
}
