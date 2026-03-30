package main

import (
	"bufio"
	"fmt"
	"os"

	"golang.org/x/term"

	"ssm/internal/cloud"
)

func prompt(label string) string {
	fmt.Printf("%s: ", label)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

func promptSecret(label string) string {
	fmt.Printf("%s: ", label)
	pass, _ := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return string(pass)
}

func runRegister() {
	server := prompt("Server URL")
	email := prompt("Email")
	password := promptSecret("Password")

	token, err := cloud.Register(server, email, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	_ = cloud.SaveCloud(&cloud.CloudConfig{Server: server, Token: token})
	fmt.Println("Account created and logged in.")
}

func runLogin() {
	server := prompt("Server URL")
	email := prompt("Email")
	password := promptSecret("Password")

	token, err := cloud.Login(server, email, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	_ = cloud.SaveCloud(&cloud.CloudConfig{Server: server, Token: token})
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
