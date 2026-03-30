package tui

import "github.com/charmbracelet/lipgloss"

var (
	accent = lipgloss.Color("#8B5CF6")
	white  = lipgloss.Color("#E4E4E7")
	dim    = lipgloss.Color("#71717A")
	dimmer = lipgloss.Color("#52525B")
	red    = lipgloss.Color("#EF4444")

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3F3F46")).
			Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(dim)

	selectedRow = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent)

	normalRow = lipgloss.NewStyle().
			Foreground(white)

	dimRow = lipgloss.NewStyle().
		Foreground(dim)

	footerKey = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true)

	footerLabel = lipgloss.NewStyle().
			Foreground(dimmer)

	footerSep = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3F3F46"))

	searchBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Padding(0, 1).
			Foreground(white)

	confirmStyle = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)

	fieldLabel = lipgloss.NewStyle().
			Width(16).
			Foreground(dim)

	fieldLabelActive = lipgloss.NewStyle().
				Width(16).
				Foreground(accent).
				Bold(true)

	fieldValue = lipgloss.NewStyle().
			Foreground(white)

	fieldPlaceholder = lipgloss.NewStyle().
				Foreground(dimmer)

	fieldCursor = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true)

	fieldInputBox = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(accent).
			PaddingBottom(0)

	fieldInputBoxInactive = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(lipgloss.Color("#3F3F46")).
				PaddingBottom(0)
)

func footerItem(key, label string) string {
	return footerKey.Render(key) + " " + footerLabel.Render(label)
}

func footerBar(items ...string) string {
	sep := footerSep.Render(" | ")
	return "\n" + lipgloss.JoinHorizontal(lipgloss.Center, joinWith(items, sep))
}

func joinWith(items []string, sep string) string {
	result := ""
	for i, item := range items {
		if i > 0 {
			result += sep
		}
		result += item
	}
	return result
}
