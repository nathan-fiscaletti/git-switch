package git

import "fmt"

func Checkout(branch string) error {
	return executeWithStdout("checkout %v", branch)
}

func ExecuteCheckout(cmd string, args ...any) error {
	return executeWithStdout(fmt.Sprintf("checkout %v", cmd), args...)
}
