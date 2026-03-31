package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type UnlockMode int

const (
	UnlockLogin UnlockMode = iota
	UnlockCreate
	UnlockCloudMerge
)

type UnlockModel struct {
	mode     UnlockMode
	password string
	confirm  string
	step     int // 0 = password, 1 = confirm (create only)
	err      string
	attempt  int
	Done     bool
	Canceled bool
	Password string
	width    int
	height   int
}

func NewUnlockModel(mode UnlockMode) UnlockModel {
	return UnlockModel{mode: mode}
}

func (m UnlockModel) Init() tea.Cmd { return nil }

func (m UnlockModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.Canceled = true
			return m, tea.Quit
		case "enter":
			return m.submit()
		case "backspace":
			if m.step == 0 && len(m.password) > 0 {
				m.password = m.password[:len(m.password)-1]
			} else if m.step == 1 && len(m.confirm) > 0 {
				m.confirm = m.confirm[:len(m.confirm)-1]
			}
			m.err = ""
		default:
			if len(msg.Runes) > 0 {
				if m.step == 0 {
					m.password += string(msg.Runes)
				} else {
					m.confirm += string(msg.Runes)
				}
				m.err = ""
			}
		}
	}
	return m, nil
}

func (m UnlockModel) submit() (tea.Model, tea.Cmd) {
	if m.mode == UnlockCreate {
		if m.step == 0 {
			if len(m.password) == 0 {
				m.err = "Password required."
				return m, nil
			}
			m.step = 1
			m.err = ""
			return m, nil
		}
		if m.password != m.confirm {
			m.confirm = ""
			m.err = "Passwords do not match."
			return m, nil
		}
		m.Password = m.password
		m.Done = true
		return m, tea.Quit
	}

	if len(m.password) == 0 {
		m.err = "Password required."
		return m, nil
	}
	m.Password = m.password
	m.Done = true
	return m, tea.Quit
}

func (m UnlockModel) SetError(msg string) UnlockModel {
	m.err = msg
	m.password = ""
	m.attempt++
	return m
}

func (m UnlockModel) View() string {
	var content strings.Builder

	content.WriteString(titleStyle.Render("  ~ ssm"))
	content.WriteString("\n\n")

	switch m.mode {
	case UnlockCreate:
		content.WriteString(normalRow.Render("  Create a master password to encrypt your vault."))
		content.WriteString("\n\n")

		label1 := fieldLabel.Render("Password")
		label2 := fieldLabel.Render("Confirm")

		if m.step == 0 {
			label1 = fieldLabelActive.Render("Password")
			dots := fieldValue.Render(strings.Repeat("*", len(m.password)))
			content.WriteString(selectedRow.Render("  > ") + label1 + " " + dots + fieldCursor.Render("|"))
			content.WriteString("\n")
			content.WriteString("    " + label2 + " " + fieldPlaceholder.Render("..."))
		} else {
			dots1 := fieldValue.Render(strings.Repeat("*", len(m.password)))
			content.WriteString("    " + label1 + " " + dots1)
			content.WriteString("\n")
			label2 = fieldLabelActive.Render("Confirm")
			dots2 := fieldValue.Render(strings.Repeat("*", len(m.confirm)))
			content.WriteString(selectedRow.Render("  > ") + label2 + " " + dots2 + fieldCursor.Render("|"))
		}
	case UnlockCloudMerge:
		content.WriteString(normalRow.Render("  Cloud vault uses a different password."))
		content.WriteString("\n")
		content.WriteString(normalRow.Render("  Enter your cloud vault password to merge."))
		content.WriteString("\n\n")

		label := fieldLabelActive.Render("Password")
		dots := fieldValue.Render(strings.Repeat("*", len(m.password)))
		content.WriteString(selectedRow.Render("  > ") + label + " " + dots + fieldCursor.Render("|"))
	default:
		content.WriteString(normalRow.Render("  Unlock your vault."))
		content.WriteString("\n\n")

		label := fieldLabelActive.Render("Password")
		dots := fieldValue.Render(strings.Repeat("*", len(m.password)))
		content.WriteString(selectedRow.Render("  > ") + label + " " + dots + fieldCursor.Render("|"))
	}

	if m.err != "" {
		content.WriteString("\n\n")
		content.WriteString(errorStyle.Render("  ! " + m.err))
	}

	content.WriteString("\n")
	content.WriteString(footerBar(
		footerItem("enter", "submit"),
		footerItem("ctrl+c", "quit"),
	))

	out := lipgloss.NewStyle().Padding(1, 3).Render(content.String())

	if m.width > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, out)
	}
	return out
}
