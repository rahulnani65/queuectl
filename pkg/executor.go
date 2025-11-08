package pkg

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

// ExecutionResult holds the result of command execution
type ExecutionResult struct {
	Success  bool
	ExitCode int
	Error    string
	Output   string
}

// ExecuteCommand runs a shell command with timeout
func ExecuteCommand(command string, timeout int64) ExecutionResult {
	ctx, cancel := context.WithTimeout(
		context.Background(), 
		time.Duration(timeout)*time.Second,
	)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		cmd.Process.Kill()
		return ExecutionResult{
			Success:  false,
			ExitCode: -1,
			Error:    "Timeout",
			Output:   output.String(),
		}
	}

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return ExecutionResult{
				Success:  false,
				ExitCode: -1,
				Error:    err.Error(),
				Output:   output.String(),
			}
		}
	}

	return ExecutionResult{
		Success:  exitCode == 0,
		ExitCode: exitCode,
		Error:    "",
		Output:   output.String(),
	}
}
