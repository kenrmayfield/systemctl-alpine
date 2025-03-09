package parser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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
}

// ParseServiceFile parses a systemd service file and returns a ServiceConfig
func ParseServiceFile(path string) (*ServiceConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open service file: %w", err)
	}
	defer file.Close()

	name := filepath.Base(path)
	name = util.NormalizeServiceName(name)

	config := &ServiceConfig{}
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
