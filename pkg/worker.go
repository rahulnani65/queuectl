package pkg

import (
	"log"
	"strconv"
	"sync"
	"time"
)

// WorkerManager coordinates a pool of workers that pull jobs from the
// database and execute them, handling retries and shutdown.
type WorkerManager struct {
	db            *DB
	activeWorkers int
	stopChan      chan bool
	wg            sync.WaitGroup
}

// NewWorkerManager returns a WorkerManager bound to the provided DB.
func NewWorkerManager(db *DB) *WorkerManager {
	return &WorkerManager{
		db:       db,
		stopChan: make(chan bool),
	}
}

// StartWorkers launches count worker goroutines and tracks them until StopWorkers is called.
func (wm *WorkerManager) StartWorkers(count int) {
	wm.activeWorkers = count
	for i := 0; i < count; i++ {
		wm.wg.Add(1)
		go wm.workerLoop(i)
	}
	log.Printf("✓ Started %d workers\n", count)
}

func (wm *WorkerManager) workerLoop(id int) {
	defer wm.wg.Done()

	for {
		select {
		case <-wm.stopChan:
			log.Printf("Worker %d stopped\n", id)
			return
		default:
			job, err := wm.db.AcquireNextPendingJob()
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			if job == nil {
				time.Sleep(1 * time.Second)
				continue
			}

			wm.processJob(job, id)
		}
	}
}

// processJob runs a single job respecting the configured timeout and
// persists its outcome. On failure, it may schedule a retry.
func (wm *WorkerManager) processJob(job *Job, workerID int) {
	log.Printf("Worker %d processing: %s\n", workerID, job.Command)

	timeoutStr, _ := wm.db.GetConfig("job-timeout")
	timeout := int64(300)
	if val, err := strconv.ParseInt(timeoutStr, 10, 64); err == nil {
		timeout = val
	}

	result := ExecuteCommand(job.Command, timeout)

	if result.Success {
		job.State = StateCompleted
		job.ExitCode = &result.ExitCode
		job.Output = result.Output
		log.Printf("✓ Job %s completed\n", job.ID)
	} else {
		wm.handleFailure(job, result)
	}

	job.UpdatedAt = time.Now()
	wm.db.SaveJob(job)
}

// handleFailure updates the job attempt count and either moves the job
// to the DLQ (DEAD) or schedules a retry with exponential backoff.
func (wm *WorkerManager) handleFailure(job *Job, result ExecutionResult) {
	job.Attempts++
	job.ErrorMessage = result.Error
	job.ExitCode = &result.ExitCode

	maxRetriesStr, _ := wm.db.GetConfig("max-retries")
	maxRetries := 3
	if val, err := strconv.Atoi(maxRetriesStr); err == nil {
		maxRetries = val
	}

	if job.Attempts >= maxRetries {
		job.State = StateDead
		log.Printf("✗ Job %s → DLQ (failed %d times)\n", job.ID, job.Attempts)
	} else {
		backoffBaseStr, _ := wm.db.GetConfig("backoff-base")
		backoffBase := 2
		if val, err := strconv.Atoi(backoffBaseStr); err == nil {
			backoffBase = val
		}

		delay := job.CalculateBackoff(backoffBase)
		scheduledAt := time.Now().Add(time.Duration(delay) * time.Second)
		job.ScheduledAt = &scheduledAt
		job.State = StatePending

		log.Printf("⟳ Job %s retry in %d seconds\n", job.ID, delay)
	}
}

// StopWorkers signals all active workers to stop and waits for them to exit.
func (wm *WorkerManager) StopWorkers() {
	log.Println("Stopping workers...")
	for i := 0; i < wm.activeWorkers; i++ {
		wm.stopChan <- true
	}
	wm.wg.Wait()
	log.Println("✓ All workers stopped")
}

// GetActiveWorkerCount returns the number of worker goroutines currently running.
func (wm *WorkerManager) GetActiveWorkerCount() int {
	return wm.activeWorkers
}
