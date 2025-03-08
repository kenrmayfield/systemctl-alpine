package cmd

import (
	"systemctl-alpine/pkg/util"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [service]",
	Short: "Start a service",
	Long: `Start a service using OpenRC's rc-service command.

Example:
  ` + cliName + ` start nginx`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := util.NormalizeServiceName(args[0])

		if err := checkServiceExists(serviceName); err != nil {
			return err
		}

		return executeServiceCommand(serviceName, "start")
	},
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(startCmd)
}
