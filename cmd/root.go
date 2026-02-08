package cmd

import (
	"github.com/spf13/cobra"
)

var (
	cliName = "systemctl"

	// Version is the version string, embedded at build time
	Version = "dev"

	// Locations to search for systemd service files
	serviceLocations = []string{
		"/etc/systemd/system",
		"/lib/systemd/system",
	}

	nowFlag   bool
	allFlag   bool
	forceFlag bool
)

var rootCmd = &cobra.Command{
	Use:   cliName,
	Short: "Helper tool to translate systemctl commands to OpenRC commands",
	Long: `` + cliName + ` is a CLI tool that helps you manage services on Alpine Linux
by translating systemctl commands to their OpenRC equivalents.

For example, you can use '` + cliName + ` enable some-service' to convert a systemd
service file to an OpenRC init script and enable it to start at boot.`,
	Version: getVersion(),
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

// getVersion returns the version string for this application
func getVersion() string {
	// This will be populated from version.go
	return Version
}

func init() {
	// Here you will define your flags and configuration settings
	rootCmd.CompletionOptions.DisableDefaultCmd = false
}
