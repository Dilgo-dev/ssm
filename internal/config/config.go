package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"ssm/internal/vault"
)

type Connection struct {
	Name         string `json:"name"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	User         string `json:"user"`
	Password     string `json:"password,omitempty"`
	IdentityFile string `json:"identity_file,omitempty"`
}

func (c Connection) SSHArgs() []string {
	var args []string
	if c.Port != 0 && c.Port != 22 {
		args = append(args, "-p", strconv.Itoa(c.Port))
	}
	if c.IdentityFile != "" {
		idFile := c.IdentityFile
		if strings.HasPrefix(idFile, "~/") {
			home, _ := os.UserHomeDir()
			idFile = filepath.Join(home, idFile[2:])
		}
		args = append(args, "-i", idFile)
	}
	args = append(args, fmt.Sprintf("%s@%s", c.User, c.Host))
	return args
}

func (c Connection) Display() string {
	port := ""
	if c.Port != 0 && c.Port != 22 {
		port = fmt.Sprintf(":%d", c.Port)
	}
	return fmt.Sprintf("%s@%s%s", c.User, c.Host, port)
}

func Dir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ssm")
}

func Path() string {
	return filepath.Join(Dir(), "connections.enc")
}

func Exists() bool {
	_, err := os.Stat(Path())
	return err == nil
}

func Load(masterPass string) ([]Connection, error) {
	os.MkdirAll(Dir(), 0700)
	data, err := os.ReadFile(Path())
	if err != nil {
		if os.IsNotExist(err) {
			return []Connection{}, nil
		}
		return nil, err
	}

	plaintext, err := vault.Decrypt(data, masterPass)
	if err != nil {
		return nil, err
	}

	var conns []Connection
	if err := json.Unmarshal(plaintext, &conns); err != nil {
		return nil, err
	}
	return conns, nil
}

func Save(conns []Connection, masterPass string) error {
	os.MkdirAll(Dir(), 0700)
	plaintext, err := json.MarshalIndent(conns, "", "  ")
	if err != nil {
		return err
	}

	encrypted, err := vault.Encrypt(plaintext, masterPass)
	if err != nil {
		return err
	}

	tmp := Path() + ".tmp"
	if err := os.WriteFile(tmp, encrypted, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, Path())
}
