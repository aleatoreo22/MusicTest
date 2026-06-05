package views

import (
	"fmt"
	"music-unified-backend/cmd/client/api"
	"music-unified-backend/pkg/utils"
	"os"
	"strings"

	"github.com/aymanbagabas/go-osc52/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SpotifyLinkView struct {
	Client   *api.Client
	IsLinked bool
	AuthURL  string
	Error    string
	Back     bool
	width    int
}

func NewSpotifyLinkView(client *api.Client) *SpotifyLinkView {
	v := &SpotifyLinkView{
		Client: client,
	}
	if linked, err := client.GetSpotifyStatus(); err == nil {
		v.IsLinked = linked
	}
	return v
}

type browserOpenedMsg struct {
	err error
}

type urlCopiedMsg struct{}

func openBrowserCmd(url string) tea.Cmd {
	return func() tea.Msg {
		return browserOpenedMsg{err: utils.OpenBrowser(url)}
	}
}

func copyURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		_, _ = osc52.New(url).WriteTo(os.Stderr)
		return urlCopiedMsg{}
	}
}

func (v *SpotifyLinkView) Update(msg tea.Msg) (*SpotifyLinkView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			v.Back = true
		case "enter":
			if !v.IsLinked {
				url, err := v.Client.GetSpotifyAuthURL()
				if err != nil {
					v.Error = err.Error()
					return v, nil
				}
				v.AuthURL = url
				v.Error = ""
				return v, openBrowserCmd(url)
			}
		case "o":
			if v.AuthURL != "" {
				return v, openBrowserCmd(v.AuthURL)
			}
		case "c":
			if v.AuthURL != "" {
				return v, copyURLCmd(v.AuthURL)
			}
		}
	case browserOpenedMsg:
		if msg.err != nil {
			v.Error = msg.err.Error()
		}
	case urlCopiedMsg:
		v.Error = "URL copiada para a área de transferência ✓"
	}
	return v, nil
}

func (v *SpotifyLinkView) View() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5")).
		Render("🎵 Link Spotify Account")

	if v.IsLinked {
		content := lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Render("✓ Spotify account linked successfully!")
		help := helpStyle.Render("Esc: Back | Ctrl+C: Quit")
		return lipgloss.JoinVertical(lipgloss.Left, title, "", content, "", help)
	}

	var body strings.Builder

	if v.Error != "" {
		body.WriteString(errorStyle.Render(fmt.Sprintf("❌ %s", v.Error)))
		body.WriteString("\n\n")
	}

	if v.AuthURL == "" {
		body.WriteString("Press [Enter] to authorize with Spotify.\n")
		body.WriteString("Your browser will open automatically.")
	} else {
		body.WriteString("Open this link in your browser (Ctrl+click):\n\n")
		body.WriteString(v.renderURL())
		body.WriteString("\n\n")
		body.WriteString(dimStyle.Render("Already authorized? Press [Enter] to check status again."))
	}

	help := helpStyle.Render("[Enter] Authorize / Check  [o] Open in browser  [c] Copy URL  [Esc] Back")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", body.String(), "", help)
}

func (v *SpotifyLinkView) renderURL() string {
	link := utils.Hyperlink(v.AuthURL, v.AuthURL)
	if v.width > 20 {
		return lipgloss.NewStyle().
			Width(v.width - 4).
			Foreground(lipgloss.Color("6")).
			Render(link)
	}
	return urlStyle.Render(link)
}
