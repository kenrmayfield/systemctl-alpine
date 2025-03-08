package cmd

import (
	"systemctl-alpine/pkg/util"

	"github.com/spf13/cobra"
)

var reloadCmd = &cobra.Command{
	Use:   "reload [service]",
	Short: "Reload a service",
	Long: `Reload a service using OpenRC's rc-service command.
This reloads the service configuration without stopping the service.

Example:
  ` + cliName + ` reload nginx`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := util.NormalizeServiceName(args[0])

		if err := checkServiceExists(serviceName); err != nil {
			return err
		}

		return executeServiceCommand(serviceName, "reload")
	},
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(reloadCmd)
}
