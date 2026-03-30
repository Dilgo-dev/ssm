package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"ssm/internal/config"
)

const (
	repo     = "Dilgo-dev/ssm"
	apiURL   = "https://api.github.com/repos/" + repo + "/releases/latest"
	cooldown = 6 * time.Hour
)

func flagPath() string {
	return filepath.Join(config.Dir(), ".update-available")
}

func CheckInBackground(currentVersion string) {
	go func() {
		if !shouldCheck() {
			return
		}
		latest, err := checkLatest()
		if err != nil || latest == "" || latest == currentVersion {
			return
		}
		_ = os.WriteFile(flagPath(), []byte(latest+"\n"+fmt.Sprint(time.Now().Unix())), 0600)
	}()
}

func GetAvailable(currentVersion string) string {
	data, err := os.ReadFile(flagPath())
	if err != nil {
		return ""
	}
	parts := strings.SplitN(string(data), "\n", 2)
	if len(parts) == 0 || parts[0] == currentVersion {
		return ""
	}
	return parts[0]
}

func ClearFlag() {
	_ = os.Remove(flagPath())
}

func Download() error {
	latest, err := checkLatest()
	if err != nil {
		return err
	}
	if latest == "" {
		return fmt.Errorf("no release found")
	}

	goos := runtime.GOOS
	goarch := runtime.GOARCH
	binary := fmt.Sprintf("ssm-%s-%s", goos, goarch)
	url := fmt.Sprintf("https://github.com/%s/releases/latest/download/%s", repo, binary)

	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find current binary: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return err
	}

	tmp := exe + ".new"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		_ = os.Remove(tmp)
		return err
	}

	if err := os.Rename(tmp, exe); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	ClearFlag()
	fmt.Printf("Updated to %s\n", latest)
	return nil
}

func shouldCheck() bool {
	data, err := os.ReadFile(flagPath())
	if err != nil {
		return true
	}
	parts := strings.SplitN(string(data), "\n", 2)
	if len(parts) < 2 {
		return true
	}
	var ts int64
	_, _ = fmt.Sscanf(parts[1], "%d", &ts)
	return time.Since(time.Unix(ts, 0)) > cooldown
}

func checkLatest() (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API: %s", resp.Status)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	return release.TagName, nil
}
