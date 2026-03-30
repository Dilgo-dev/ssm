//go:build windows

package ssh

import "os"

func notifyResize(ch chan os.Signal) {}
