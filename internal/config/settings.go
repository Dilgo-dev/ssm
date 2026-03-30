package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"ssm/internal/vault"
)

type Settings struct {
	PasswordCache string `json:"password_cache"`
	VimKeys       bool   `json:"vim_keys"`
	AutoUpdate    bool   `json:"auto_update"`
}

func DefaultSettings() *Settings {
	return &Settings{PasswordCache: "always", VimKeys: true, AutoUpdate: true}
}

func settingsPath() string {
	return filepath.Join(Dir(), "settings.json")
}

func cachePath() string {
	return fmt.Sprintf("/tmp/ssm-cache-%d", os.Getuid())
}

func LoadSettings() *Settings {
	data, err := os.ReadFile(settingsPath())
	if err != nil {
		return DefaultSettings()
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return DefaultSettings()
	}
	if s.PasswordCache == "" {
		s.PasswordCache = "always"
	}
	return &s
}

func SaveSettings(s *Settings) error {
	_ = os.MkdirAll(Dir(), 0700)
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath(), data, 0600)
}

func cacheKey() string {
	machineID, _ := os.ReadFile("/etc/machine-id")
	if len(machineID) == 0 {
		machineID = []byte("fallback")
	}
	uid := strconv.Itoa(os.Getuid())
	window := strconv.FormatInt(time.Now().Unix()/1800, 10) // 30 min window
	h := sha256.Sum256([]byte(string(machineID) + uid + window))
	return hex.EncodeToString(h[:])
}

func CachePassword(password string) {
	key := cacheKey()
	encrypted, err := vault.Encrypt([]byte(password), key)
	if err != nil {
		return
	}
	_ = os.WriteFile(cachePath(), encrypted, 0600)
}

func GetCachedPassword() string {
	data, err := os.ReadFile(cachePath())
	if err != nil {
		return ""
	}

	info, err := os.Stat(cachePath())
	if err != nil || time.Since(info.ModTime()) > 30*time.Minute {
		ClearPasswordCache()
		return ""
	}

	key := cacheKey()
	decrypted, err := vault.Decrypt(data, key)
	if err != nil {
		ClearPasswordCache()
		return ""
	}

	return string(decrypted)
}

func ClearPasswordCache() {
	os.Remove(cachePath())
}
