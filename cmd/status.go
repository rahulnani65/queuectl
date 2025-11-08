package cmd

import (
	"fmt"
	"queuectl/pkg"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show queue status",
	RunE: func(cmd *cobra.Command, args []string) error {
		summary, err := db.GetStatusSummary()
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}
		fmt.Println("\n════════════════════════════════════")
		fmt.Println("         Queue Status Report        ")
		fmt.Println("════════════════════════════════════")
		fmt.Printf("  Pending:      %4d jobs\n", summary[pkg.StatePending])
		fmt.Printf("  Processing:   %4d jobs\n", summary[pkg.StateProcessing])
		fmt.Printf("  Completed:    %4d jobs\n", summary[pkg.StateCompleted])
		fmt.Printf("  Failed:       %4d jobs\n", summary[pkg.StateFailed])
		fmt.Printf("  Dead (DLQ):   %4d jobs\n", summary[pkg.StateDead])
		fmt.Println("────────────────────────────────────")
		fmt.Printf("  Active Workers: %d\n", wm.GetActiveWorkerCount())
		fmt.Println("════════════════════════════════════\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
