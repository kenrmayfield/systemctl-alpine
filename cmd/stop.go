package cmd

import (
	"systemctl-alpine/pkg/util"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [service]",
	Short: "Stop a service",
	Long: `Stop a service using OpenRC's rc-service command.

Example:
  ` + cliName + ` stop nginx`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := util.NormalizeServiceName(args[0])

		if err := checkServiceExists(serviceName); err != nil {
			return err
		}

		return executeServiceCommand(serviceName, "stop")
	},
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
