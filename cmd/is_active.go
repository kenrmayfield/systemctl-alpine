package cmd

import (
	"fmt"
	"os"

	"systemctl-alpine/pkg/util"

	"github.com/spf13/cobra"
)

var isActiveCmd = &cobra.Command{
	Use:   "is-active [service]",
	Short: "Check if a service is currently active (running)",
	Long: `Check if a service is currently active (running).

Prints the active state of the service to stdout (active, inactive, or failed).
Returns exit code 0 if the service is active, non-zero otherwise.

Example:
  ` + cliName + ` is-active nginx`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := util.NormalizeServiceName(args[0])

		if err := checkServiceExists(serviceName); err != nil {
			return err
		}

		state, exitCode, err := getServiceState(serviceName)
		if err != nil {
			return err
		}

		fmt.Println(state)

		// Exit with the same code as rc-service for scriptability
		if exitCode != 0 {
			os.Exit(exitCode)
		}

		return nil
	},
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(isActiveCmd)
}
