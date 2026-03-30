//go:build windows

package config

import (
	"os"
	"os/exec"
	"strings"
)

func userID() string {
	return os.Getenv("USERNAME")
}

func machineID() []byte {
	out, err := exec.Command("reg", "query",
		`HKLM\SOFTWARE\Microsoft\Cryptography`,
		"/v", "MachineGuid").Output()
	if err != nil {
		return []byte("windows")
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "MachineGuid") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				return []byte(parts[len(parts)-1])
			}
		}
	}
	return []byte("windows")
}
