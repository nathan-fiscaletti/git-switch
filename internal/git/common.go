package git

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

func executeHide(format string, args ...any) (string, error) {
	slog.Debug("executing git command", slog.String("command", fmt.Sprintf(format, args...)))

	cmdFormatted := fmt.Sprintf(format, args...)
	cmd := exec.Command("git", strings.Split(cmdFormatted, " ")...)
	output, err := cmd.CombinedOutput()

	slog.Debug("git command output", slog.String("command", cmdFormatted), slog.String("output", string(output)))

	return string(output), err
}

func executeWithStdout(format string, args ...any) error {
	slog.Debug("executing git command", slog.String("command", fmt.Sprintf(format, args...)))

	cmdFormatted := fmt.Sprintf(format, args...)
	cmd := exec.Command("git", strings.Split(cmdFormatted, " ")...)

	// Forward output to stdout and stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	slog.Debug("git command finished", slog.String("command", cmdFormatted), slog.Any("error", err))

	return err
}
