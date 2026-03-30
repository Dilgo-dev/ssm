package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"ssm/internal/vault"
)

type SSHKey struct {
	Name       string `json:"name"`
	PrivateKey string `json:"private_key"`
}

type Connection struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password,omitempty"`
	KeyName  string `json:"key_name,omitempty"`
}

type Vault struct {
	Connections []Connection `json:"connections"`
	Keys        []SSHKey     `json:"keys"`
}

func (v *Vault) GetKey(name string) *SSHKey {
	for i, k := range v.Keys {
		if k.Name == name {
			return &v.Keys[i]
		}
	}
	return nil
}

func (v *Vault) KeyNames() []string {
	names := make([]string, len(v.Keys))
	for i, k := range v.Keys {
		names[i] = k.Name
	}
	return names
}

func (c Connection) SSHArgs() []string {
	var args []string
	if c.Port != 0 && c.Port != 22 {
		args = append(args, "-p", strconv.Itoa(c.Port))
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

func Load(masterPass string) (*Vault, error) {
	os.MkdirAll(Dir(), 0700)
	data, err := os.ReadFile(Path())
	if err != nil {
		if os.IsNotExist(err) {
			return &Vault{}, nil
		}
		return nil, err
	}

	plaintext, err := vault.Decrypt(data, masterPass)
	if err != nil {
		return nil, err
	}

	// Try new Vault format first
	var v Vault
	if err := json.Unmarshal(plaintext, &v); err == nil && (v.Connections != nil || v.Keys != nil) {
		return &v, nil
	}

	// Migration: old format was just []Connection
	var conns []Connection
	if err := json.Unmarshal(plaintext, &conns); err != nil {
		return nil, err
	}
	return &Vault{Connections: conns}, nil
}

func Save(v *Vault, masterPass string) error {
	os.MkdirAll(Dir(), 0700)
	plaintext, err := json.MarshalIndent(v, "", "  ")
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
