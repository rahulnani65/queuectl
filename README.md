# queuectl

A lightweight, production-ready CLI tool for managing background job queues with automatic retries, exponential backoff, and Dead Letter Queue (DLQ) support.

## 🎯 What is queuectl?

`queuectl` is a command-line job queue system that lets you run shell commands as background jobs. It handles failures gracefully by automatically retrying failed jobs with exponential backoff, and moves permanently failed jobs to a Dead Letter Queue for manual inspection.

Think of it as a simple task runner that:
- Runs your commands in the background
- Retries failed jobs automatically
- Keeps track of everything in a database
- Lets you monitor and manage jobs through a CLI

## 🚀 Quick Start

### Prerequisites

- Go 1.24.5 or higher
- SQLite3 (usually pre-installed on macOS/Linux)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/rahulnani65/queuectl.git
cd queuectl
```

2. Build the application:
```bash
go build -o queuectl .
```

3. Verify installation:
```bash
./queuectl --help
```

## 📖 Usage Examples

### Enqueue a Job

Add a simple job to the queue:
```bash
./queuectl enqueue 'echo "Hello World"'
```

Enqueue with custom retry limit (JSON format):
```bash
./queuectl enqueue '{"command":"curl -s https://api.example.com","max_retries":5}'
```

### Start Workers

Start 2 workers to process jobs:
```bash
./queuectl worker start --count 2
```

Workers will continue running until you press `Ctrl+C`. They'll process jobs automatically.

### Check Queue Status

See a summary of all jobs:
```bash
./queuectl status
```

Example output:
```
════════════════════════════════════
         Queue Status Report        
════════════════════════════════════
  Pending:         3 jobs
  Processing:      1 jobs
  Completed:       5 jobs
  Failed:          0 jobs
  Dead (DLQ):      0 jobs
────────────────────────────────────
  Active Workers: 2
