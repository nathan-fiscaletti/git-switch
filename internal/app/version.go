package app

import (
	"debug/buildinfo"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func GetVersionString() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %v", err)
	}

	info, err := buildinfo.ReadFile(exe)
	if err != nil {
		return "", fmt.Errorf("failed to read build info: %v", err)
	}

	cmd := strings.TrimSuffix(filepath.Base(exe), ".exe")

	return fmt.Sprintf("%v version %v %s/%s", cmd, info.Main.Version, runtime.GOOS, runtime.GOARCH), nil
}
