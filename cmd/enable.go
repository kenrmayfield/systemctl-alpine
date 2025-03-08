package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"systemctl-alpine/pkg/converter"
	"systemctl-alpine/pkg/parser"
	"systemctl-alpine/pkg/util"

	"github.com/spf13/cobra"
)

var enableCmd = &cobra.Command{
	Use:   "enable [service...]",
	Short: "Enable one or more services to start at boot",
	Long: `Enable one or more services to start at boot by converting systemd service files
to OpenRC init scripts and adding them to the default runlevel.

Example:
  ` + cliName + ` enable nginx
  ` + cliName + ` enable --now nginx mysql redis  # Enable and start multiple services`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, arg := range args {
			if err := enableService(arg); err != nil {
				return fmt.Errorf("failed to enable %s: %w", arg, err)
			}
		}
		return nil
	},
	SilenceUsage: true,
}

func enableService(serviceName string) error {
	// If the service name doesn't end with .service, add it
	serviceFileName := serviceName
	if !strings.HasSuffix(serviceFileName, ".service") {
		serviceFileName = serviceFileName + ".service"
	}

	// Remove .service suffix for the OpenRC script name
	openrcName := util.NormalizeServiceName(serviceName)

	var serviceFile string
	var found bool

	for _, location := range serviceLocations {
		path := filepath.Join(location, serviceFileName)
		if _, err := os.Stat(path); err == nil {
			serviceFile = path
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("service file not found for %s", serviceFileName)
	}

	// Parse the service file
	config, err := parser.ParseServiceFile(serviceFile)
	if err != nil {
		return fmt.Errorf("failed to parse service file: %w", err)
	}

	// Convert to OpenRC
	script, err := converter.ConvertToOpenRC(config, openrcName)
	if err != nil {
		return fmt.Errorf("failed to convert to OpenRC: %w", err)
	}

	// Write the OpenRC script
	if err := converter.WriteOpenRCScript(script, openrcName); err != nil {
		return fmt.Errorf("failed to write OpenRC script: %w", err)
	}

	// Enable the service
	if err := converter.EnableService(openrcName); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	fmt.Printf("Service %s has been converted and enabled\n", openrcName)

	// Start the service if --now flag is provided
	if nowFlag {
		if err := executeServiceCommand(openrcName, "start"); err != nil {
			return fmt.Errorf("failed to start service: %w", err)
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(enableCmd)
	enableCmd.Flags().BoolVar(&nowFlag, "now", false, "Start the service after enabling it")
}
