package main

import (
	"fmt"
	"os"
	"queuectl/cmd"
)

// main is the entry point for the queuectl CLI.
func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
