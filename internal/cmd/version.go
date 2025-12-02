package cmd

import (
	"runtime"

	"github.com/spf13/cobra"
)

func addVersionCommand(rootCmd *cobra.Command) {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  `Display detailed version information including build details.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("kubectl-auth_vault")
			cmd.Printf("  Version:    %s\n", Version)
			cmd.Printf("  Commit:     %s\n", Commit)
			cmd.Printf("  Built:      %s\n", BuildDate)
			cmd.Printf("  Go version: %s\n", runtime.Version())
			cmd.Printf("  OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}

	rootCmd.AddCommand(versionCmd)
}
