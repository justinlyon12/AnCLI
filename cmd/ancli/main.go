package main

import (
	"os"
)

func main() {
	os.Exit(run(&DefaultConfigLoader{}))
}

// run is the main entry point that returns an exit code
// This function is testable and can be called with different config loaders
func run(loader ConfigLoader) int {
	// Create and execute root command
	cmd := NewRootCmd(loader)
	if err := cmd.Execute(); err != nil {
		return 1
	}

	return 0
}
