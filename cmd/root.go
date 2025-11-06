package cmd

import (
	"fmt"
	"queuectl/pkg"

	"github.com/spf13/cobra"
)

var (
	db *pkg.DB
	wm *pkg.WorkerManager
)

// rootCmd is the base command for the queuectl CLI.
var rootCmd = &cobra.Command{
	Use:   "queuectl",
	Short: "Background job queue system",
	Long:  "A production-grade job queue with retries and dead letter queue",
}

// Execute initializes dependencies and runs the CLI command tree.
func Execute() error {
	var err error

	// Initialize database
	db, err = pkg.NewDB("./data/queuectl.db")
	if err != nil {
		return fmt.Errorf("failed to init DB: %w", err)
	}
	defer db.Close()

	// Initialize workers
	wm = pkg.NewWorkerManager(db)

	return rootCmd.Execute()
}

func init() {
    // Subcommands are registered in their respective files' init functions.
}
