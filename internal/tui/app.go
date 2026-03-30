package tui

import (
	"strconv"

	tea "github.com/charmbracelet/bubbletea"

	"ssm/internal/config"
)

type appState int

const (
	stateList appState = iota
	stateAddConn
	stateEditConn
	stateKeys
	stateAddKey
	stateAddKeyFromConn
	stateSettings
)

type AppResult struct {
	Quit     bool
	Connect  *config.Connection
	ConnectV *config.Vault
}

type AppModel struct {
	state      appState
	prevState  appState
	list       ListModel
	form       FormModel
	savedForm  FormModel
	keys       KeysModel
	settings   SettingsModel
	vault      *config.Vault
	masterPass string
	editIdx    int
	Result     AppResult
	width      int
	height     int
}

func NewApp(v *config.Vault, masterPass string) AppModel {
	return AppModel{
		state:      stateList,
		list:       NewListModel(v, masterPass),
		vault:      v,
		masterPass: masterPass,
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.EnableBracketedPaste
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = ws.Width
		m.height = ws.Height
	}

	switch m.state {
	case stateList:
		return m.updateList(msg)
	case stateAddConn:
		return m.updateConnForm(msg)
	case stateEditConn:
		return m.updateConnForm(msg)
	case stateKeys:
		return m.updateKeys(msg)
	case stateAddKey:
		return m.updateAddKey(msg)
	case stateAddKeyFromConn:
		return m.updateAddKeyFromConn(msg)
	case stateSettings:
		return m.updateSettings(msg)
	}
	return m, nil
}

func (m AppModel) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, _ := m.list.Update(msg)
	m.list = updated.(ListModel)

	switch m.list.Action {
	case ActionConnect:
		m.list.Action = ActionNone
		m.Result = AppResult{Connect: m.list.Selected, ConnectV: m.vault}
		return m, tea.Quit
	case ActionAdd:
		m.list.Action = ActionNone
		m.state = stateAddConn
		m.form = m.newConnForm()
		return m, nil
	case ActionKeys:
		m.list.Action = ActionNone
		m.state = stateKeys
		m.keys = NewKeysModel(m.vault, m.masterPass)
		m.keys.width = m.width
		m.keys.height = m.height
		return m, nil
	case ActionSettings:
		m.list.Action = ActionNone
		m.state = stateSettings
		m.settings = NewSettingsModel(config.LoadSettings())
		m.settings.width = m.width
		m.settings.height = m.height
		return m, nil
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		if key.String() == "q" || key.String() == "ctrl+c" {
			m.Result = AppResult{Quit: true}
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m AppModel) updateConnForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, _ := m.form.Update(msg)
	m.form = updated.(FormModel)

	if m.form.Canceled {
		m.form.Canceled = false
		m.goToList()
		return m, nil
	}

	if m.form.AddKey {
		m.form.AddKey = false
		m.savedForm = m.form
		m.prevState = m.state
		m.state = stateAddKeyFromConn
		m.form = m.newKeyForm()
		return m, nil
	}

	if m.form.Done {
		m.form.Done = false
		if m.state == stateAddConn {
			m.saveNewConn()
		} else {
			m.saveEditConn()
		}
		m.goToList()
		return m, nil
	}

	return m, nil
}

func (m AppModel) updateKeys(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, _ := m.keys.Update(msg)
	m.keys = updated.(KeysModel)

	if m.keys.Action == KeyActionAdd {
		m.keys.Action = KeyActionNone
		m.state = stateAddKey
		m.form = m.newKeyForm()
		return m, nil
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		str := key.String()
		if (str == "q" || str == "ctrl+c" || str == "esc") && m.keys.deleting < 0 {
			m.goToList()
			return m, nil
		}
	}

	return m, nil
}

func (m AppModel) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, _ := m.settings.Update(msg)
	m.settings = updated.(SettingsModel)

	if key, ok := msg.(tea.KeyMsg); ok {
		if key.String() == "esc" || key.String() == "q" || key.String() == "ctrl+c" {
			_ = config.SaveSettings(m.settings.Settings())
			m.goToList()
			return m, nil
		}
	}

	return m, nil
}

func (m AppModel) updateAddKey(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, _ := m.form.Update(msg)
	m.form = updated.(FormModel)

	if m.form.Canceled {
		m.form.Canceled = false
		m.state = stateKeys
		m.reloadVault()
		m.keys = NewKeysModel(m.vault, m.masterPass)
		m.keys.width = m.width
		m.keys.height = m.height
		return m, nil
	}

	if m.form.Done {
		m.form.Done = false
		m.saveKey()
		m.state = stateKeys
		m.reloadVault()
		m.keys = NewKeysModel(m.vault, m.masterPass)
		m.keys.width = m.width
		m.keys.height = m.height
		return m, nil
	}

	return m, nil
}

