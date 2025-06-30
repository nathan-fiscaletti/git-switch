package git

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

var (
	ErrNotARepository = errors.New("not a git repository")
)

func ValidateGitInstallation() error {
	_, err := exec.LookPath("git")
	return err
}

func IsGitRepository() (bool, error) {
	res, err := executeHide("rev-parse --is-inside-work-tree")
	if err != nil {
		print(res)
		return false, err
	}

	return strings.TrimSpace(res) == "true", nil
}

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
