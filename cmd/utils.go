package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// checkServiceExists verifies that the service exists in /etc/init.d/
func checkServiceExists(serviceName string) error {
	path := filepath.Join("/etc/init.d", serviceName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("service %s does not exist", serviceName)
	}
	return nil
}

// executeServiceCommand runs an rc-service command on the specified service
func executeServiceCommand(serviceName, command string) error {
	titleCaser := cases.Title(language.English)
	fmt.Printf("%s service %s...\n", titleCaser.String(command)+"ing", serviceName)

	cmd := exec.Command("rc-service", serviceName, command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// For status command, don't treat non-zero exit as an error
		if command == "status" {
			fmt.Printf("Service %s might be stopped or has issues\n", serviceName)
			return nil
		}
		return fmt.Errorf("failed to %s service: %w", command, err)
	}

	// Only print success message for non-status commands
	if command != "status" {
		fmt.Printf("Service %s %sed\n", serviceName, command)
	}

	return nil
}

// isServiceEnabled checks if a service is enabled in the default runlevel
func isServiceEnabled(serviceName string) (bool, error) {
	// Run rc-update show default to get the list of enabled services
	rcUpdateCmd := exec.Command("rc-update", "show", "default")
	var out bytes.Buffer
	rcUpdateCmd.Stdout = &out

	if err := rcUpdateCmd.Run(); err != nil {
		return false, fmt.Errorf("failed to get enabled services: %w", err)
	}

	// Check if the service is in the output
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)

		// The output format is: "servicename | default"
		if len(fields) >= 3 && fields[0] == serviceName && fields[1] == "|" && fields[2] == "default" {
			return true, nil
		}
	}

	return false, nil
}

// getEnabledServices returns a map of service names that are enabled in OpenRC
func getEnabledServices() (map[string]bool, error) {
	enabledServices := make(map[string]bool)

	// Run rc-update show default to get the list of enabled services
	cmd := exec.Command("rc-update", "show", "default")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse the output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[1] == "|" && fields[2] == "default" {
			serviceName := fields[0]
			enabledServices[serviceName] = true
		}
	}

	return enabledServices, nil
}
