package pkg

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func NewDB(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(); err != nil {
		return nil, err
	}
	if err := initSchema(conn); err != nil {
		return nil, err
	}
	if err := recoverStuckJobs(conn); err != nil {
		return nil, err
	}
	log.Println("✓ Database initialized")
	return &DB{conn: conn}, nil
}

func initSchema(conn *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS jobs (
		id TEXT PRIMARY KEY,
		command TEXT NOT NULL,
		state TEXT NOT NULL DEFAULT 'PENDING',
		attempts INTEGER DEFAULT 0,
		max_retries INTEGER DEFAULT 3,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		scheduled_at TIMESTAMP,
		error_message TEXT,
		exit_code INTEGER,
		output TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_state_scheduled 
	ON jobs(state, scheduled_at);

	CREATE TABLE IF NOT EXISTS config (
		key TEXT PRIMARY KEY,
		value TEXT
	);
	`

	_, err := conn.Exec(schema)
	if err != nil {
		return err
	}
	_, _ = conn.Exec(`
		INSERT OR IGNORE INTO config (key, value) 
		VALUES 
			('max-retries', '3'),
			('backoff-base', '2'),
			('job-timeout', '300')
	`)

	return nil
}

// Recover jobs stuck in PROCESSING state after crash
func recoverStuckJobs(conn *sql.DB) error {
	result, err := conn.Exec(
		"UPDATE jobs SET state = ? WHERE state = ?",
		StatePending, StateProcessing,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected > 0 {
		log.Printf("✓ Recovered %d stuck job(s) from PROCESSING state\n", rowsAffected)
	}
	return nil
}

func (db *DB) SaveJob(job *Job) error {
	_, err := db.conn.Exec(`
		INSERT OR REPLACE INTO jobs 
		(id, command, state, attempts, max_retries, 
		 created_at, updated_at, scheduled_at, 
		 error_message, exit_code, output)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		job.ID, job.Command, job.State, job.Attempts, 
		job.MaxRetries, job.CreatedAt, job.UpdatedAt, 
		job.ScheduledAt, job.ErrorMessage, 
		job.ExitCode, job.Output,
	)
	return err
}
func (db *DB) FindJobsByState(state JobState) ([]Job, error) {
	rows, err := db.conn.Query(
		`SELECT id, command, state, attempts, max_retries, created_at, updated_at, scheduled_at, error_message, exit_code, output 
		 FROM jobs WHERE state = ? LIMIT 100`, 
		state,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	jobs := []Job{}
	for rows.Next() {
		job := Job{}
		err := rows.Scan(
			&job.ID, &job.Command, &job.State, 
			&job.Attempts, &job.MaxRetries,
			&job.CreatedAt, &job.UpdatedAt, 
			&job.ScheduledAt, &job.ErrorMessage, 
			&job.ExitCode, &job.Output,
		)
		if err != nil {
			continue
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (db *DB) FindJobByID(id string) (*Job, error) {
	job := &Job{}
	err := db.conn.QueryRow(
		`SELECT id, command, state, attempts, max_retries, created_at, updated_at, scheduled_at, error_message, exit_code, output FROM jobs WHERE id = ?`, id).
		Scan(&job.ID, &job.Command, &job.State, &job.Attempts, &job.MaxRetries,
			&job.CreatedAt, &job.UpdatedAt, &job.ScheduledAt, &job.ErrorMessage,
			&job.ExitCode, &job.Output)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return job, err
}

func (db *DB) GetStatusSummary() (map[JobState]int, error) {
	summary := make(map[JobState]int)
	summary[StatePending] = 0
	summary[StateProcessing] = 0
	summary[StateCompleted] = 0
	summary[StateFailed] = 0
	summary[StateDead] = 0

	rows, err := db.conn.Query("SELECT state, COUNT(*) as count FROM jobs GROUP BY state")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var state JobState
		var count int
		err := rows.Scan(&state, &count)
		if err != nil {
			continue
		}
		summary[state] = count
	}
	return summary, nil
}

// AcquireNextPendingJob atomically grabs the next pending job and marks it as processing
func (db *DB) AcquireNextPendingJob() (*Job, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	job := &Job{}
	err = tx.QueryRow(`
		SELECT id, command, state, attempts, max_retries, created_at, updated_at, scheduled_at, error_message, exit_code, output
		FROM jobs
		WHERE state = ? AND (scheduled_at IS NULL OR scheduled_at <= datetime('now'))
		ORDER BY created_at
		LIMIT 1
	`, StatePending).Scan(
		&job.ID, &job.Command, &job.State, &job.Attempts, &job.MaxRetries,
		&job.CreatedAt, &job.UpdatedAt, &job.ScheduledAt, &job.ErrorMessage,
		&job.ExitCode, &job.Output,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec("UPDATE jobs SET state = ? WHERE id = ?", StateProcessing, job.ID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	job.State = StateProcessing
	return job, nil
}

func (db *DB) GetConfig(key string) (string, error) {
	var value string
	err := db.conn.QueryRow(
		"SELECT value FROM config WHERE key = ?",
		key,
	).Scan(&value)
	return value, err
}

func (db *DB) SetConfig(key, value string) error {
	_, err := db.conn.Exec(
		"INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)",
		key, value,
	)
	return err
}

func (db *DB) Close() error {
	return db.conn.Close()
}
