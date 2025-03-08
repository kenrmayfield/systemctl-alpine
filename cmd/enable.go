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
	serviceFileName := util.NormalizeServiceName(serviceName)
	if !strings.HasSuffix(serviceFileName, ".service") {
		serviceFileName = serviceFileName + ".service"
	}

	openrcName := util.NormalizeServiceName(serviceName)

	// Check if the OpenRC service already exists
	openrcPath := filepath.Join("/etc/init.d", openrcName)
	openrcExists := false
	if _, err := os.Stat(openrcPath); err == nil {
		openrcExists = true
	}

	// Look for the systemd service file
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

	// If neither systemd nor OpenRC service exists, return an error
	if !found && !openrcExists {
		return fmt.Errorf("service file not found for %s", serviceFileName)
	}

	// If systemd service file not found but OpenRC service exists, just enable the OpenRC service
	if !found {
		fmt.Printf("No systemd service file found for %s, but OpenRC service exists. Enabling existing service.\n", serviceName)

		// Enable the service
		if err := converter.EnableService(openrcName); err != nil {
			return err
		}

		fmt.Printf("Service %s has been enabled\n", openrcName)

		// Start the service if --now flag is provided
		if nowFlag {
			if err := executeServiceCommand(openrcName, "start"); err != nil {
				return fmt.Errorf("failed to start service: %w", err)
			}
		}

		return nil
	} else {
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

		fmt.Printf("Service %s has been converted to OpenRC\n", openrcName)
	}

	// Enable the service
	if err := converter.EnableService(openrcName); err != nil {
		return err
	}

	fmt.Printf("Service %s has been enabled\n", openrcName)

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
