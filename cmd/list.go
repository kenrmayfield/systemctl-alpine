package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all systemd services and their OpenRC status",
	Long: `List all systemd services from /lib/systemd/system/ and /etc/systemd/system/
and show whether they are enabled in OpenRC.

Example:
  ` + cliName + ` list
  ` + cliName + ` ls`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get list of enabled services in OpenRC
		enabledServices, err := getEnabledServices()
		if err != nil {
			return fmt.Errorf("failed to get enabled services: %w", err)
		}

		// Track services we've already seen to avoid duplicates
		seenServices := make(map[string]bool)

		// Process each location
		for _, location := range serviceLocations {
			// Check if directory exists
			if _, err := os.Stat(location); os.IsNotExist(err) {
				fmt.Printf("Directory %s does not exist, skipping\n", location)
				continue
			}

			// Find all .service files
			pattern := filepath.Join(location, "*.service")
			files, err := filepath.Glob(pattern)
			if err != nil {
				return fmt.Errorf("failed to list service files in %s: %w", location, err)
			}

			// Process each service file
			for _, file := range files {
				// Extract service name from path
				_, serviceName := filepath.Split(file)
				serviceName = strings.TrimSuffix(serviceName, ".service")

				// Skip if we've already seen this service
				if seenServices[serviceName] {
					continue
				}
				seenServices[serviceName] = true

				// Check if service exists in OpenRC
				openrcPath := filepath.Join("/etc/init.d", serviceName)
				openrcExists := false
				if _, err := os.Stat(openrcPath); err == nil {
					openrcExists = true
				}

				// Determine status
				status := "disabled"
				if enabledServices[serviceName] {
					status = "enabled"
				}

				// Format output
				if openrcExists {
					fmt.Printf("%-30s %s\n", serviceName, status)
				} else {
					fmt.Printf("%-30s %s (not converted)\n", serviceName, status)
				}
			}
		}

		return nil
	},
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(listCmd)
}
