package cmd

import (
	"fmt"
	"runtime"

	"systemctl-alpine/pkg/util"

	"github.com/spf13/cobra"
)

var (
	propertyFlags []string
	valueOnlyFlag bool
)

var showCmd = &cobra.Command{
	Use:   "show [service]",
	Short: "Show properties of a service or the service manager",
	Long: `Show low-level properties of a service in key=value format.

If no service is specified, shows properties of the systemd manager itself.

Use -p/--property to show specific properties, and --value to show only values.

Example:
  ` + cliName + ` show nginx
  ` + cliName + ` show nginx -p ActiveState
  ` + cliName + ` show nginx --property=ActiveState --property=UnitFileState
  ` + cliName + ` show nginx --property=ActiveState --value`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// Show manager properties when no service specified
			return showManagerProperties()
		}

		serviceName := util.NormalizeServiceName(args[0])

		if err := checkServiceExists(serviceName); err != nil {
			return err
		}

		properties, err := getServiceProperties(serviceName)
		if err != nil {
			return fmt.Errorf("failed to get service properties: %w", err)
		}

		output := formatShowOutput(properties, propertyFlags, valueOnlyFlag)
		fmt.Print(output)

		return nil
	},
	SilenceUsage: true,
}

// showManagerProperties displays properties of the systemd manager itself
func showManagerProperties() error {
	properties := make(map[string]string)

	// Basic manager properties
	properties["Version"] = "230" // Fake version (Alpine uses OpenRC)
	properties["Architecture"] = runtime.GOARCH
	properties["DefaultStandardOutput"] = "stdout"
	properties["DefaultStandardError"] = "inherit"

	output := formatShowOutput(properties, propertyFlags, valueOnlyFlag)
	fmt.Print(output)

	return nil
}

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.Flags().StringSliceVarP(&propertyFlags, "property", "p", []string{}, "Show specific properties (can be used multiple times)")
	showCmd.Flags().BoolVar(&valueOnlyFlag, "value", false, "Show only values, not keys")
}
