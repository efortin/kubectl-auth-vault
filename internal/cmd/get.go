package cmd

import (
	"fmt"
	"os"

	"github.com/efortin/kubectl-auth-vault/internal/cache"
	"github.com/efortin/kubectl-auth-vault/internal/credential"
	"github.com/efortin/kubectl-auth-vault/internal/vault"
	"github.com/spf13/cobra"
)

type getOptions struct {
	vaultAddr string
	tokenPath string
	cacheFile string
	noCache   bool
}

func addGetCommand(rootCmd *cobra.Command) {
	opts := &getOptions{}

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get an OIDC token from Vault and output ExecCredential",
		Long: `Fetches an OIDC token from HashiCorp Vault and outputs it in the
ExecCredential format expected by kubectl.

The token is cached locally and reused until expiration.`,
		Example: `  # Using environment variable for vault address
  export VAULT_ADDR=https://vault.example.com
  kubectl-auth_vault get --token-path identity/oidc/token/my_role

  # Using flags
  kubectl-auth_vault get --vault-addr https://vault.example.com --token-path identity/oidc/token/my_role

  # Disable caching
  kubectl-auth_vault get --token-path identity/oidc/token/my_role --no-cache`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, opts)
		},
	}

	getCmd.Flags().StringVar(&opts.vaultAddr, "vault-addr", "", "Vault server address (env: VAULT_ADDR)")
	getCmd.Flags().StringVar(&opts.tokenPath, "token-path", "identity/oidc/token/enablers_kubernetes_admin", "Vault OIDC token path")
	getCmd.Flags().StringVar(&opts.cacheFile, "cache-file", "", "Token cache file path (default: ~/.kube/vault_<sanitized_path>_token.json)")
	getCmd.Flags().BoolVar(&opts.noCache, "no-cache", false, "Disable token caching")

	rootCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, opts *getOptions) error {
	vaultAddr := opts.vaultAddr
	if vaultAddr == "" {
		vaultAddr = os.Getenv("VAULT_ADDR")
	}
	if vaultAddr == "" {
		return fmt.Errorf("VAULT_ADDR is required (use --vault-addr or VAULT_ADDR env var)")
	}

	cacheFile := opts.cacheFile
	if cacheFile == "" {
		cacheFile = cache.DefaultCacheFile(opts.tokenPath)
	}

	tokenCache := cache.New(cacheFile)

	if !opts.noCache {
		if token, ok := tokenCache.Load(); ok {
			return credential.Output(cmd.OutOrStdout(), token)
		}
	}

	client, err := vault.NewClient(vaultAddr)
	if err != nil {
		return fmt.Errorf("failed to create Vault client: %w", err)
	}

	token, exp, err := client.GetOIDCToken(cmd.Context(), opts.tokenPath)
	if err != nil {
		return fmt.Errorf("failed to fetch token from Vault: %w", err)
	}

	if !opts.noCache {
		if err := tokenCache.Save(token, exp); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: failed to cache token: %v\n", err)
		}
	}

	return credential.Output(cmd.OutOrStdout(), token)
}
