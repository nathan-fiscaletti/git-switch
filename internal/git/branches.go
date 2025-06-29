package git

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

func ListRemotes() ([]string, error) {
	out, err := execute("remote")
	if err != nil {
		return nil, err
	}

	return strings.Split(out, "\n"), nil
}

func AllBranches() ([]string, error) {
	remotes, err := ListRemotes()
	if err != nil {
		return nil, err
	}

	out, err := execute("branch -a --format=%%(refname:short)")
	if err != nil {
		return nil, err
	}

	results := strings.Split(out, "\n")

	return lo.Uniq(lo.FilterMap(results, func(b string, _ int) (string, bool) {
		// Don't include branches that are just the remotes themselves.
		if lo.Contains(remotes, b) {
			return "", false
		}

		// Remove the remote from each branch if it's prefixed with it
		for _, remote := range remotes {
			if strings.HasPrefix(b, remote) {
				return strings.TrimPrefix(b, remote+"/"), true
			}
		}

		return b, true
	})), nil
}

func Checkout(branch string) error {
	return executeWithStdout("checkout %v", branch)
}

func ExecuteCheckout(cmd string, args ...any) error {
	return executeWithStdout(fmt.Sprintf("checkout %v", cmd), args...)
}

func GetCurrentBranch() (string, error) {
	res, err := execute("branch %v", "--show-current")
	if err != nil {
		return res, err
	}

	return strings.TrimSpace(res), nil
}
