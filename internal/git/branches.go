package git

import (
	"context"
	"strings"

	"github.com/samber/lo"
)

func RemoteBranches(ctx context.Context) ([]string, error) {
	out, err := execute(ctx, "branch -r --format=%%(refname:short)")
	if err != nil {
		return nil, err
	}

	return lo.FilterMap(strings.Split(out, "\n"), func(b string, _ int) (string, bool) {
		if !strings.Contains(b, "/") {
			return "", false
		}

		idx := strings.Index(b, "/")
		if idx == -1 || idx == len(b)-1 {
			return "", false
		}

		return b[idx+1:], true
	}), nil
}

func Checkout(ctx context.Context, branch string) error {
	_, err := execute(ctx, "checkout %v", branch)
	return err
}
