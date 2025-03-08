package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var daemonReloadCmd = &cobra.Command{
	Use:   "daemon-reload",
	Short: "Reload systemd manager configuration (not needed in OpenRC)",
	Long: `In systemd, daemon-reload is used to reload unit files and restart systemd services.
This command is not needed in OpenRC as service scripts are read directly each time.

This command is provided for compatibility with systemd workflows.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("The 'daemon-reload' command is not needed in OpenRC.")
		fmt.Println("OpenRC reads service scripts directly each time they are used.")
		fmt.Println("No action was performed.")
	},
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(daemonReloadCmd)
}
