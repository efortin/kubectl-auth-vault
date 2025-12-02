package main

import (
	"os"

	"github.com/efortin/kubectl-auth-vault/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
