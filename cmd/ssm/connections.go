package main

import (
	"fmt"
	"os"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"

	"ssm/internal/config"
	"ssm/internal/ssh"
	"ssm/internal/tui"
)

func runTUI() {
	for {
		v, err := config.Load(masterPass)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		p := tea.NewProgram(tui.NewApp(v, masterPass), tea.WithAltScreen())
		result, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		app := result.(tui.AppModel)
		if app.Result.Connect != nil {
			picker := func() *config.Connection {
				v2, _ := config.Load(masterPass)
				p2 := tea.NewProgram(tui.NewApp(v2, masterPass), tea.WithAltScreen())
				r2, err := p2.Run()
				if err != nil {
					return nil
				}
				a2 := r2.(tui.AppModel)
				return a2.Result.Connect
			}
			ssh.ConnectWithManager(*app.Result.Connect, app.Result.ConnectV, picker)
			continue
		}
		break
	}
}

func keyOptions() []string {
	v, _ := config.Load(masterPass)
	opts := []string{"(none)"}
	opts = append(opts, v.KeyNames()...)
	opts = append(opts, "+ Add new key")
	return opts
}

func runAdd() {
	fields := []tui.Field{
		{Label: "Name", Required: true},
		{Label: "Host", Required: true},
		{Label: "Port", Value: "22", Placeholder: "22"},
		{Label: "User", Required: true},
		{Label: "Password", Password: true},
		{Label: "SSH Key", Value: "(none)", Options: keyOptions()},
	}

	p := tea.NewProgram(tui.NewFormModel("New connection", fields), tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	fm := result.(tui.FormModel)
	if fm.Canceled || !fm.Done {
		return
	}

	name := fm.GetValue("Name")
	v, _ := config.Load(masterPass)
	for _, c := range v.Connections {
		if c.Name == name {
			fmt.Printf("Connection \"%s\" already exists.\n", name)
			return
		}
	}

	port, _ := strconv.Atoi(fm.GetValue("Port"))
	if port == 0 {
		port = 22
	}

	keyName := fm.GetValue("SSH Key")
	if keyName == "(none)" {
		keyName = ""
	}

	conn := config.Connection{
		Name:     name,
		Host:     fm.GetValue("Host"),
		Port:     port,
		User:     fm.GetValue("User"),
		Password: fm.GetValue("Password"),
		KeyName:  keyName,
	}

	v.Connections = append(v.Connections, conn)
	if err := config.Save(v, masterPass); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Connection \"%s\" added.\n", name)
}

func runRemove(name string) {
	v, err := config.Load(masterPass)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	found := -1
	for i, c := range v.Connections {
		if c.Name == name {
			found = i
			break
		}
	}
	if found == -1 {
		fmt.Printf("Connection \"%s\" not found.\n", name)
		os.Exit(1)
	}

	v.Connections = append(v.Connections[:found], v.Connections[found+1:]...)
	if err := config.Save(v, masterPass); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Connection \"%s\" removed.\n", name)
}

func runExec(name, cmd string) {
	v, err := config.Load(masterPass)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for _, c := range v.Connections {
		if c.Name == name {
			os.Exit(ssh.Exec(c, v, cmd))
		}
	}
	fmt.Printf("Connection \"%s\" not found.\n", name)
	os.Exit(1)
}

func runEdit(name string) {
	v, err := config.Load(masterPass)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	found := -1
	for i, c := range v.Connections {
		if c.Name == name {
			found = i
			break
		}
	}
	if found == -1 {
		fmt.Printf("Connection \"%s\" not found.\n", name)
		os.Exit(1)
	}

	c := v.Connections[found]
	keyVal := c.KeyName
	if keyVal == "" {
		keyVal = "(none)"
	}

	fields := []tui.Field{
		{Label: "Host", Value: c.Host, Required: true},
		{Label: "Port", Value: strconv.Itoa(c.Port), Placeholder: "22"},
		{Label: "User", Value: c.User, Required: true},
		{Label: "Password", Value: c.Password, Password: true},
		{Label: "SSH Key", Value: keyVal, Options: keyOptions()},
	}

	p := tea.NewProgram(tui.NewFormModel("Edit: "+name, fields), tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	fm := result.(tui.FormModel)
	if fm.Canceled || !fm.Done {
		return
	}

	port, _ := strconv.Atoi(fm.GetValue("Port"))
	if port == 0 {
		port = 22
	}

	keyName := fm.GetValue("SSH Key")
	if keyName == "(none)" {
		keyName = ""
	}

	c.Host = fm.GetValue("Host")
	c.Port = port
	c.User = fm.GetValue("User")
	c.Password = fm.GetValue("Password")
	c.KeyName = keyName

	v.Connections[found] = c
	if err := config.Save(v, masterPass); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Connection \"%s\" updated.\n", name)
}
