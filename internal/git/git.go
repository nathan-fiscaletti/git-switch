package git

import (
	"errors"
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

func GetRepositoryPath() (string, error) {
	res, err := executeHide("rev-parse --show-toplevel")
	if err != nil {
		print(res)
		return "", err
	}

	return strings.ReplaceAll(strings.TrimSpace(res), "/", string(os.PathSeparator)), nil
}
