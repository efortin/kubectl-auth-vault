package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "kubectl-auth_vault",
		Short: "Kubectl plugin for Vault OIDC token authentication",
		Long: `A kubectl credential plugin that fetches OIDC tokens from HashiCorp Vault.
It caches tokens locally and reuses them until expiration.

Usage as kubectl plugin:
  kubectl auth-vault get --token-path identity/oidc/token/my_role

Usage in kubeconfig:
  users:
  - name: vault-user
    user:
      exec:
        apiVersion: client.authentication.k8s.io/v1
        command: kubectl-auth_vault
        args: ["get", "--token-path", "identity/oidc/token/my_role"]`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, BuildDate),
	}

	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	addGetCommand(rootCmd)
	addConfigCommand(rootCmd)
	addVersionCommand(rootCmd)

	return rootCmd
}

func Execute() error {
	return NewRootCmd().Execute()
}
