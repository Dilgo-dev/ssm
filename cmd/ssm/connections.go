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
		conns, err := config.Load(masterPass)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		p := tea.NewProgram(tui.NewListModel(conns, masterPass), tea.WithAltScreen())
		result, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		m := result.(tui.ListModel)
		switch m.Action {
		case tui.ActionConnect:
			ssh.Connect(*m.Selected)
		case tui.ActionAdd:
			runAdd()
			continue
		}
		break
	}
}

func runAdd() {
	fields := []tui.Field{
		{Label: "Name", Required: true},
		{Label: "Host", Required: true},
		{Label: "Port", Value: "22", Placeholder: "22"},
		{Label: "User", Required: true},
		{Label: "Password", Password: true},
		{Label: "Identity file"},
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
	conns, _ := config.Load(masterPass)
	for _, c := range conns {
		if c.Name == name {
			fmt.Printf("Connection \"%s\" already exists.\n", name)
			return
		}
	}

	port, _ := strconv.Atoi(fm.GetValue("Port"))
	if port == 0 {
		port = 22
	}

	conn := config.Connection{
		Name:         name,
		Host:         fm.GetValue("Host"),
		Port:         port,
		User:         fm.GetValue("User"),
		Password:     fm.GetValue("Password"),
		IdentityFile: fm.GetValue("Identity file"),
	}

	conns = append(conns, conn)
	if err := config.Save(conns, masterPass); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Connection \"%s\" added.\n", name)
}

func runRemove(name string) {
	conns, err := config.Load(masterPass)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	found := -1
	for i, c := range conns {
		if c.Name == name {
			found = i
			break
		}
	}
	if found == -1 {
		fmt.Printf("Connection \"%s\" not found.\n", name)
		os.Exit(1)
	}

	conns = append(conns[:found], conns[found+1:]...)
	if err := config.Save(conns, masterPass); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Connection \"%s\" removed.\n", name)
}

func runEdit(name string) {
	conns, err := config.Load(masterPass)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	found := -1
	for i, c := range conns {
		if c.Name == name {
			found = i
			break
		}
	}
	if found == -1 {
		fmt.Printf("Connection \"%s\" not found.\n", name)
		os.Exit(1)
	}

	c := conns[found]
	fields := []tui.Field{
		{Label: "Host", Value: c.Host, Required: true},
		{Label: "Port", Value: strconv.Itoa(c.Port), Placeholder: "22"},
		{Label: "User", Value: c.User, Required: true},
		{Label: "Password", Value: c.Password, Password: true},
		{Label: "Identity file", Value: c.IdentityFile},
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

	c.Host = fm.GetValue("Host")
	c.Port = port
	c.User = fm.GetValue("User")
	c.Password = fm.GetValue("Password")
	c.IdentityFile = fm.GetValue("Identity file")

	conns[found] = c
	if err := config.Save(conns, masterPass); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Connection \"%s\" updated.\n", name)
}
