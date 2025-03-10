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
	// Check if this is a template service (contains @)
	var templateName string
	var instanceName string

	if strings.Contains(serviceName, "@") {
		parts := strings.SplitN(serviceName, "@", 2)
		templateName = parts[0] + "@"
		instanceName = parts[1]

		// If the service name doesn't end with .service, add it for the template
		if !strings.HasSuffix(templateName, ".service") {
			templateName = templateName + ".service"
		}
	} else {
		// If the service name doesn't end with .service, add it
		templateName = util.NormalizeServiceName(serviceName)
		if !strings.HasSuffix(templateName, ".service") {
			templateName = templateName + ".service"
		}
	}

	// The OpenRC service name will include the instance name if provided
	openrcName := util.NormalizeServiceName(serviceName)

	// Check if the OpenRC service already exists
	openrcPath := filepath.Join("/etc/init.d", openrcName)
	openrcExists := false
	if _, err := os.Stat(openrcPath); err == nil {
		openrcExists = true
	}

	// Look for the systemd service file (using the template name)
	var serviceFile string
	var found bool

	for _, location := range serviceLocations {
		path := filepath.Join(location, templateName)
		if _, err := os.Stat(path); err == nil {
			serviceFile = path
			found = true
			break
		}
	}

	// If OpenRC service exists, check if it has been modified
	if openrcExists {
		content, err := os.ReadFile(openrcPath)
		if err != nil {
			return fmt.Errorf("failed to read service file: %w", err)
		}

		// Check for modification comment
		if strings.Contains(string(content), "# Modified by systemctl edit") && !forceFlag {
			fmt.Printf("Service %s has been manually modified. Use --force to overwrite.\n", openrcName)

			// If systemd service file not found, just enable the existing OpenRC service
			if !found {
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

			// If systemd service file found but we're not forcing, just enable without converting
			fmt.Printf("Skipping conversion due to manual modifications. Enabling existing service.\n")
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
	}

	// If neither systemd nor OpenRC service exists, return an error
	if !found && !openrcExists {
		return fmt.Errorf("service file not found for %s", templateName)
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
		config, err := parser.ParseServiceFile(serviceFile, instanceName)
		if err != nil {
			return fmt.Errorf("failed to parse service file: %w", err)
		}

		// Convert to OpenRC
		script, err := converter.ConvertToOpenRC(config, openrcName, instanceName)
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
	enableCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force overwrite of manually modified service files")
}