func (m AppModel) updateAddKeyFromConn(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, _ := m.form.Update(msg)
	m.form = updated.(FormModel)

	if m.form.Canceled {
		m.form.Canceled = false
		m.state = m.prevState
		m.form = m.savedForm
		return m, nil
	}

	if m.form.Done {
		m.form.Done = false
		addedName := m.saveKey()
		m.state = m.prevState
		m.form = m.savedForm
		m.reloadVault()
		m.refreshKeyOptions()
		if addedName != "" {
			m.setKeyValue(addedName)
		}
		return m, nil
	}

	return m, nil
}

func (m *AppModel) goToList() {
	m.reloadVault()
	m.state = stateList
	m.list = NewListModel(m.vault, m.masterPass)
	m.list.width = m.width
	m.list.height = m.height
}

func (m *AppModel) reloadVault() {
	v, _ := config.Load(m.masterPass)
	m.vault = v
}

func (m *AppModel) keyOpts() []string {
	opts := []string{"(none)"}
	opts = append(opts, m.vault.KeyNames()...)
	opts = append(opts, "+ Add new key")
	return opts
}

func (m *AppModel) newConnForm() FormModel {
	f := NewFormModel("New connection", []Field{
		{Label: "Name", Required: true},
		{Label: "Host", Required: true},
		{Label: "Port", Value: "22", Placeholder: "22"},
		{Label: "User", Required: true},
		{Label: "Password", Password: true},
		{Label: "SSH Key", Value: "(none)", Options: m.keyOpts()},
	})
	f.width = m.width
	f.height = m.height
	return f
}

func (m *AppModel) newKeyForm() FormModel {
	f := NewFormModel("Add SSH key", []Field{
		{Label: "Name", Required: true, Placeholder: "production-key"},
		{Label: "Private key", Required: true, Placeholder: "paste your key here"},
	})
	f.width = m.width
	f.height = m.height
	return f
}

func (m *AppModel) refreshKeyOptions() {
	opts := m.keyOpts()
	for i, f := range m.form.Fields {
		if f.Label == "SSH Key" {
			m.form.Fields[i].Options = opts
		}
	}
}

func (m *AppModel) setKeyValue(name string) {
	for i, f := range m.form.Fields {
		if f.Label == "SSH Key" {
			m.form.Fields[i].Value = name
		}
	}
}

func (m *AppModel) saveNewConn() {
	name := m.form.GetValue("Name")
	port, _ := strconv.Atoi(m.form.GetValue("Port"))
	if port == 0 {
		port = 22
	}
	keyName := m.form.GetValue("SSH Key")
	if keyName == "(none)" {
		keyName = ""
	}
	m.vault.Connections = append(m.vault.Connections, config.Connection{
		Name:     name,
		Host:     m.form.GetValue("Host"),
		Port:     port,
		User:     m.form.GetValue("User"),
		Password: m.form.GetValue("Password"),
		KeyName:  keyName,
	})
	_ = config.Save(m.vault, m.masterPass)
}

func (m *AppModel) saveEditConn() {
	port, _ := strconv.Atoi(m.form.GetValue("Port"))
	if port == 0 {
		port = 22
	}
	keyName := m.form.GetValue("SSH Key")
	if keyName == "(none)" {
		keyName = ""
	}
	c := &m.vault.Connections[m.editIdx]
	c.Host = m.form.GetValue("Host")
	c.Port = port
	c.User = m.form.GetValue("User")
	c.Password = m.form.GetValue("Password")
	c.KeyName = keyName
	_ = config.Save(m.vault, m.masterPass)
}

func (m *AppModel) saveKey() string {
	name := m.form.GetValue("Name")
	content := m.form.GetValue("Private key")
	for _, k := range m.vault.Keys {
		if k.Name == name {
			return ""
		}
	}
	m.vault.Keys = append(m.vault.Keys, config.SSHKey{
		Name:       name,
		PrivateKey: content,
	})
	_ = config.Save(m.vault, m.masterPass)
	return name
}

func (m AppModel) View() string {
	switch m.state {
	case stateList:
		return m.list.View()
	case stateAddConn, stateEditConn:
		return m.form.View()
	case stateKeys:
		return m.keys.View()
	case stateAddKey, stateAddKeyFromConn:
		return m.form.View()
	case stateSettings:
		return m.settings.View()
	}
	return ""
}
