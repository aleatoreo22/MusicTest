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
	focus           int
	IsAuthenticated bool
	Error           string
	Pending         bool
	width           int
	height          int
}

func NewAuthView() *AuthView {
	return &AuthView{
		IsLogin: true,
		focus:   0,
	}
}

const fieldWidth = 44

func (v *AuthView) submit(client *api.Client) {
	if v.Pending {
		return
	}
	if v.Email == "" || v.Password == "" {
		return
	}
	v.Pending = true
	var resp *api.AuthResponse
	var err error
	if v.IsLogin {
		resp, err = client.Login(v.Email, v.Password)
	} else {
		resp, err = client.Register(v.Email, v.Password)
	}
	v.Pending = false
	if err != nil {
		v.Error = err.Error()
		return
	}
	v.Error = ""
	v.IsAuthenticated = true
	client.SetToken(resp.Token)
}

func (v *AuthView) Update(msg tea.Msg, client *api.Client) (*AuthView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			v.focus = (v.focus + 1) % 2
		case "shift+tab":
			v.focus = (v.focus - 1 + 2) % 2
		case "ctrl+l":
			v.IsLogin = true
		case "ctrl+r":
			v.IsLogin = false
		case "enter":
			if v.focus == 0 {
				v.focus = 1
			} else {
				v.submit(client)
			}
		case "backspace":
			if v.focus == 0 && len(v.Email) > 0 {
				v.Email = v.Email[:len(v.Email)-1]
			} else if v.focus == 1 && len(v.Password) > 0 {
				v.Password = v.Password[:len(v.Password)-1]
			}
		default:
			if len(msg.String()) == 1 {
				if v.focus == 0 {
					v.Email += msg.String()
				} else if v.focus == 1 {
					v.Password += msg.String()
				}
			}
		}
	}
	return v, nil
}

func (v *AuthView) renderField(value string, isFocused bool) string {
	display := value
	if isFocused {
		display += "▌"
	}
	runes := []rune(display)
	if len(runes) > fieldWidth {
		runes = runes[:fieldWidth]
		display = string(runes)
	}
	pad := fieldWidth - len(runes)
	padded := display + strings.Repeat(" ", pad)

	borderColor := "8"
	if isFocused {
		borderColor = "5"
	}

	return lipgloss.NewStyle().
		Width(fieldWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1).
		Render(padded)
}

func (v *AuthView) View() string {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5")).
		Render("🎵 Music Client")

	activeBtn := lipgloss.NewStyle().
		Background(lipgloss.Color("5")).
		Foreground(lipgloss.Color("0")).
		Bold(true).
		Padding(0, 2)
	inactiveBtn := dim.Padding(0, 2)

	var modeRow string
	if v.IsLogin {
		modeRow = lipgloss.JoinHorizontal(
			lipgloss.Center,
			activeBtn.Render("Login"),
			"  ",
			inactiveBtn.Render("Register"),
		)
	} else {
		modeRow = lipgloss.JoinHorizontal(
			lipgloss.Center,
			inactiveBtn.Render("Login"),
			"  ",
			activeBtn.Render("Register"),
		)
	}

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Bold(true)
	focusedLabel := labelStyle.Foreground(lipgloss.Color("5"))

	emailLabel := labelStyle.Render("EMAIL")
	passwordLabel := labelStyle.Render("PASSWORD")
	if v.focus == 0 {
		emailLabel = focusedLabel.Render("EMAIL")
	}
	if v.focus == 1 {
		passwordLabel = focusedLabel.Render("PASSWORD")
	}

	maskedPassword := strings.Repeat("•", len(v.Password))

	emailBox := v.renderField(v.Email, v.focus == 0)
	passwordBox := v.renderField(maskedPassword, v.focus == 1)

	pending := ""
	if v.Pending {
		pending = dim.Render("verificando…")
	}

	errorRow := ""
	if v.Error != "" {
		errorRow = errorStyle.Render("✗ " + v.Error)
	}

	help := dim.Render("Tab/Shift+Tab: campo  •  Enter: próximo/submeter  •  Ctrl+L/R: modo  •  Ctrl+C: sair")

	form := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		emailLabel,
		emailBox,
		"",
		passwordLabel,
		passwordBox,
		"",
		modeRow,
		pending,
		errorRow,
		"",
		help,
	)

	w := v.width
	h := v.height
	if w == 0 {
		w = 80
	}
	if h == 0 {
		h = 24
	}
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, form)
}
