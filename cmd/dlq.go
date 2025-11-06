package cmd

import (
	"fmt"
	"queuectl/pkg"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// dlqCmd groups Dead Letter Queue operations.
var dlqCmd = &cobra.Command{
	Use:   "dlq",
	Short: "Manage Dead Letter Queue",
	Long:  "List and retry permanently failed jobs",
}

// dlqListCmd lists jobs that are currently in the DLQ.
var dlqListCmd = &cobra.Command{
	Use:   "list",
	Short: "List dead jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		jobs, err := db.FindJobsByState(pkg.StateDead)
		if err != nil {
			return fmt.Errorf("failed to list DLQ: %w", err)
		}

		fmt.Println("\n" + strings.Repeat("═", 120))
		fmt.Println("Dead Letter Queue - Failed Jobs")
		fmt.Println(strings.Repeat("═", 120))
		fmt.Printf("%-36s | %-35s | %-8s | %-10s\n", "ID", "Command", "Attempts", "Error")
		fmt.Println(strings.Repeat("─", 120))

		for _, job := range jobs {
			cmdDisplay := job.Command
			if len(cmdDisplay) > 35 {
				cmdDisplay = cmdDisplay[:32] + "..."
			}

			errDisplay := job.ErrorMessage
			if len(errDisplay) > 10 {
				errDisplay = errDisplay[:7] + "..."
			}

			fmt.Printf("%-36s | %-35s | %-8d | %-10s\n", job.ID, cmdDisplay, job.Attempts, errDisplay)
		}

		fmt.Println(strings.Repeat("═", 120))
		fmt.Printf("Total dead jobs: %d\n\n", len(jobs))

		return nil
	},
}

// dlqRetryCmd resets a DLQ job and re-enqueues it for processing.
var dlqRetryCmd = &cobra.Command{
	Use:   "retry [job-id]",
	Short: "Retry a dead job",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jobID := args[0]
		job, err := db.FindJobByID(jobID)
		if err != nil {
			return fmt.Errorf("failed to find job: %w", err)
		}

		if job == nil {
			return fmt.Errorf("job not found: %s", jobID)
		}

		if job.State != pkg.StateDead {
			return fmt.Errorf("job is not in DLQ. Current state: %s", job.State)
		}

		job.State = pkg.StatePending
		job.Attempts = 0
		job.ErrorMessage = ""
		job.ExitCode = nil
		job.UpdatedAt = time.Now()
		now := time.Now()
		job.ScheduledAt = &now

		if err := db.SaveJob(job); err != nil {
			return fmt.Errorf("failed to requeue job: %w", err)
		}

		fmt.Printf("✓ Job requeued: %s\n", jobID)
		return nil
	},
}

func init() {
	dlqCmd.AddCommand(dlqListCmd)
	dlqCmd.AddCommand(dlqRetryCmd)
	rootCmd.AddCommand(dlqCmd)
}
