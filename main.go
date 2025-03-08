package main

import (
	"os"

	"systemctl-alpine/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
