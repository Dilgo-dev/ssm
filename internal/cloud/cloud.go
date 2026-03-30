package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"ssm/internal/config"
)

type CloudConfig struct {
	Server string `json:"server"`
	Token  string `json:"token"`
}

func cloudPath() string {
	return filepath.Join(config.Dir(), "cloud.json")
}

func LoadCloud() (*CloudConfig, error) {
	data, err := os.ReadFile(cloudPath())
	if err != nil {
		return nil, fmt.Errorf("not logged in (run: ssm login)")
	}
	var cfg CloudConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveCloud(cfg *CloudConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cloudPath(), data, 0600)
}

func DeleteCloud() error {
	return os.Remove(cloudPath())
}

func Register(server, email, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})

	resp, err := http.Post(server+"/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	return parseTokenResponse(resp)
}

func Login(server, email, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})

	resp, err := http.Post(server+"/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	return parseTokenResponse(resp)
}

func Push(cfg *CloudConfig) error {
	data, err := os.ReadFile(config.Path())
	if err != nil {
		return fmt.Errorf("no local vault found")
	}

	req, err := http.NewRequest("PUT", cfg.Server+"/sync", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return parseError(resp)
	}
	return nil
}

func Pull(cfg *CloudConfig) error {
	req, err := http.NewRequest("GET", cfg.Server+"/sync", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return fmt.Errorf("no vault found on server (run: ssm push)")
	}
	if resp.StatusCode != 200 {
		return parseError(resp)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	os.MkdirAll(config.Dir(), 0700)
	return os.WriteFile(config.Path(), data, 0600)
}

func parseTokenResponse(resp *http.Response) (string, error) {
	if resp.StatusCode >= 400 {
		return "", parseError(resp)
	}
	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Token, nil
}

func parseError(resp *http.Response) error {
	var result struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("server error (%d)", resp.StatusCode)
	}
	return fmt.Errorf("%s", result.Error)
}
