//go:build !windows

package config

import (
	"os"
	"strconv"
)

func userID() string {
	return strconv.Itoa(os.Getuid())
}

func machineID() []byte {
	data, _ := os.ReadFile("/etc/machine-id")
	return data
}
