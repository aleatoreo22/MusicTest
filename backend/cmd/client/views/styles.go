package views

import "github.com/charmbracelet/lipgloss"

var (
	urlStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).
			Underline(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
)
