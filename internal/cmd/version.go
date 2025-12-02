package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

func addVersionCommand(rootCmd *cobra.Command) {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  `Display detailed version information including build details.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "kubectl-auth_vault\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Version:    %s\n", Version)
			fmt.Fprintf(cmd.OutOrStdout(), "  Commit:     %s\n", Commit)
			fmt.Fprintf(cmd.OutOrStdout(), "  Built:      %s\n", BuildDate)
			fmt.Fprintf(cmd.OutOrStdout(), "  Go version: %s\n", runtime.Version())
			fmt.Fprintf(cmd.OutOrStdout(), "  OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}

	rootCmd.AddCommand(versionCmd)
}
