package views

import (
	"music-unified-backend/cmd/client/api"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AuthView struct {
	Email           string
	Password        string
	IsLogin         bool
	Focus           int
	IsAuthenticated bool
	Error           string
}

func NewAuthView() *AuthView {
	return &AuthView{
		IsLogin: true,
		Focus:   0,
	}
}

func (v *AuthView) Update(msg tea.Msg, client *api.Client) (*AuthView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			v.Focus = (v.Focus + 1) % 3
		case "shift+tab":
			v.Focus = (v.Focus - 1 + 3) % 3
		case "ctrl+l":
			v.IsLogin = true
		case "ctrl+r":
			v.IsLogin = false
		case "enter":
			if v.Focus == 2 {
				if v.IsLogin {
					resp, err := client.Login(v.Email, v.Password)
					if err != nil {
						v.Error = err.Error()
					} else {
						v.IsAuthenticated = true
						client.SetToken(resp.Token)
					}
				} else {
					resp, err := client.Register(v.Email, v.Password)
					if err != nil {
						v.Error = err.Error()
					} else {
						v.IsAuthenticated = true
						client.SetToken(resp.Token)
					}
				}
			}
		case "backspace":
			if v.Focus == 0 && len(v.Email) > 0 {
				v.Email = v.Email[:len(v.Email)-1]
			} else if v.Focus == 1 && len(v.Password) > 0 {
				v.Password = v.Password[:len(v.Password)-1]
			}
		default:
			if len(msg.String()) == 1 {
				if v.Focus == 0 {
					v.Email += msg.String()
				} else if v.Focus == 1 {
					v.Password += msg.String()
				}
			}
		}
	}
	return v, nil
}

func (v *AuthView) View() string {
	title := "🎵 Music Client"
	subtitle := "Login or Register"

	emailStyle := lipgloss.NewStyle()
	if v.Focus == 0 {
		emailStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("4")).
			Foreground(lipgloss.Color("15")).
			Padding(0, 1)
	}
	emailInput := emailStyle.Render("Email: " + v.Email)

	passStyle := lipgloss.NewStyle()
	if v.Focus == 1 {
		passStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("4")).
			Foreground(lipgloss.Color("15")).
			Padding(0, 1)
	}
	passInput := passStyle.Render("Password: " + strings.Repeat("•", len(v.Password)))

	submitStyle := lipgloss.NewStyle()
	if v.Focus == 2 {
		submitStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("2")).
			Foreground(lipgloss.Color("15")).
			Padding(0, 1)
	}
	mode := "Login"
	if !v.IsLogin {
		mode = "Register"
	}
	submitBtn := submitStyle.Render("[" + mode + "]")

	errorMsg := ""
	if v.Error != "" {
		errorMsg = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Render("Error: " + v.Error)
	}

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("Tab: Navigate | Ctrl+L: Login | Ctrl+R: Register | Ctrl+C: Quit")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")).Render(title),
		subtitle,
		"",
		emailInput,
		passInput,
		submitBtn,
		errorMsg,
		"",
		help,
	)
}
