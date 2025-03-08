package cmd

import (
	"fmt"
	"os/exec"

	"systemctl-alpine/pkg/util"

	"github.com/spf13/cobra"
)

var disableCmd = &cobra.Command{
	Use:   "disable [service...]",
	Short: "Disable one or more services from starting at boot",
	Long: `Disable one or more services from starting at boot by removing them from the default runlevel.

Example:
  ` + cliName + ` disable nginx
  ` + cliName + ` disable --now nginx mysql redis  # Stop and disable multiple services`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, arg := range args {
			if err := disableService(arg); err != nil {
				return fmt.Errorf("failed to disable %s: %w", arg, err)
			}
		}
		return nil
	},
	SilenceUsage: true,
}

func disableService(serviceName string) error {
	serviceName = util.NormalizeServiceName(serviceName)

	if err := checkServiceExists(serviceName); err != nil {
		return err
	}

	// Stop the service if --now flag is provided
	if nowFlag {
		if err := executeServiceCommand(serviceName, "stop"); err != nil {
			return err
		}
	}

	// Disable the service
	disableCmd := exec.Command("rc-update", "del", serviceName, "default")
	if err := disableCmd.Run(); err != nil {
		return fmt.Errorf("failed to disable service: %w", err)
	}

	fmt.Printf("Service %s has been disabled\n", serviceName)
	return nil
}

func init() {
	rootCmd.AddCommand(disableCmd)
	disableCmd.Flags().BoolVar(&nowFlag, "now", false, "Stop the service before disabling it")
}
