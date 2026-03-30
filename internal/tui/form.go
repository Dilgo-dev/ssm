package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Field struct {
	Label       string
	Value       string
	Placeholder string
	Password    bool
	Required    bool
	Options     []string // if set, field becomes a select (left/right to cycle)
}

type FormModel struct {
	title    string
	Fields   []Field
	cursor   int
	Done     bool
	Canceled bool
	AddKey   bool
	err      string
	width    int
	height   int
}

func NewFormModel(title string, fields []Field) FormModel {
	return FormModel{
		title:  title,
		Fields: fields,
	}
}

func (m FormModel) GetValue(label string) string {
	for _, f := range m.Fields {
		if f.Label == label {
			return strings.TrimSpace(f.Value)
		}
	}
	return ""
}

func (m FormModel) Init() tea.Cmd {
	return tea.EnableBracketedPaste
}

func (m FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		f := &m.Fields[m.cursor]

		if f.Options != nil {
			return m.handleSelect(msg)
		}

		if msg.Paste {
			f.Value += string(msg.Runes)
			m.err = ""
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "esc":
			m.Canceled = true
			return m, tea.Quit
		case "up":
			if m.cursor > 0 {
				m.cursor--
				m.err = ""
			}
		case "down":
			if m.cursor < len(m.Fields)-1 {
				m.cursor++
				m.err = ""
			}
		case "tab":
			if m.cursor < len(m.Fields)-1 {
				m.cursor++
				m.err = ""
			}
		case "shift+tab":
			if m.cursor > 0 {
				m.cursor--
				m.err = ""
			}
		case "enter":
			return m.validate()
		case "backspace":
			if len(f.Value) > 0 {
				f.Value = f.Value[:len(f.Value)-1]
			}
		default:
			if len(msg.Runes) > 0 {
				f.Value += string(msg.Runes)
				m.err = ""
			}
		}
	}
	return m, nil
}

func (m FormModel) handleSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	f := &m.Fields[m.cursor]
	switch msg.String() {
	case "ctrl+c", "esc":
		m.Canceled = true
		return m, nil
	case "left", "h":
		idx := m.selectIndex(f)
		if idx > 0 {
			f.Value = f.Options[idx-1]
		}
	case "right", "l":
		idx := m.selectIndex(f)
		if idx < len(f.Options)-1 {
			f.Value = f.Options[idx+1]
		}
	case "up", "shift+tab":
		if m.cursor > 0 {
			m.cursor--
			m.err = ""
		}
	case "down", "tab":
		if m.cursor < len(m.Fields)-1 {
			m.cursor++
			m.err = ""
		}
	case "enter":
		if f.Value == "+ Add new key" {
			m.AddKey = true
			return m, tea.Quit
		}
		return m.validate()
	}
	return m, nil
}

func (m FormModel) selectIndex(f *Field) int {
	for i, opt := range f.Options {
		if opt == f.Value {
			return i
		}
	}
	return 0
}

func (m FormModel) validate() (tea.Model, tea.Cmd) {
	for i, f := range m.Fields {
		if f.Required && strings.TrimSpace(f.Value) == "" {
			m.cursor = i
			m.err = fmt.Sprintf("\"%s\" is required", f.Label)
			return m, nil
		}
	}
	for _, f := range m.Fields {
		if f.Label == "Port" && f.Value != "" {
			for _, ch := range f.Value {
				if ch < '0' || ch > '9' {
					m.err = "Invalid port"
					return m, nil
				}
			}
		}
	}
	m.Done = true
	return m, tea.Quit
}

func (m FormModel) View() string {
	var content strings.Builder

	content.WriteString(titleStyle.Render("  " + m.title))
	content.WriteString("\n")
	progress := fmt.Sprintf("  %d / %d", m.cursor+1, len(m.Fields))
	content.WriteString(subtitleStyle.Render(progress))
	content.WriteString("\n\n")

	for i, f := range m.Fields {
		active := i == m.cursor

		label := f.Label
		if f.Required {
			label += " *"
		}

		var labelStr string
		if active {
			labelStr = fieldLabelActive.Render(label)
		} else {
			labelStr = fieldLabel.Render(label)
		}

		var valueStr string
		if f.Options != nil {
			if active {
				valueStr = fieldCursor.Render("< ") + fieldValue.Render(f.Value) + fieldCursor.Render(" >")
			} else {
				valueStr = fieldValue.Render(f.Value)
			}
		} else if f.Value == "" {
			if f.Placeholder != "" {
				valueStr = fieldPlaceholder.Render(f.Placeholder)
			} else {
				valueStr = fieldPlaceholder.Render("...")
			}
		} else if f.Password {
			valueStr = fieldValue.Render(strings.Repeat("*", len(f.Value)))
		} else {
			valueStr = fieldValue.Render(f.Value)
		}

		if active && f.Options == nil {
			valueStr += fieldCursor.Render("|")
		}

		var inputLine string
		if active {
			inputLine = fieldInputBox.Render(valueStr)
		} else {
			inputLine = fieldInputBoxInactive.Render(valueStr)
		}

		indicator := "  "
		if active {
			indicator = selectedRow.Render("> ")
		}

		content.WriteString(indicator + labelStr + " " + inputLine)
		content.WriteString("\n")
	}

	if m.err != "" {
		content.WriteString("\n")
		content.WriteString(errorStyle.Render("  ! " + m.err))
	}

	content.WriteString(footerBar(
		footerItem("↑↓", "navigate"),
		footerItem("tab", "next"),
		footerItem("enter", "submit"),
		footerItem("esc", "cancel"),
	))

	box := boxStyle.Width(58).Render(content.String())

	if m.width > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
	}
	return box
}
