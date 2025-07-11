package git

import (
	"strings"

	"github.com/samber/lo"
)

func PruneRemoteBranches() error {
	remotes, err := ListRemotes()
	if err != nil {
		return err
	}

	for _, remote := range remotes {
		out, err := executeHide("remote prune %v", remote)
		if err != nil {
			print(out)
			return err
		}
	}

	return nil
}

func ListBranches() ([]string, error) {
	remotes, err := ListRemotes()
	if err != nil {
		return nil, err
	}

	out, err := executeHide("branch -a --format=%%(refname:short)")
	if err != nil {
		print(out)
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

func GetCurrentBranch() (string, error) {
	res, err := executeHide("branch %v", "--show-current")
	if err != nil {
		print(res)
		return res, err
	}

	return strings.TrimSpace(res), nil
}
