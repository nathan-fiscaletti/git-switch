package git

import "strings"

func ListRemotes() ([]string, error) {
	out, err := executeHide("remote")
	if err != nil {
		print(out)
		return nil, err
	}

	return strings.Split(out, "\n"), nil
}
