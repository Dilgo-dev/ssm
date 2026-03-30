package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ssm/internal/config"
)

type settingField struct {
	label   string
	options []string
	value   string
}

type SettingsModel struct {
	fields []settingField
	cursor int
	width  int
	height int
}

func NewSettingsModel(s *config.Settings) SettingsModel {
	return SettingsModel{
		fields: []settingField{
			{
				label:   "Password cache",
				options: []string{"always", "session"},
				value:   s.PasswordCache,
			},
			{
				label:   "Vim keys",
				options: []string{"on", "off"},
				value:   boolToOnOff(s.VimKeys),
			},
			{
				label:   "Check updates",
				options: []string{"on", "off"},
				value:   boolToOnOff(s.AutoUpdate),
			},
		},
	}
}

func boolToOnOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func (m SettingsModel) Settings() *config.Settings {
	return &config.Settings{
		PasswordCache: m.fields[0].value,
		VimKeys:       m.fields[1].value == "on",
		AutoUpdate:    m.fields[2].value == "on",
	}
}

func (m SettingsModel) Init() tea.Cmd { return nil }

func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "ctrl+c":
			_ = config.SaveSettings(m.Settings())
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.fields)-1 {
				m.cursor++
			}
		case "left", "h":
			f := &m.fields[m.cursor]
			idx := m.optionIndex(f)
			if idx > 0 {
				f.value = f.options[idx-1]
			}
		case "right", "l":
			f := &m.fields[m.cursor]
			idx := m.optionIndex(f)
			if idx < len(f.options)-1 {
				f.value = f.options[idx+1]
			}
		}
	}
	return m, nil
}

func (m SettingsModel) optionIndex(f *settingField) int {
	for i, opt := range f.options {
		if opt == f.value {
			return i
		}
	}
	return 0
}

func (m SettingsModel) View() string {
	var content strings.Builder

	content.WriteString(titleStyle.Render("  Settings"))
	content.WriteString("\n\n")

	for i, f := range m.fields {
		active := i == m.cursor

		var labelStr string
		if active {
			labelStr = fieldLabelActive.Render(f.label)
		} else {
			labelStr = fieldLabel.Render(f.label)
		}

		var valueStr string
		if active {
			valueStr = fieldCursor.Render("< ") + fieldValue.Render(f.value) + fieldCursor.Render(" >")
		} else {
			valueStr = fieldValue.Render(f.value)
		}

		indicator := "  "
		if active {
			indicator = selectedRow.Render("> ")
		}

		content.WriteString(indicator + labelStr + " " + valueStr)
		content.WriteString("\n")
	}

	content.WriteString(footerBar(
		footerItem("←→", "change"),
		footerItem("esc", "back"),
	))

	out := lipgloss.NewStyle().Padding(1, 3).Render(content.String())

	if m.width > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, out)
	}
	return out
}
