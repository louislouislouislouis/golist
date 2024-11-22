package ui

import "github.com/charmbracelet/lipgloss"

var (
	docStyle = lipgloss.
			NewStyle()

	bigTitleStyle = lipgloss.
			NewStyle().
			Align(lipgloss.Center)

	focusStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	unfocusStyle = focusStyle.
			BorderForeground(lipgloss.Color("#bababa"))

	titleStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Padding(1, 2).
			Margin(0, 1)
	statusFocusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("62")).
				Bold(true)
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a3a3a3"))

	orangeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e36e00"))
	blueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0053e3"))
	violetStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4f00e3"))
	selectedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0")).Bold(true)
	unselectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
)
