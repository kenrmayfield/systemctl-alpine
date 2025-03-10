package parser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"systemctl-alpine/pkg/util"
)

// ServiceConfig represents a parsed systemd service file
type ServiceConfig struct {
	Description         string
	User                string
	Group               string
	WorkingDirectory    string
	EnvironmentFile     string
	Environment         []string
	ExecStartPre        []string
	ExecStart           string
	ExecStop            string
	Restart             string
	RestartSec          string
	WantedBy            string
	AmbientCapabilities string
	Type                string
	SourcePath          string
}

// ParseServiceFile parses a systemd service file and returns a ServiceConfig
func ParseServiceFile(path string, instanceName string) (*ServiceConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open service file: %w", err)
	}
	defer file.Close()

	name := filepath.Base(path)
	name = util.NormalizeServiceName(name)

	config := &ServiceConfig{
		SourcePath: path,
	}
	scanner := bufio.NewScanner(file)

	var section string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check if this is a section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = line[1 : len(line)-1]
			continue
		}

		// Parse key-value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Process template substitutions if this is a template unit
		if instanceName != "" {
			value = ProcessTemplateSubstitutions(value, name, instanceName)
		}

		switch section {
		case "Unit":
			if key == "Description" {
				config.Description = value
			}
		case "Service":
			switch key {
			case "User":
				config.User = value
			case "Group":
				config.Group = value
			case "WorkingDirectory":
				config.WorkingDirectory = value
			case "EnvironmentFile":
				// Remove the leading dash if present (indicates optional)
				if strings.HasPrefix(value, "-") {
					value = strings.TrimPrefix(value, "-")
				}
				config.EnvironmentFile = value
			case "Environment":
				// Remove quotes if present
				if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
					value = value[1 : len(value)-1]
				}
				config.Environment = append(config.Environment, value)
			case "ExecStartPre":
				config.ExecStartPre = append(config.ExecStartPre, value)
			case "ExecStart":
				config.ExecStart = value
			case "ExecStop":
				config.ExecStop = value
			case "Restart":
				config.Restart = value
			case "RestartSec":
				config.RestartSec = value
			case "AmbientCapabilities":
				config.AmbientCapabilities = value
			case "Type":
				config.Type = value
			}
		case "Install":
			if key == "WantedBy" {
				config.WantedBy = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading service file: %w", err)
	}

	return config, nil
}

// ProcessTemplateSubstitutions replaces template specifiers in a string with their values
func ProcessTemplateSubstitutions(input string, unitName string, instanceName string) string {
	// Extract prefix (part before @)
	var prefix string
	if strings.Contains(unitName, "@") {
		prefix = strings.Split(unitName, "@")[0]
	} else {
		prefix = unitName
	}

	// Get system information for substitutions
	hostname, _ := os.Hostname()
	shortHostname := strings.Split(hostname, ".")[0]

	// Get architecture
	arch := runtime.GOARCH
	switch arch {
	case "arm64":
		arch = "aarch64"
	case "amd64":
		arch = "x86_64"
	case "386":
		arch = "x86"
	}

	// Get machine ID if available
	machineID := ""
	if data, err := os.ReadFile("/etc/machine-id"); err == nil {
		machineID = strings.TrimSpace(string(data))
	}

	// Get OS ID if available
	osID := ""
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "ID=") {
				osID = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
				break
			}
		}
	}

	// Perform substitutions
	result := input
	result = strings.ReplaceAll(result, "%%", "%%PLACEHOLDER%%")
	result = strings.ReplaceAll(result, "%a", arch)
	result = strings.ReplaceAll(result, "%i", instanceName)
	result = strings.ReplaceAll(result, "%I", unescapeValue(instanceName))
	result = strings.ReplaceAll(result, "%l", shortHostname)
	result = strings.ReplaceAll(result, "%m", machineID)
	result = strings.ReplaceAll(result, "%n", unitName)
	result = strings.ReplaceAll(result, "%N", strings.TrimSuffix(unitName, ".service"))
	result = strings.ReplaceAll(result, "%o", osID)
	result = strings.ReplaceAll(result, "%p", prefix)
	result = strings.ReplaceAll(result, "%P", unescapeValue(prefix))
	result = strings.ReplaceAll(result, "%%PLACEHOLDER%%", "%")

	return result
}

// unescapeValue undoes systemd escaping
func unescapeValue(value string) string {
	// Systemd escaping rules: replace "-" with "/", "\x2d" with "-"
	result := strings.ReplaceAll(value, "-", "/")
	result = strings.ReplaceAll(result, "\\x2d", "-")
	return result
}
