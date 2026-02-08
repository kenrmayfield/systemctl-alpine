package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var (
	typeFilter  string
	stateFilter string
)

var listUnitFilesCmd = &cobra.Command{
	Use:   "list-unit-files",
	Short: "List all installed unit files and their enablement state",
	Long: `List all installed systemd unit files and OpenRC services with their enablement state.

The output shows two columns: UNIT FILE (name) and STATE (enabled/disabled/static).

States:
  enabled  - Service is configured to start at boot
  disabled - Service exists but is not configured to start at boot
  static   - Service is OpenRC-only with no systemd unit file

Use --type and --state to filter the results.

Example:
  ` + cliName + ` list-unit-files
  ` + cliName + ` list-unit-files --type=service
  ` + cliName + ` list-unit-files --state=enabled
  ` + cliName + ` list-unit-files --type=service --state=enabled`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listUnitFiles()
	},
	SilenceUsage: true,
}

func listUnitFiles() error {
	unitFiles := make(map[string]string) // serviceName -> state

	// Get all systemd service files
	systemdFiles, _ := getSystemdServiceFiles()
	for serviceName := range systemdFiles {
		isEnabled, _ := isServiceEnabled(serviceName)
		state := "disabled"
		if isEnabled {
			state = "enabled"
		}
		unitFiles[serviceName] = state
	}

	// Add OpenRC-only services (static state)
	openrcDir := "/etc/init.d"
	if files, err := os.ReadDir(openrcDir); err == nil {
		for _, file := range files {
			// Skip directories and hidden files
			if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
				continue
			}

			serviceName := file.Name()

			// Only add if not already in systemd files
			if _, exists := systemdFiles[serviceName]; !exists {
				unitFiles[serviceName] = "static"
			}
		}
	}

	// Apply filters
	var filteredFiles []string
	for serviceName, state := range unitFiles {
		// Apply type filter (only service type supported)
		if typeFilter != "" && typeFilter != "service" {
			continue
		}

		// Apply state filter
		if stateFilter != "" && state != stateFilter {
			continue
		}

		filteredFiles = append(filteredFiles, serviceName)
	}

	// Sort alphabetically
	sort.Strings(filteredFiles)

	// Print header
	fmt.Printf("%-50s %s\n", "UNIT FILE", "STATE")

	// Print each unit file
	for _, serviceName := range filteredFiles {
		unitName := serviceName + ".service"
		state := unitFiles[serviceName]
		fmt.Printf("%-50s %s\n", unitName, state)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(listUnitFilesCmd)
	listUnitFilesCmd.Flags().StringVar(&typeFilter, "type", "", "Filter by unit type (service)")
	listUnitFilesCmd.Flags().StringVar(&stateFilter, "state", "", "Filter by state (enabled, disabled, static)")
}
