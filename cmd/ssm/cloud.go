package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"ssm/internal/cloud"
	"ssm/internal/tui"
)

const defaultServer = "https://api.gossm.sh"

func runRegister() {
	fields := []tui.Field{
		{Label: "Server", Value: defaultServer},
		{Label: "Email", Required: true},
		{Label: "Password", Required: true, Password: true},
		{Label: "Confirm", Required: true, Password: true},
	}

	p := tea.NewProgram(tui.NewFormModel("Create account", fields), tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fm := result.(tui.FormModel)
	if fm.Canceled || !fm.Done {
		return
	}

	password := fm.GetValue("Password")
	if password != fm.GetValue("Confirm") {
		fmt.Fprintln(os.Stderr, "Passwords do not match.")
		os.Exit(1)
	}

	server := fm.GetValue("Server")
	email := fm.GetValue("Email")

	fmt.Println("Creating account...")
	token, err := cloud.Register(server, email, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cfg := &cloud.CloudConfig{Server: server, Token: token, Email: email}
	_ = cloud.SaveCloud(cfg)

	check := func() bool {
		return cloud.CheckVerified(cfg)
	}

	vp := tea.NewProgram(tui.NewVerifyModel(email, check), tea.WithAltScreen())
	vResult, err := vp.Run()
	if err != nil {
		fmt.Println("Account created. Check your email to verify your account.")
		return
	}
	vm := vResult.(tui.VerifyModel)
	if vm.Verified() {
		fmt.Println("Account verified and ready.")
	} else {
		fmt.Println("Account created. Verify your email to use cloud sync.")
	}
}

func runLogin() {
	fields := []tui.Field{
		{Label: "Server", Value: defaultServer},
		{Label: "Email", Required: true},
		{Label: "Password", Required: true, Password: true},
	}

	p := tea.NewProgram(tui.NewFormModel("Login", fields), tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fm := result.(tui.FormModel)
	if fm.Canceled || !fm.Done {
		return
	}

	server := fm.GetValue("Server")
	email := fm.GetValue("Email")
	password := fm.GetValue("Password")

	fmt.Println("Logging in...")
	token, err := cloud.Login(server, email, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	_ = cloud.SaveCloud(&cloud.CloudConfig{Server: server, Token: token, Email: email})
	fmt.Println("Logged in.")
}

func runLogout() {
	if err := cloud.DeleteCloud(); err != nil {
		fmt.Fprintln(os.Stderr, "Not logged in.")
		os.Exit(1)
	}
	fmt.Println("Logged out.")
}

func runPush() {
	cfg, err := cloud.LoadCloud()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := cloud.Push(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Vault pushed to cloud.")
}

func runPull() {
	cfg, err := cloud.LoadCloud()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := cloud.Pull(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Vault pulled from cloud.")
}
