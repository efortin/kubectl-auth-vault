package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/efortin/kubectl-auth-vault/internal/vault"
)

// printf is a helper that uses cobra's Printf (ignores write errors for CLI output).
func printf(cmd *cobra.Command, format string, args ...interface{}) {
	cmd.Printf(format, args...)
}

type configOptions struct {
	vaultAddr string
	tokenPath string
}

func addConfigCommand(rootCmd *cobra.Command) {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Test and validate Vault configuration",
		Long: `Test the Vault configuration by attempting to fetch a token.
This is useful for validating your setup before using it in kubeconfig.`,
	}

	testOpts := &configOptions{}
	configTestCmd := &cobra.Command{
		Use:   "test",
		Short: "Test Vault connectivity and token retrieval",
		Long: `Attempts to connect to Vault and fetch an OIDC token.
Displays detailed information about the token on success.`,
		Example: `  # Test with environment variable
  export VAULT_ADDR=https://vault.example.com
  kubectl-auth_vault config test --token-path identity/oidc/token/my_role

  # Test with explicit vault address
  kubectl-auth_vault config test --vault-addr https://vault.example.com --token-path identity/oidc/token/my_role`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigTest(cmd, testOpts)
		},
	}

	showOpts := &configOptions{}
	configShowCmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  `Display the current configuration including environment variables and defaults.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigShow(cmd, showOpts)
		},
	}

	configTestCmd.Flags().StringVar(&testOpts.vaultAddr, "vault-addr", "", "Vault server address (env: VAULT_ADDR)")
	configTestCmd.Flags().StringVar(&testOpts.tokenPath, "token-path", "identity/oidc/token/kubernetes", "Vault OIDC token path")

	configShowCmd.Flags().StringVar(&showOpts.vaultAddr, "vault-addr", "", "Vault server address (env: VAULT_ADDR)")
	configShowCmd.Flags().StringVar(&showOpts.tokenPath, "token-path", "identity/oidc/token/kubernetes", "Vault OIDC token path")

	configCmd.AddCommand(configTestCmd)
	configCmd.AddCommand(configShowCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigTest(cmd *cobra.Command, opts *configOptions) error {
	vaultAddr := opts.vaultAddr
	if vaultAddr == "" {
		vaultAddr = os.Getenv("VAULT_ADDR")
	}
	if vaultAddr == "" {
		return fmt.Errorf("VAULT_ADDR is required (use --vault-addr or VAULT_ADDR env var)")
	}

	printf(cmd, "Testing Vault configuration...\n")
	printf(cmd, "  Vault Address: %s\n", vaultAddr)
	printf(cmd, "  Token Path:    %s\n\n", opts.tokenPath)

	client, err := vault.NewClient(vaultAddr)
	if err != nil {
		return fmt.Errorf("failed to create Vault client: %w", err)
	}

	printf(cmd, "Fetching OIDC token...\n")

	token, exp, err := client.GetOIDCToken(cmd.Context(), opts.tokenPath)
	if err != nil {
		printf(cmd, "❌ Failed to fetch token: %v\n", err)
		return err
	}

	printf(cmd, "✅ Successfully retrieved token!\n\n")
	printf(cmd, "Token Details:\n")
	printf(cmd, "  Length:     %d characters\n", len(token))
	printf(cmd, "  Expiration: %d (Unix timestamp)\n", exp)
	printf(cmd, "  Preview:    %s...\n", token[:min(50, len(token))])

	return nil
}

func runConfigShow(cmd *cobra.Command, opts *configOptions) error {
	vaultAddr := opts.vaultAddr
	if vaultAddr == "" {
		vaultAddr = os.Getenv("VAULT_ADDR")
	}

	printf(cmd, "Current Configuration:\n\n")
	printf(cmd, "Environment Variables:\n")
	printf(cmd, "  VAULT_ADDR:  %s\n", envOrDefault("VAULT_ADDR", "(not set)"))
	printf(cmd, "  VAULT_TOKEN: %s\n", envOrDefault("VAULT_TOKEN", "(not set)"))
	printf(cmd, "\n")
	printf(cmd, "Effective Settings:\n")
	printf(cmd, "  Vault Address: %s\n", valueOrDefault(vaultAddr, "(not set)"))
	printf(cmd, "  Token Path:    %s\n", opts.tokenPath)
	printf(cmd, "\n")
	printf(cmd, "Kubeconfig Example:\n")
	printf(cmd, "  users:\n")
	printf(cmd, "  - name: vault-user\n")
	printf(cmd, "    user:\n")
	printf(cmd, "      exec:\n")
	printf(cmd, "        apiVersion: client.authentication.k8s.io/v1\n")
	printf(cmd, "        command: kubectl-auth_vault\n")
	printf(cmd, "        interactiveMode: Never\n")
	if vaultAddr != "" {
		printf(cmd, "        env:\n")
		printf(cmd, "        - name: VAULT_ADDR\n")
		printf(cmd, "          value: %s\n", vaultAddr)
	}
	printf(cmd, "        args:\n")
	printf(cmd, "        - get\n")
	printf(cmd, "        - --token-path\n")
	printf(cmd, "        - %s\n", opts.tokenPath)

	return nil
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		if key == "VAULT_TOKEN" {
			return "(set, hidden)"
		}
		return v
	}
	return def
}

func valueOrDefault(val, def string) string {
	if val != "" {
		return val
	}
	return def
}
