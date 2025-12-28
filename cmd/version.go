package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Show the version of systemctl-alpine.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%s (alpine translator) version %s\n", cliName, Version)
		return nil
	},
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
