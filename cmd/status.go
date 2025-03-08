package cmd

import (
	"systemctl-alpine/pkg/util"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [service]",
	Short: "Check the status of a service",
	Long: `Check the status of a service using OpenRC's rc-service command.

Example:
  ` + cliName + ` status nginx`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := util.NormalizeServiceName(args[0])

		if err := checkServiceExists(serviceName); err != nil {
			return err
		}

		return executeServiceCommand(serviceName, "status")
	},
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
