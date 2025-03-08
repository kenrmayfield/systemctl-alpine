package cmd

import (
	"systemctl-alpine/pkg/util"

	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart [service]",
	Short: "Restart a service",
	Long: `Restart a service using OpenRC's rc-service command.

Example:
  ` + cliName + ` restart nginx`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := util.NormalizeServiceName(args[0])

		if err := checkServiceExists(serviceName); err != nil {
			return err
		}

		return executeServiceCommand(serviceName, "restart")
	},
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