════════════════════════════════════
```

### List Jobs by State

View pending jobs:
```bash
./queuectl list --state pending
```

View completed jobs:
```bash
./queuectl list --state completed
```

Available states: `pending`, `processing`, `completed`, `failed`, `dead`

### Manage Dead Letter Queue

List all dead jobs:
```bash
./queuectl dlq list
```

Retry a dead job (resets attempts and re-enqueues):
```bash
./queuectl dlq retry <job-id>
```

### Configuration

Get a configuration value:
```bash
./queuectl config get max-retries
```

Set a configuration value:
```bash
./queuectl config set max-retries 5
./queuectl config set backoff-base 2
./queuectl config set job-timeout 300
```

Available config keys:
- `max-retries`: Maximum retry attempts before moving to DLQ (default: 3)
- `backoff-base`: Base for exponential backoff calculation (default: 2)
- `job-timeout`: Command execution timeout in seconds (default: 300)

## 🏗️ Architecture Overview

### Job Lifecycle

Jobs progress through the following states:

1. **PENDING** → Job is queued and waiting for a worker
2. **PROCESSING** → A worker is currently executing the job
3. **COMPLETED** → Job finished successfully
4. **FAILED** → Job failed but will retry (if attempts < max_retries)
5. **DEAD** → Job permanently failed and moved to DLQ

### How It Works

1. **Enqueue**: Jobs are added to the database in `PENDING` state
2. **Worker Acquisition**: Workers poll the database for pending jobs, atomically locking them
3. **Execution**: Workers execute commands via shell with timeout protection
4. **Retry Logic**: Failed jobs are rescheduled with exponential backoff (delay = base^attempts seconds)
5. **DLQ**: Jobs that exceed max_retries are moved to `DEAD` state

### Data Persistence

All job data is stored in SQLite database (`data/queuectl.db`). This ensures:
- Jobs persist across restarts
- No data loss if the application crashes
- Easy inspection and debugging

### Worker Management

- Multiple workers can run concurrently
- Jobs are atomically acquired (no duplicate processing)
- Graceful shutdown: workers finish current jobs before exiting
- Workers poll for jobs every second when idle

### Retry & Backoff

When a job fails:
1. Attempt count is incremented
2. If attempts < max_retries: Job is rescheduled with exponential backoff
3. If attempts >= max_retries: Job is moved to DLQ

Backoff formula: `delay = backoff_base ^ attempts` seconds

Example with base=2:
- 1st retry: 2 seconds
- 2nd retry: 4 seconds
- 3rd retry: 8 seconds

## 📁 Project Structure

```
queuectl/
├── cmd/              # CLI commands
│   ├── root.go       # Main CLI setup
│   ├── enqueue.go    # Enqueue command
│   ├── worker.go     # Worker management
│   ├── list.go       # List jobs
│   ├── status.go     # Status summary
│   ├── dlq.go        # Dead Letter Queue
│   └── config.go     # Configuration
├── pkg/              # Core logic
│   ├── models.go     # Job data structures
│   ├── database.go   # Database operations
│   ├── executor.go   # Command execution
│   └── worker.go     # Worker pool logic
├── data/             # Database storage
│   └── queuectl.db   # SQLite database
├── main.go           # Entry point
├── test.sh           # Integration test script
└── README.md         # This file
```

## 🧪 Testing

### Running Integration Tests

A test script is included to verify core functionality:

```bash
chmod +x test.sh
./test.sh
```

The test script validates:
- Job enqueueing
- Worker processing
- Status reporting
- Job listing
- Configuration management

### Manual Testing

1. **Test successful job:**
```bash
./queuectl enqueue 'echo "Success"'
./queuectl worker start --count 1
# Press Ctrl+C after job completes
./queuectl list --state completed
```

2. **Test retry mechanism:**
```bash
./queuectl enqueue 'echo "Test" && exit 1'
./queuectl config set max-retries 2
./queuectl worker start --count 1
# Watch the job retry with backoff
```

3. **Test DLQ:**
```bash
./queuectl enqueue 'false'  # This will always fail
./queuectl config set max-retries 1
./queuectl worker start --count 1
# Job should move to DLQ after 1 attempt
./queuectl dlq list
```

## 💡 Assumptions & Trade-offs

### Assumptions

1. **Single Machine**: Designed for single-machine deployment. For distributed systems, consider using Redis or RabbitMQ.
2. **Shell Commands**: Jobs are executed as shell commands. Ensure commands are available in the system PATH.
3. **SQLite**: Using SQLite for simplicity. For high-throughput scenarios, consider PostgreSQL or MySQL.
4. **File-based Storage**: Database is file-based, so ensure proper file permissions and disk space.

### Trade-offs

1. **Polling vs Events**: Workers poll the database every second. For lower latency, consider event-driven architecture.
2. **No Job Priorities**: Jobs are processed FIFO. Priority queues could be added as an enhancement.
3. **Simple Backoff**: Exponential backoff is straightforward. More sophisticated algorithms (e.g., jitter) could be added.
4. **Limited Observability**: Basic logging is included. For production, consider structured logging and metrics.

### Design Decisions

1. **SQLite over JSON**: Chosen for ACID guarantees and better concurrent access.
2. **Atomic Job Acquisition**: Using database transactions to prevent race conditions.
3. **Graceful Shutdown**: Workers finish current jobs before exiting to avoid data loss.
4. **CLI-first**: Prioritized CLI usability over web interfaces (can be added later).

## 🎓 Features Implemented

✅ Job enqueueing with custom retry limits  
✅ Multiple worker support  
✅ Automatic retry with exponential backoff  
✅ Dead Letter Queue (DLQ)  
✅ Persistent job storage (SQLite)  
✅ Configuration management  
✅ Job timeout handling  
✅ Graceful worker shutdown  
✅ Atomic job acquisition (no duplicates)  
✅ Job output capture  
✅ Exit code tracking  
✅ Scheduled job support (via `scheduled_at`)  

## 🔮 Future Enhancements

Some ideas for future improvements:

- Job priority queues
- Web dashboard for monitoring
- Job scheduling (cron-like syntax)
- Job dependencies
- Metrics export
- Distributed worker support
- Job output streaming


## 🙏 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📧 Contact

rahulview65@gmail.com

---

**Built with ❤️ using Go**

