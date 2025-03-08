package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all systemd services and their OpenRC status",
	Long: `List systemd services and enabled OpenRC services.

By default, shows only enabled OpenRC services and all systemd services.
Use --all or -a to show all services including disabled OpenRC services.

Example:
  ` + cliName + ` list
  ` + cliName + ` ls --all`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get list of enabled services in OpenRC
		enabledServices, err := getEnabledServices()
		if err != nil {
			return fmt.Errorf("failed to get enabled services: %w", err)
		}

		// Track services we've already seen to avoid duplicates
		seenServices := make(map[string]bool)

		// Map to store original systemd paths for services
		systemdPaths := make(map[string]string)

		// First, find all systemd service files and track their paths
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

				// Store the systemd path for this service
				systemdPaths[serviceName] = file

				// Mark as seen
				seenServices[serviceName] = true
			}
		}

		// Now find all OpenRC services
		openrcDir := "/etc/init.d"
		if _, err := os.Stat(openrcDir); err == nil {
			files, err := os.ReadDir(openrcDir)
			if err != nil {
				return fmt.Errorf("failed to list OpenRC services: %w", err)
			}

			for _, file := range files {
				// Skip directories and hidden files
				if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
					continue
				}

				// Add to seen services if not already there
				serviceName := file.Name()

				// Only add OpenRC services that are either enabled or if --all flag is used
				if allFlag || enabledServices[serviceName] || systemdPaths[serviceName] != "" {
					seenServices[serviceName] = true
				}
			}
		}

		// Get a sorted list of service names
		var serviceNames []string
		for serviceName := range seenServices {
			serviceNames = append(serviceNames, serviceName)
		}
		sort.Strings(serviceNames)

		// Now print all services
		fmt.Printf("%-30s %s\n", "SERVICE", "STATUS")
		for _, serviceName := range serviceNames {
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
				// If we have a systemd path, show it's converted
				if originalPath, exists := systemdPaths[serviceName]; exists {
					fmt.Printf("%-30s %s (from %s)\n", serviceName, status, originalPath)
				} else {
					fmt.Printf("%-30s %s\n", serviceName, status)
				}
			} else {
				fmt.Printf("%-30s %s (not converted)\n", serviceName, status)
			}
		}

		return nil
	},
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Show all services including disabled OpenRC services")
}
