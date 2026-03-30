package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ssm/internal/config"
)

type Action int

const (
	ActionNone Action = iota
	ActionConnect
	ActionAdd
	ActionKeys
	ActionSettings
)

type ListModel struct {
	vault       *config.Vault
	allConns    []config.Connection
	connections []config.Connection
	cursor      int
	Action      Action
	Selected    *config.Connection
	deleting    int
	searching   bool
	search      string
	masterPass  string
	width       int
	height      int
}

func NewListModel(v *config.Vault, masterPass string) ListModel {
	return ListModel{
		vault:       v,
		allConns:    v.Connections,
		connections: v.Connections,
		deleting:    -1,
		masterPass:  masterPass,
	}
}

func (m *ListModel) applyFilter() {
	if m.search == "" {
		m.connections = m.allConns
	} else {
		query := strings.ToLower(m.search)
		var filtered []config.Connection
		for _, c := range m.allConns {
			if strings.Contains(strings.ToLower(c.Name), query) ||
				strings.Contains(strings.ToLower(c.Host), query) ||
				strings.Contains(strings.ToLower(c.User), query) {
				filtered = append(filtered, c)
			}
		}
		m.connections = filtered
	}
	if m.cursor >= len(m.connections) {
		m.cursor = max(0, len(m.connections)-1)
	}
}

func (m ListModel) Init() tea.Cmd { return nil }

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if m.deleting >= 0 {
			return m.handleDelete(msg)
		}
		if m.searching {
			return m.handleSearch(msg)
		}
		return m.handleNormal(msg)
	}
	return m, nil
}

func (m ListModel) handleNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	vim := config.LoadSettings().VimKeys
	key := msg.String()

	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down":
		if m.cursor < len(m.connections)-1 {
			m.cursor++
		}
	case "k":
		if vim {
			if m.cursor > 0 {
				m.cursor--
			}
		} else {
			m.Action = ActionKeys
			return m, tea.Quit
		}
	case "j":
		if vim && m.cursor < len(m.connections)-1 {
			m.cursor++
		}
	case "K":
		if vim {
			m.Action = ActionKeys
			return m, tea.Quit
		}
	case "enter":
		if len(m.connections) > 0 {
			m.Action = ActionConnect
			m.Selected = &m.connections[m.cursor]
			return m, tea.Quit
		}
	case "a":
		m.Action = ActionAdd
		return m, tea.Quit
	case "d":
		if len(m.connections) > 0 {
			m.deleting = m.cursor
		}
	case "/":
		m.searching = true
		m.search = ""
	case "s":
		m.Action = ActionSettings
		return m, tea.Quit
	}
	return m, nil
}

func (m ListModel) handleSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "escape":
		m.searching = false
		m.search = ""
		m.applyFilter()
	case "enter":
		m.searching = false
		if len(m.connections) > 0 {
			m.Action = ActionConnect
			m.Selected = &m.connections[m.cursor]
			return m, tea.Quit
		}
	case "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down":
		if m.cursor < len(m.connections)-1 {
			m.cursor++
		}
	case "backspace":
		if len(m.search) > 0 {
			m.search = m.search[:len(m.search)-1]
			m.applyFilter()
		} else {
			m.searching = false
			m.applyFilter()
		}
	default:
		if len(msg.Runes) > 0 {
			m.search += string(msg.Runes)
			m.applyFilter()
		}
	}
	return m, nil
}

func (m ListModel) handleDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		target := m.connections[m.deleting]
		for i, c := range m.allConns {
			if c.Name == target.Name {
				m.allConns = append(m.allConns[:i], m.allConns[i+1:]...)
				break
			}
		}
		m.vault.Connections = m.allConns
		config.Save(m.vault, m.masterPass)
		m.deleting = -1
		m.applyFilter()
	case "n", "escape":
		m.deleting = -1
	}
	return m, nil
}

func (m ListModel) View() string {
	var content strings.Builder

	header := titleStyle.Render("  ssm") + "  " + subtitleStyle.Render("SSH Manager")
	count := subtitleStyle.Render(fmt.Sprintf("%d connections", len(m.allConns)))
	headerLine := header + strings.Repeat(" ", max(0, 50-lipgloss.Width(header)-lipgloss.Width(count))) + count
	content.WriteString(headerLine)
	content.WriteString("\n")

	if m.searching {
		searchContent := "/ " + m.search + fieldCursor.Render("_")
		content.WriteString("\n")
		content.WriteString(searchBox.Render(searchContent))
		content.WriteString("\n")
	}

	content.WriteString("\n")

	if len(m.connections) == 0 {
		if m.searching {
			content.WriteString(dimRow.Render("  No results matching \"" + m.search + "\""))
		} else {
			content.WriteString(dimRow.Render("  No saved connections yet."))
			content.WriteString("\n")
			content.WriteString(dimRow.Render("  Press ") + footerKey.Render("a") + dimRow.Render(" to add your first connection."))
		}
	} else {
		for i, c := range m.connections {
			selected := i == m.cursor

			icon := "  "
			var nameStr, detailStr, portStr string

			if c.Password != "" {
				portStr = " "
			} else if c.KeyName != "" {
				portStr = " "
			} else {
				portStr = " "
			}

			port := ""
			if c.Port != 0 && c.Port != 22 {
				port = fmt.Sprintf(":%d", c.Port)
			}
			detail := fmt.Sprintf("%s@%s%s", c.User, c.Host, port)

			if selected {
				icon = selectedRow.Render(" > ")
				nameStr = selectedRow.Render(fmt.Sprintf("%-18s", c.Name))
				detailStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#A78BFA")).Render(detail)
			} else {
				icon = "   "
				nameStr = normalRow.Render(fmt.Sprintf("%-18s", c.Name))
				detailStr = dimRow.Render(detail)
			}

			line := icon + portStr + nameStr + "  " + detailStr
			if selected {
				line = lipgloss.NewStyle().
					Background(lipgloss.Color("#27272A")).
					Width(54).
					Render(line)
			}

			content.WriteString(line)
			content.WriteString("\n")
		}
	}

	if m.deleting >= 0 {
		name := m.connections[m.deleting].Name
		content.WriteString("\n")
		content.WriteString(confirmStyle.Render(fmt.Sprintf("  Delete \"%s\"? ", name)))
		content.WriteString(footerKey.Render("y") + dimRow.Render("es") + dimRow.Render(" / ") + footerKey.Render("n") + dimRow.Render("o"))
	}

	if m.searching {
		content.WriteString(footerBar(
			footerItem("enter", "connect"),
			footerItem("↑↓", "navigate"),
			footerItem("esc", "cancel"),
		))
	} else if m.deleting < 0 {
		keysKey := "K"
		if !config.LoadSettings().VimKeys {
			keysKey = "k"
		}
		content.WriteString(footerBar(
			footerItem("enter", "connect"),
			footerItem("a", "add"),
			footerItem("d", "delete"),
			footerItem("/", "search"),
			footerItem(keysKey, "keys"),
			footerItem("s", "settings"),
			footerItem("q", "quit"),
		))
	}

	box := boxStyle.Width(58).Render(content.String())

	if m.width > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
	}
	return box
}
