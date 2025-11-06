package cmd

import (
	"encoding/json"
	"fmt"
	"queuectl/pkg"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// enqueueCmd enqueues a new job from a raw command or a JSON payload.
var enqueueCmd = &cobra.Command{
	Use:   "enqueue [command]",
	Short: "Enqueue a new job",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		command := args[0]

		var jobReq struct {
			ID         string `json:"id"`
			Command    string `json:"command"`
			MaxRetries int    `json:"max_retries"`
		}

		err := json.Unmarshal([]byte(command), &jobReq)
		if err != nil {
			jobReq.Command = command
			jobReq.MaxRetries = 3
		}

		if jobReq.ID == "" {
			jobReq.ID = uuid.New().String()
		}

		job := &pkg.Job{
			ID:         jobReq.ID,
			Command:    jobReq.Command,
			State:      pkg.StatePending,
			Attempts:   0,
			MaxRetries: jobReq.MaxRetries,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if err := db.SaveJob(job); err != nil {
			return fmt.Errorf("failed to enqueue job: %w", err)
		}

		fmt.Println("✓ Job enqueued successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(enqueueCmd)
}
