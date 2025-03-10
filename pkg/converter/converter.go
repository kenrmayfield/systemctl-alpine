package converter

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"systemctl-alpine/pkg/parser"
)

//go:embed openrc.tpl
var openrcTemplate string

// TemplateData holds the data for the OpenRC template
type TemplateData struct {
	Name                 string
	Description          string
	User                 string
	Group                string
	WorkingDirectory     string
	EnvironmentFile      string
	Environment          []string
	ExecStartPreCommands []string
	Command              string
	CommandArgs          string
	StopCommand          string
	Capabilities         string
	CommandBackground    bool
	SourcePath           string
	InstanceName         string
}

// ConvertToOpenRC converts a systemd service to an OpenRC init script
func ConvertToOpenRC(config *parser.ServiceConfig, serviceName string, instanceName string) (string, error) {
	// Split ExecStart into command and arguments
	execParts := strings.Fields(config.ExecStart)
	if len(execParts) == 0 {
		return "", fmt.Errorf("ExecStart is empty")
	}

	command := execParts[0]
	var commandArgs string
	if len(execParts) > 1 {
		commandArgs = strings.Join(execParts[1:], " ")
	}

	// Process ExecStartPre commands
	var execStartPreCommands []string
	for _, cmd := range config.ExecStartPre {
		// Handle commands that might start with - (which means "ignore errors" in systemd)
		if strings.HasPrefix(cmd, "-") {
			cmd = strings.TrimPrefix(cmd, "-")
			// Wrap command in a conditional to ignore errors
			execStartPreCommands = append(execStartPreCommands, cmd+" || true")
		} else {
			execStartPreCommands = append(execStartPreCommands, cmd)
		}
	}

	// Process ExecStop command if present
	var stopCommand string
	if config.ExecStop != "" {
		stopCommand = config.ExecStop
	}

	// Process AmbientCapabilities if present
	var capabilities string
	if config.AmbientCapabilities != "" {
		// Split space-separated capabilities
		caps := strings.Fields(config.AmbientCapabilities)

		// Convert to OpenRC format (comma-separated with ^ prefix)
		var formattedCaps []string
		for _, c := range caps {
			formattedCaps = append(formattedCaps, "^"+strings.ToLower(c))
		}

		capabilities = strings.Join(formattedCaps, ",")
	}

	// Determine if command should run in background based on Type
	commandBackground := true
	if config.Type == "forking" {
		commandBackground = false
	}
	// For Type=simple, Type=notify, or no Type specified, keep commandBackground=true

	// Prepare template data
	data := TemplateData{
		Name:                 serviceName,
		Description:          config.Description,
		User:                 config.User,
		Group:                config.Group,
		WorkingDirectory:     config.WorkingDirectory,
		EnvironmentFile:      config.EnvironmentFile,
		Environment:          config.Environment,
		ExecStartPreCommands: execStartPreCommands,
		Command:              command,
		CommandArgs:          commandArgs,
		StopCommand:          stopCommand,
		Capabilities:         capabilities,
		CommandBackground:    commandBackground,
		SourcePath:           config.SourcePath,
		InstanceName:         instanceName,
	}

	// Create template
	tmpl, err := template.New("openrc").Parse(openrcTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Render template
	var output strings.Builder
	if err := tmpl.Execute(&output, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Post-process the output to remove multiple empty lines
	return removeEmptyLines(output.String()), nil
}

// WriteOpenRCScript writes the OpenRC init script to the appropriate location
func WriteOpenRCScript(script, serviceName string) error {
	// Ensure the directory exists
	if err := os.MkdirAll("/etc/init.d", 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write the script
	path := filepath.Join("/etc/init.d", serviceName)
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		return fmt.Errorf("failed to write script: %w", err)
	}

	return nil
}

// EnableService adds the service to the default runlevel
func EnableService(serviceName string) error {
	cmd := exec.Command("rc-update", "add", serviceName, "default")
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// removeEmptyLines removes multiple consecutive empty lines and trims space
func removeEmptyLines(input string) string {
	// Split into lines
	lines := strings.Split(input, "\n")

	var result []string
	var prevEmpty bool

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isEmpty := trimmed == ""

		// Keep the line if:
		// 1. It's not empty, or
		// 2. It's empty but the previous line wasn't empty
		if !isEmpty || !prevEmpty {
			// If it's a non-empty line, keep original indentation
			if !isEmpty {
				result = append(result, line)
			} else {
				result = append(result, "")
			}
		}
		prevEmpty = isEmpty
	}

	// Join lines and trim any leading/trailing whitespace
	return strings.TrimSpace(strings.Join(result, "\n")) + "\n"
}
