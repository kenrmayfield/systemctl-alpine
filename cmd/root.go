package cmd

import (
	"github.com/spf13/cobra"
)

var (
	cliName = "systemctl"

	// Locations to search for systemd service files
	serviceLocations = []string{
		"/lib/systemd/system",
		"/etc/systemd/system",
	}

	nowFlag   bool
	allFlag   bool
	forceFlag bool
)

var rootCmd = &cobra.Command{
	Use:   cliName,
	Short: "A tool to translate systemctl commands to OpenRC commands",
	Long: `` + cliName + ` is a CLI tool that helps you manage services on Alpine Linux
by translating systemctl commands to their OpenRC equivalents.

For example, you can use '` + cliName + ` enable some-service' to convert a systemd
service file to an OpenRC init script and enable it to start at boot.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Here you will define your flags and configuration settings
}
