package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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

// getServiceState returns the active state and exit code of a service
// Maps rc-service exit codes to systemd states:
// - 0 (running) → "active"
// - 1 (stopped) → "inactive"
// - 3 (crashed) → "failed"
// Returns (state, exitCode, error)
func getServiceState(serviceName string) (string, int, error) {
	// Check if service exists
	if err := checkServiceExists(serviceName); err != nil {
		return "unknown", 1, err
	}

	// Run rc-service status and capture exit code
	cmd := exec.Command("rc-service", serviceName, "status")
	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	var state string

	if len(output) > 0 {
		_, state, _ = strings.Cut(string(output), ": ")
		state = strings.TrimSpace(state)
	}

	// Map exit codes to systemd states
	switch exitCode {
	case 0:
		state = "active"
	case 1, 2:
		state = "failed"
	case 3:
		state = "inactive"
	default:
		if state == "" {
			state = "unknown"
		}
	}

	return state, exitCode, nil
}

// parseOpenRCScript extracts configuration from an OpenRC init script
// Returns a map of configuration keys to values
func parseOpenRCScript(serviceName string) (map[string]string, error) {
	config := make(map[string]string)

	path := filepath.Join("/etc/init.d", serviceName)
	content, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Extract description
		if strings.Contains(line, "description=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				desc := strings.Trim(parts[1], "\"'")
				config["description"] = desc
			}
		}

		// Extract command
		if strings.Contains(line, "command=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				cmd := strings.Trim(parts[1], "\"'")
				config["command"] = cmd
			}
		}

		// Extract command_user
		if strings.Contains(line, "command_user=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				user := strings.Trim(parts[1], "\"'")
				config["command_user"] = user
			}
		}

		// Extract directory
		if strings.Contains(line, "directory=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				dir := strings.Trim(parts[1], "\"'")
				config["directory"] = dir
			}
		}

		// Extract pidfile
		if strings.Contains(line, "pidfile=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				pidfile := strings.Trim(parts[1], "\"'")
				config["pidfile"] = pidfile
			}
		}

		// Extract command_background
		if strings.Contains(line, "command_background=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				bg := strings.Trim(parts[1], "\"'")
				config["command_background"] = bg
			}
		}
	}

	return config, nil
}

// getServiceProperties gathers all properties for a service
func getServiceProperties(serviceName string) (map[string]string, error) {
	properties := make(map[string]string)

	// Basic properties
	properties["Id"] = serviceName + ".service"

	// Check if service exists and is loaded
	loadState := "not-found"
	if err := checkServiceExists(serviceName); err == nil {
		loadState = "loaded"
	}
	properties["LoadState"] = loadState

	// Parse OpenRC script for static properties
	openrcConfig, _ := parseOpenRCScript(serviceName)

	// Get description
	if desc, ok := openrcConfig["description"]; ok {
		properties["Description"] = desc
	} else {
		properties["Description"] = ""
	}

	// Get active state
	state, _, _ := getServiceState(serviceName)
	properties["ActiveState"] = state
	properties["SubState"] = getSubState(state)

	// Get unit file state (enabled/disabled)
	isEnabled, _ := isServiceEnabled(serviceName)
	unitFileState := "disabled"
	if isEnabled {
		unitFileState = "enabled"
	}
	properties["UnitFileState"] = unitFileState

	// Get ExecStart from command
	if cmd, ok := openrcConfig["command"]; ok {
		properties["ExecStart"] = cmd
	}

	// Determine Type based on command_background
	serviceType := "simple"
	if bg, ok := openrcConfig["command_background"]; ok && bg == "true" {
		serviceType = "simple"
	} else {
		serviceType = "forking"
	}
	properties["Type"] = serviceType

	// Get User
	if user, ok := openrcConfig["command_user"]; ok {
		properties["User"] = user
	}

	// Get WorkingDirectory
	if dir, ok := openrcConfig["directory"]; ok {
		properties["WorkingDirectory"] = dir
	}

	// Get PIDFile
	if pidfile, ok := openrcConfig["pidfile"]; ok {
		properties["PIDFile"] = pidfile
	}

	return properties, nil
}

// formatShowOutput formats properties for display with optional filtering
func formatShowOutput(properties map[string]string, requestedProps []string, valueOnly bool) string {
	var output strings.Builder

	// If specific properties requested, filter them
	var propsToShow map[string]string
	if len(requestedProps) > 0 {
		propsToShow = make(map[string]string)
		for _, prop := range requestedProps {
			if val, ok := properties[prop]; ok {
				propsToShow[prop] = val
			}
		}
	} else {
		propsToShow = properties
	}

	// Sort keys for consistent output
	var keys []string
	for k := range propsToShow {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Format output
	for _, key := range keys {
		if valueOnly {
			output.WriteString(propsToShow[key] + "\n")
		} else {
			output.WriteString(key + "=" + propsToShow[key] + "\n")
		}
	}

	return output.String()
}

// getSystemdServiceFiles returns a map of all systemd service files
// Maps service name to file path
func getSystemdServiceFiles() (map[string]string, error) {
	serviceFiles := make(map[string]string)

	locations := []string{
		"/etc/systemd/system",
		"/lib/systemd/system",
	}

	for _, location := range locations {
		// Check if directory exists
		if _, err := os.Stat(location); os.IsNotExist(err) {
			continue
		}

		// Find all .service files
		pattern := filepath.Join(location, "*.service")
		files, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		// Process each service file
		for _, file := range files {
			_, fileName := filepath.Split(file)
			serviceName := strings.TrimSuffix(fileName, ".service")
			serviceFiles[serviceName] = file
		}
	}

	return serviceFiles, nil
}

// getAllServices returns a list of all available services (from systemd and OpenRC)
func getAllServices() ([]string, error) {
	serviceSet := make(map[string]bool)

	// Add systemd services
	systemdFiles, _ := getSystemdServiceFiles()
	for serviceName := range systemdFiles {
		serviceSet[serviceName] = true
	}

	// Add OpenRC services
	openrcDir := "/etc/init.d"
	if files, err := os.ReadDir(openrcDir); err == nil {
		for _, file := range files {
			// Skip directories and hidden files
			if !file.IsDir() && !strings.HasPrefix(file.Name(), ".") {
				serviceSet[file.Name()] = true
			}
		}
	}

	// Convert to sorted slice
	var services []string
	for service := range serviceSet {
		services = append(services, service)
	}
	sort.Strings(services)

	return services, nil
}

// getServiceDescription returns the description of a service from OpenRC or systemd
func getServiceDescription(serviceName string) string {
	// Try to get from OpenRC script
	config, err := parseOpenRCScript(serviceName)
	if err == nil {
		if desc, ok := config["description"]; ok && desc != "" {
			return desc
		}
	}

	// Try to get from systemd file
	systemdFiles, _ := getSystemdServiceFiles()
	if filePath, ok := systemdFiles[serviceName]; ok {
		if content, err := os.ReadFile(filePath); err == nil {
			// Look for Description= line
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "Description=") {
					desc := strings.TrimPrefix(line, "Description=")
					return strings.TrimSpace(desc)
				}
			}
		}
	}

	return ""
}

// getSubState derives a sub-state from the active state
// Maps systemd substates based on main state
func getSubState(activeState string) string {
	switch activeState {
	case "active":
		return "running"
	case "inactive":
		return "dead"
	case "failed":
		return "failed"
	default:
		return "unknown"
	}
}
