package git

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

func ValidateGitInstallation() error {
	_, err := exec.LookPath("git")
	return err
}

func IsGitRepository(ctx context.Context) (bool, error) {
	res, err := execute(ctx, "rev-parse --is-inside-work-tree")
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(res) == "true", nil
}

func execute(ctx context.Context, format string, args ...any) (string, error) {
	slog.Debug("executing git command", slog.String("command", fmt.Sprintf(format, args...)))

	cmdFormatted := fmt.Sprintf(format, args...)
	cmd := exec.CommandContext(ctx, "git", strings.Split(cmdFormatted, " ")...)
	output, err := cmd.CombinedOutput()

	slog.Debug("git command output", slog.String("command", cmdFormatted), slog.String("output", string(output)))

	return string(output), err
}
