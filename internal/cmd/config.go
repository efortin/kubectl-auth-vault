package cmd

import (
	"fmt"
	"os"

	"github.com/efortin/kubectl-auth-vault/internal/vault"
	"github.com/spf13/cobra"
)

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
	configTestCmd.Flags().StringVar(&testOpts.tokenPath, "token-path", "identity/oidc/token/enablers_kubernetes_admin", "Vault OIDC token path")

	configShowCmd.Flags().StringVar(&showOpts.vaultAddr, "vault-addr", "", "Vault server address (env: VAULT_ADDR)")
	configShowCmd.Flags().StringVar(&showOpts.tokenPath, "token-path", "identity/oidc/token/enablers_kubernetes_admin", "Vault OIDC token path")

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

	fmt.Fprintf(cmd.OutOrStdout(), "Testing Vault configuration...\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  Vault Address: %s\n", vaultAddr)
	fmt.Fprintf(cmd.OutOrStdout(), "  Token Path:    %s\n\n", opts.tokenPath)

	client, err := vault.NewClient(vaultAddr)
	if err != nil {
		return fmt.Errorf("failed to create Vault client: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Fetching OIDC token...\n")

	token, exp, err := client.GetOIDCToken(cmd.Context(), opts.tokenPath)
	if err != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "❌ Failed to fetch token: %v\n", err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✅ Successfully retrieved token!\n\n")
	fmt.Fprintf(cmd.OutOrStdout(), "Token Details:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  Length:     %d characters\n", len(token))
	fmt.Fprintf(cmd.OutOrStdout(), "  Expiration: %d (Unix timestamp)\n", exp)
	fmt.Fprintf(cmd.OutOrStdout(), "  Preview:    %s...\n", token[:min(50, len(token))])

	return nil
}

func runConfigShow(cmd *cobra.Command, opts *configOptions) error {
	vaultAddr := opts.vaultAddr
	if vaultAddr == "" {
		vaultAddr = os.Getenv("VAULT_ADDR")
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Current Configuration:\n\n")
	fmt.Fprintf(cmd.OutOrStdout(), "Environment Variables:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  VAULT_ADDR:  %s\n", envOrDefault("VAULT_ADDR", "(not set)"))
	fmt.Fprintf(cmd.OutOrStdout(), "  VAULT_TOKEN: %s\n", envOrDefault("VAULT_TOKEN", "(not set)"))
	fmt.Fprintf(cmd.OutOrStdout(), "\n")
	fmt.Fprintf(cmd.OutOrStdout(), "Effective Settings:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  Vault Address: %s\n", valueOrDefault(vaultAddr, "(not set)"))
	fmt.Fprintf(cmd.OutOrStdout(), "  Token Path:    %s\n", opts.tokenPath)
	fmt.Fprintf(cmd.OutOrStdout(), "\n")
	fmt.Fprintf(cmd.OutOrStdout(), "Kubeconfig Example:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  users:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  - name: vault-user\n")
	fmt.Fprintf(cmd.OutOrStdout(), "    user:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "      exec:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "        apiVersion: client.authentication.k8s.io/v1\n")
	fmt.Fprintf(cmd.OutOrStdout(), "        command: kubectl-auth_vault\n")
	fmt.Fprintf(cmd.OutOrStdout(), "        interactiveMode: Never\n")
	if vaultAddr != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "        env:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "        - name: VAULT_ADDR\n")
		fmt.Fprintf(cmd.OutOrStdout(), "          value: %s\n", vaultAddr)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "        args:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "        - get\n")
	fmt.Fprintf(cmd.OutOrStdout(), "        - --token-path\n")
	fmt.Fprintf(cmd.OutOrStdout(), "        - %s\n", opts.tokenPath)

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
