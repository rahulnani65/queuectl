package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

// workerCmd groups worker-related subcommands.
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Manage worker processes",
}

// workerStartCmd starts N worker goroutines until interrupted.
var workerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start worker processes",
	RunE: func(cmd *cobra.Command, args []string) error {
		count, _ := cmd.Flags().GetInt("count")

		if count <= 0 {
			return fmt.Errorf("count must be greater than 0")
		}

		wm.StartWorkers(count)
		fmt.Printf("✓ Started %d worker(s)\n", count)
		fmt.Println("Press Ctrl+C to stop workers...")

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		fmt.Println("\n✓ Shutting down workers gracefully...")
		wm.StopWorkers()
		return nil
	},
}

func init() {
	workerStartCmd.Flags().IntP("count", "c", 1, "Number of workers to start")
	workerCmd.AddCommand(workerStartCmd)
	rootCmd.AddCommand(workerCmd)
}
