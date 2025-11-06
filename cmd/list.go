package cmd

import (
	"fmt"
	"queuectl/pkg"
	"strings"

	"github.com/spf13/cobra"
)

// listCmd prints jobs filtered by state in a compact table format.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List jobs by state",
	RunE: func(cmd *cobra.Command, args []string) error {
		stateStr, _ := cmd.Flags().GetString("state")
		state := pkg.JobState(strings.ToUpper(stateStr))
		jobs, err := db.FindJobsByState(state)
		if err != nil {
			return fmt.Errorf("failed to list jobs: %w", err)
		}

		fmt.Println("\n" + strings.Repeat("═", 110))
		fmt.Printf("%-36s | %-40s | %-12s | %-8s\n", "ID", "Command", "State", "Attempts")
		fmt.Println(strings.Repeat("─", 110))

		for _, job := range jobs {
			cmdDisplay := job.Command
			if len(cmdDisplay) > 40 {
				cmdDisplay = cmdDisplay[:37] + "..."
			}
			fmt.Printf("%-36s | %-40s | %-12s | %-8d\n", job.ID, cmdDisplay, job.State, job.Attempts)
		}
		fmt.Println(strings.Repeat("═", 110))
		fmt.Printf("Total: %d jobs\n\n", len(jobs))

		return nil
	},
}

func init() {
	listCmd.Flags().StringP("state", "s", "pending", "Job state to filter")
	rootCmd.AddCommand(listCmd)
}
