package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

var (
	unitsAllFlag bool
	unitsType    string
	unitsState   string
)

var listUnitsCmd = &cobra.Command{
	Use:   "list-units",
	Short: "List loaded systemd units",
	Long: `List currently loaded systemd units with their status.

By default, shows only enabled or currently running services.
Use --all to show all units including disabled and inactive ones.

The output shows five columns:
  UNIT     - The unit name
  LOAD     - Load state (loaded, not-found)
  ACTIVE   - General activation state (active, inactive, failed)
  SUB      - Low-level activation sub-state (running, dead, exited, failed)
  DESCRIPTION - Unit description

Use --all, --type, and --state to filter the results.

Example:
  ` + cliName + ` list-units
  ` + cliName + ` list-units --all
  ` + cliName + ` list-units --type=service
  ` + cliName + ` list-units --state=active
  ` + cliName + ` list-units -a --type=service`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listUnits()
	},
	SilenceUsage: true,
}

func listUnits() error {
	allServices, err := getAllServices()
	if err != nil {
		return fmt.Errorf("failed to get services: %w", err)
	}

	// Get enabled services for default filtering
	enabledServices, _ := getEnabledServices()

	var unitsToShow []map[string]string

	for _, serviceName := range allServices {
		// Skip if service doesn't exist in /etc/init.d
		if err := checkServiceExists(serviceName); err != nil {
			// Only include services from systemd files if --all is set
			systemdFiles, _ := getSystemdServiceFiles()
			if _, exists := systemdFiles[serviceName]; !exists {
				continue
			}
		}

		// Apply filters when --all is not specified
		if !unitsAllFlag {
			// Only show enabled or currently active services
			isEnabled := enabledServices[serviceName]
			state, _, _ := getServiceState(serviceName)
			if !isEnabled && state != "active" {
				continue
			}
		}

		// Apply type filter (only service type supported)
		if unitsType != "" && unitsType != "service" {
			continue
		}

		// Get unit information
		loadState := "loaded"
		if err := checkServiceExists(serviceName); err != nil {
			loadState = "not-found"
		}

		activeState, _, _ := getServiceState(serviceName)
		subState := getSubState(activeState)
		description := getServiceDescription(serviceName)

		// Apply state filter
		if unitsState != "" {
			// Check if this state matches the filter
			if !stateMatches(activeState, subState, unitsState) {
				continue
			}
		}

		unit := map[string]string{
			"unit":        serviceName + ".service",
			"load":        loadState,
			"active":      activeState,
			"sub":         subState,
			"description": description,
		}

		unitsToShow = append(unitsToShow, unit)
	}

	// Sort by unit name
	sort.Slice(unitsToShow, func(i, j int) bool {
		return unitsToShow[i]["unit"] < unitsToShow[j]["unit"]
	})

	// Print header
	fmt.Printf("%-35s %-10s %-7s %-7s %s\n", "UNIT", "LOAD", "ACTIVE", "SUB", "DESCRIPTION")

	// Print each unit
	for _, unit := range unitsToShow {
		fmt.Printf("%-35s %-10s %-7s %-7s %s\n",
			unit["unit"],
			unit["load"],
			unit["active"],
			unit["sub"],
			unit["description"],
		)
	}

	return nil
}

// stateMatches checks if the given active and sub states match the filter
func stateMatches(activeState, subState, filter string) bool {
	// Filter can be an active state (active, inactive, failed) or a sub-state (running, dead, etc.)
	if activeState == filter || subState == filter {
		return true
	}

	// Handle common aliases
	switch filter {
	case "running":
		return activeState == "active" && subState == "running"
	case "dead":
		return activeState == "inactive" && subState == "dead"
	case "active":
		return activeState == "active"
	case "inactive":
		return activeState == "inactive"
	case "failed":
		return activeState == "failed" || subState == "failed"
	}

	return false
}

func init() {
	rootCmd.AddCommand(listUnitsCmd)
	listUnitsCmd.Flags().BoolVarP(&unitsAllFlag, "all", "a", false, "Show all units including disabled and inactive ones")
	listUnitsCmd.Flags().StringVar(&unitsType, "type", "", "Filter by unit type (service)")
	listUnitsCmd.Flags().StringVar(&unitsState, "state", "", "Filter by active state (active, inactive, failed, running, dead, etc.)")
}
