package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"systemctl-alpine/pkg/util"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [service]",
	Short: "Edit an OpenRC service script",
	Long: `Edit an OpenRC service script using the system's available editor.

Example:
  ` + cliName + ` edit nginx`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := util.NormalizeServiceName(args[0])

		if err := checkServiceExists(serviceName); err != nil {
			return err
		}

		return editService(serviceName)
	},
	SilenceUsage: true,
}

func editService(serviceName string) error {
	// Path to the OpenRC service script
	scriptPath := filepath.Join("/etc/init.d", serviceName)

	// Find an available editor
	editor := findEditor()
	if editor == "" {
		return fmt.Errorf("no editor found. Please install vi, nano, or ed")
	}

	// Read the current content of the file
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read service file: %w", err)
	}

	// Open the file in the editor
	cmd := exec.Command(editor, scriptPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	// Read the file again to see if it was modified
	newContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read service file after editing: %w", err)
	}

	// If the file was modified, add a modification comment
	if string(content) != string(newContent) {
		timestamp := time.Now().Format(time.RFC3339)

		// Check if the file already has a modification comment
		if strings.Contains(string(newContent), "# Modified by systemctl edit") {
			// Update the existing modification comment with a new timestamp
			updatedContent := updateModificationTimestamp(string(newContent), timestamp)

			// Write the updated content back to the file
			if err := os.WriteFile(scriptPath, []byte(updatedContent), 0755); err != nil {
				return fmt.Errorf("failed to update modification timestamp: %w", err)
			}

			fmt.Printf("Service %s has been modified and saved (timestamp updated)\n", serviceName)
		} else {
			// Add new modification comment
			modComment := fmt.Sprintf("\n# Modified by systemctl edit on %s\n", timestamp)

			// Append the comment to the file
			f, err := os.OpenFile(scriptPath, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("failed to open file for appending: %w", err)
			}
			defer f.Close()

			if _, err := f.WriteString(modComment); err != nil {
				return fmt.Errorf("failed to append modification comment: %w", err)
			}

			fmt.Printf("Service %s has been modified and saved\n", serviceName)
		}
	} else {
		fmt.Printf("Service %s was not modified\n", serviceName)
	}

	return nil
}

// findEditor looks for available editors in the system
func findEditor() string {
	// Check for environment variable
	if editor := os.Getenv("EDITOR"); editor != "" {
		if _, err := exec.LookPath(editor); err == nil {
			return editor
		}
	}

	// Check for common editors in order of preference
	editors := []string{"vi", "nano", "ed"}
	for _, editor := range editors {
		if path, err := exec.LookPath(editor); err == nil {
			return path
		}
	}

	return ""
}

// updateModificationTimestamp updates the timestamp in the modification comment
func updateModificationTimestamp(content, timestamp string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, "# Modified by systemctl edit on ") {
			lines[i] = "# Modified by systemctl edit on " + timestamp
		}
	}
	return strings.Join(lines, "\n")
}

func init() {
	rootCmd.AddCommand(editCmd)
}
