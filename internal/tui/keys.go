package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ssm/internal/config"
)

type KeyAction int

const (
	KeyActionNone KeyAction = iota
	KeyActionAdd
)

type KeysModel struct {
	vault      *config.Vault
	cursor     int
	Action     KeyAction
	deleting   int
	masterPass string
	width      int
	height     int
}

func NewKeysModel(v *config.Vault, masterPass string) KeysModel {
	return KeysModel{
		vault:      v,
		deleting:   -1,
		masterPass: masterPass,
	}
}

func (m KeysModel) Init() tea.Cmd { return nil }

func (m KeysModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if m.deleting >= 0 {
			return m.handleDelete(msg)
		}
		return m.handleNormal(msg)
	}
	return m, nil
}

func (m KeysModel) handleNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	vim := config.LoadSettings().VimKeys
	switch msg.String() {
	case "q", "ctrl+c", "esc":
		return m, tea.Quit
	case "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down":
		if m.cursor < len(m.vault.Keys)-1 {
			m.cursor++
		}
	case "k":
		if vim && m.cursor > 0 {
			m.cursor--
		}
	case "j":
		if vim && m.cursor < len(m.vault.Keys)-1 {
			m.cursor++
		}
	case "a":
		m.Action = KeyActionAdd
		return m, tea.Quit
	case "d":
		if len(m.vault.Keys) > 0 {
			m.deleting = m.cursor
		}
	}
	return m, nil
}

func (m KeysModel) handleDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		m.vault.Keys = append(m.vault.Keys[:m.deleting], m.vault.Keys[m.deleting+1:]...)
		config.Save(m.vault, m.masterPass)
		m.deleting = -1
		if m.cursor >= len(m.vault.Keys) && m.cursor > 0 {
			m.cursor--
		}
	case "n", "escape":
		m.deleting = -1
	}
	return m, nil
}

func (m KeysModel) View() string {
	var content strings.Builder

	header := titleStyle.Render("  SSH Keys") + "  " + subtitleStyle.Render(fmt.Sprintf("%d keys", len(m.vault.Keys)))
	content.WriteString(header)
	content.WriteString("\n\n")

	if len(m.vault.Keys) == 0 {
		content.WriteString(dimRow.Render("  No keys saved yet."))
		content.WriteString("\n")
		content.WriteString(dimRow.Render("  Press ") + footerKey.Render("a") + dimRow.Render(" to add your first key."))
	} else {
		for i, k := range m.vault.Keys {
			selected := i == m.cursor
			lines := strings.Count(k.PrivateKey, "\n") + 1
			detail := fmt.Sprintf("%d lines", lines)

			if selected {
				line := selectedRow.Render(" > ") + selectedRow.Render(fmt.Sprintf("%-20s", k.Name)) + "  " +
					lipgloss.NewStyle().Foreground(lipgloss.Color("#A78BFA")).Render(detail)
				line = lipgloss.NewStyle().
					Background(lipgloss.Color("#27272A")).
					Width(54).
					Render(line)
				content.WriteString(line)
			} else {
				content.WriteString("   " + normalRow.Render(fmt.Sprintf("%-20s", k.Name)) + "  " + dimRow.Render(detail))
			}
			content.WriteString("\n")
		}
	}

	if m.deleting >= 0 {
		name := m.vault.Keys[m.deleting].Name
		content.WriteString("\n")
		content.WriteString(confirmStyle.Render(fmt.Sprintf("  Delete \"%s\"? ", name)))
		content.WriteString(footerKey.Render("y") + dimRow.Render("es") + dimRow.Render(" / ") + footerKey.Render("n") + dimRow.Render("o"))
	} else {
		content.WriteString(footerBar(
			footerItem("a", "add"),
			footerItem("d", "delete"),
			footerItem("esc", "back"),
		))
	}

	box := boxStyle.Width(58).Render(content.String())

	if m.width > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
	}
	return box
}
