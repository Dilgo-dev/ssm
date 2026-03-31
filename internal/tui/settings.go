package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ssm/internal/cloud"
	"ssm/internal/config"
)

type settingField struct {
	label   string
	options []string
	value   string
}

type SettingsModel struct {
	fields    []settingField
	cursor    int
	cloudUser string
	lastPush  string
	lastPull  string
	width     int
	height    int
}

func NewSettingsModel(s *config.Settings) SettingsModel {
	cloudUser := ""
	if cfg, err := cloud.LoadCloud(); err == nil {
		cloudUser = cfg.Email
	}
	return SettingsModel{
		cloudUser: cloudUser,
		lastPush:  timeAgo(s.LastPush),
		lastPull:  timeAgo(s.LastPull),
		fields: []settingField{
			{
				label:   "Password cache",
				options: []string{"always", "session"},
				value:   s.PasswordCache,
			},
			{
				label:   "ThePrimeagen mode (vim keybind)",
				options: []string{"on", "off"},
				value:   boolToOnOff(s.VimKeys),
			},
			{
				label:   "Check updates",
				options: []string{"on", "off"},
				value:   boolToOnOff(s.AutoUpdate),
			},
			{
				label:   "Auto sync",
				options: []string{"on", "off"},
				value:   boolToOnOff(s.AutoSync),
			},
		},
	}
}

func timeAgo(ts string) string {
	if ts == "" {
		return "never"
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return "never"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(math.Round(d.Minutes()))
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(math.Round(d.Hours()))
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	default:
		days := int(math.Round(d.Hours() / 24))
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

func boolToOnOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func (m SettingsModel) Settings() *config.Settings {
	s := config.LoadSettings()
	s.PasswordCache = m.fields[0].value
	s.VimKeys = m.fields[1].value == "on"
	s.AutoUpdate = m.fields[2].value == "on"
	s.AutoSync = m.fields[3].value == "on"
	return s
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

	settingsLabelDim := fieldLabel.Width(34)
	if m.cloudUser != "" {
		content.WriteString("  " + settingsLabelDim.Render("Cloud account") + " " + fieldValue.Render(m.cloudUser))
	} else {
		content.WriteString("  " + settingsLabelDim.Render("Cloud account") + " " + dimRow.Render("not logged in"))
	}
	content.WriteString("\n")
	content.WriteString("  " + settingsLabelDim.Render("Last push") + " " + dimRow.Render(m.lastPush))
	content.WriteString("\n")
	content.WriteString("  " + settingsLabelDim.Render("Last pull") + " " + dimRow.Render(m.lastPull))
	content.WriteString("\n\n")

	settingsLabel := fieldLabel.Width(34)
	settingsLabelActive := fieldLabelActive.Width(34)

	for i, f := range m.fields {
		active := i == m.cursor

		var labelStr string
		if active {
			labelStr = settingsLabelActive.Render(f.label)
		} else {
			labelStr = settingsLabel.Render(f.label)
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
