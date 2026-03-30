package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"

	"ssm/internal/config"
	"ssm/internal/update"
	"ssm/internal/vault"
)

var (
	masterPass string
	version    = "dev"
)

func main() {
	if len(os.Args) < 2 {
		checkUpdate()
		unlock()
		runTUI()
		showUpdateNotice()
		return
	}

	switch os.Args[1] {
	case "--version", "-v":
		fmt.Printf("ssm %s\n", version)
		return
	case "--help", "-h", "help":
		fmt.Printf("ssm %s - SSH connection manager\n", version)
		fmt.Print(`
Usage:
  ssm                  open interactive connection list
  ssm add              add a new connection
  ssm edit <name>      edit a connection
  ssm remove <name>    remove a connection
  ssm exec <name> <cmd> run a command on a remote server
  ssm keys             list saved SSH keys
  ssm keys add         add a new SSH key
  ssm keys remove <n>  remove a SSH key
  ssm update           update ssm to the latest version

Cloud (optional):
  ssm login            authenticate with sync server
  ssm register         create a sync account
  ssm push             upload encrypted vault
  ssm pull             download encrypted vault
  ssm logout           remove sync credentials

Shortcuts (in TUI):
  enter       connect        /    search
  a           add            d    delete
  K or k      manage keys    s    settings
  Ctrl+T n    new tab        Ctrl+T 1-9  switch tab
  Ctrl+T w    close tab      Ctrl+T d    detach
`)
		return
	case "update":
		fmt.Println("Checking for updates...")
		if err := update.Download(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
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
	case "keys":
		unlock()
		if len(os.Args) >= 3 {
			switch os.Args[2] {
			case "add":
				runKeysAdd()
			case "remove":
				if len(os.Args) < 4 {
					fmt.Println("Usage: ssm keys remove <name>")
					os.Exit(1)
				}
				runKeysRemove(os.Args[3])
			default:
				fmt.Printf("Unknown keys command: %s\n", os.Args[2])
				os.Exit(1)
			}
		} else {
			runKeysList()
		}
	case "exec":
		if len(os.Args) < 4 {
			fmt.Println("Usage: ssm exec <name> <command>")
			os.Exit(1)
		}
		unlock()
		runExec(os.Args[2], strings.Join(os.Args[3:], " "))
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
		fmt.Println("Usage: ssm [add|remove|edit|keys|exec|update|login|register|push|pull|logout]")
		os.Exit(1)
	}
}

func checkUpdate() {
	settings := config.LoadSettings()
	if settings.AutoUpdate && version != "dev" {
		update.CheckInBackground(version)
	}
}

func showUpdateNotice() {
	if version == "dev" {
		return
	}
	if v := update.GetAvailable(version); v != "" {
		fmt.Printf("\nssm %s available. Run 'ssm update' to upgrade.\n", v)
	}
}

func unlock() {
	if !config.Exists() {
		fmt.Print("Create a master password: ")
		pass1, err := term.ReadPassword(syscall.Stdin)
		fmt.Println()
		if err != nil || len(pass1) == 0 {
			fmt.Fprintln(os.Stderr, "Password required.")
			os.Exit(1)
		}

		fmt.Print("Confirm password: ")
		pass2, err := term.ReadPassword(syscall.Stdin)
		fmt.Println()
		if err != nil || string(pass1) != string(pass2) {
			fmt.Fprintln(os.Stderr, "Passwords do not match.")
			os.Exit(1)
		}

		masterPass = string(pass1)
		_ = config.Save(&config.Vault{}, masterPass)
		fmt.Println("Vault created.")
		settings := config.LoadSettings()
		if settings.PasswordCache == "session" {
			config.CachePassword(masterPass)
		}
		return
	}

	settings := config.LoadSettings()
	if settings.PasswordCache == "session" {
		if cached := config.GetCachedPassword(); cached != "" {
			if _, err := config.Load(cached); err == nil {
				masterPass = cached
				return
			}
			config.ClearPasswordCache()
		}
	}

	for attempts := 0; attempts < 3; attempts++ {
		fmt.Print("Master password: ")
		pass, err := term.ReadPassword(syscall.Stdin)
		fmt.Println()
		if err != nil {
			os.Exit(1)
		}

		_, err = config.Load(string(pass))
		if err == nil {
			masterPass = string(pass)
			if settings.PasswordCache == "session" {
				config.CachePassword(masterPass)
			}
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
