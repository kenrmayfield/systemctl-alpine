package main

import (
	_ "embed"
	"os"
	"strings"

	"systemctl-alpine/cmd"
)

//go:embed .version
var version string

func main() {
	cmd.Version = strings.TrimSpace(version)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
