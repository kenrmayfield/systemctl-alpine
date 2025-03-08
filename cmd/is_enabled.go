package cmd

import (
	"fmt"
	"os"

	"systemctl-alpine/pkg/util"

	"github.com/spf13/cobra"
)

var isEnabledCmd = &cobra.Command{
	Use:   "is-enabled [service]",
	Short: "Check if a service is enabled to start at boot",
	Long: `Check if a service is enabled to start at boot by examining the default runlevel.

Example:
  ` + cliName + ` is-enabled nginx`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := util.NormalizeServiceName(args[0])

		if err := checkServiceExists(serviceName); err != nil {
			return err
		}

		isEnabled, err := isServiceEnabled(serviceName)
		if err != nil {
			return err
		}

		if isEnabled {
			fmt.Println("enabled")
			return nil
		} else {
			fmt.Println("disabled")
			os.Exit(1)
		}
		return nil
	},
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(isEnabledCmd)
}
