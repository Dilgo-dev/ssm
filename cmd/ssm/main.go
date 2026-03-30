package main

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/term"

	"ssm/internal/config"
	"ssm/internal/vault"
)

var masterPass string

func main() {
	if len(os.Args) < 2 {
		unlock()
		runTUI()
		return
	}

	switch os.Args[1] {
	case "add":
		unlock()
		runAdd()
	case "remove":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ssm remove <name>")
			os.Exit(1)
		}
		unlock()
		runRemove(os.Args[2])
	case "edit":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ssm edit <name>")
			os.Exit(1)
		}
		unlock()
		runEdit(os.Args[2])
	case "register":
		runRegister()
	case "login":
		runLogin()
	case "logout":
		runLogout()
	case "push":
		runPush()
	case "pull":
		runPull()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		fmt.Println("Usage: ssm [add|remove|edit|login|register|push|pull|logout]")
		os.Exit(1)
	}
}

func unlock() {
	if !config.Exists() {
		fmt.Print("Create a master password: ")
		pass1, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil || len(pass1) == 0 {
			fmt.Fprintln(os.Stderr, "Password required.")
			os.Exit(1)
		}

		fmt.Print("Confirm password: ")
		pass2, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil || string(pass1) != string(pass2) {
			fmt.Fprintln(os.Stderr, "Passwords do not match.")
			os.Exit(1)
		}

		masterPass = string(pass1)
		config.Save([]config.Connection{}, masterPass)
		fmt.Println("Vault created.")
		return
	}

	for attempts := 0; attempts < 3; attempts++ {
		fmt.Print("Master password: ")
		pass, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			os.Exit(1)
		}

		_, err = config.Load(string(pass))
		if err == nil {
			masterPass = string(pass)
			return
		}
		if err == vault.ErrWrongPassword {
			fmt.Fprintln(os.Stderr, "Wrong password.")
			continue
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Too many attempts.")
	os.Exit(1)
}
