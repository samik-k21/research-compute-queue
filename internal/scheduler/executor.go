package scheduler

import (
	"fmt"
	"log"
	"time"

	"github.com/samik-k21/research-compute-queue/internal/database"
)

// Executor handles job execution
type Executor struct {
	db *database.DB
}

// NewExecutor creates a new executor
func NewExecutor(db *database.DB) *Executor {
	return &Executor{db: db}
}

// StartJob assigns a job to a worker and starts it
func (e *Executor) StartJob(job *JobWithPriority, worker *Worker) error {
	now := time.Now()
	
	// Update job status to running
	_, err := e.db.Exec(`
		UPDATE jobs
		SET status = 'running', started_at = $1, worker_id = $2
		WHERE id = $3
	`, now, worker.ID, job.ID)
	
	if err != nil {
		return fmt.Errorf("failed to start job: %w", err)
	}
	
	log.Printf("Started job %d on worker %s", job.ID, worker.Hostname)
	
	// Simulate job execution in background
	go e.simulateJobExecution(job, worker)
	
	return nil
}

// simulateJobExecution simulates a job running (since we don't have real compute)
func (e *Executor) simulateJobExecution(job *JobWithPriority, worker *Worker) {
	// Simulate execution time (use estimated hours, or default to 1-5 minutes for testing)
	var duration time.Duration
	if job.EstimatedHours > 0 {
		// Use actual estimated time (in production)
		duration = time.Duration(job.EstimatedHours * float64(time.Hour))
	} else {
		// For testing: random duration between 30 seconds and 2 minutes
		duration = time.Second * time.Duration(30+job.ID%90)
	}
	
	log.Printf("Job %d will run for %v", job.ID, duration)
	
	// Wait for "execution" to complete
	time.Sleep(duration)
	
	// Mark job as completed
	e.completeJob(job.ID, worker.ID, true)
}

// completeJob marks a job as completed or failed
func (e *Executor) completeJob(jobID int, workerID int, success bool) {
	now := time.Now()
	status := "completed"
	exitCode := 0
	
	if !success {
		status = "failed"
		exitCode = 1
	}
	
	// Update job status
	_, err := e.db.Exec(`
		UPDATE jobs
		SET status = $1, completed_at = $2, exit_code = $3
		WHERE id = $4
	`, status, now, exitCode, jobID)
	
	if err != nil {
		log.Printf("Error completing job %d: %v", jobID, err)
		return
	}
	
	// Mark worker as idle again
	_, err = e.db.Exec("UPDATE workers SET status = 'idle' WHERE id = $1", workerID)
	if err != nil {
		log.Printf("Error marking worker %d as idle: %v", workerID, err)
	}
	
	// Log usage for fair-share calculation
	e.logUsage(jobID)
	
	log.Printf("Job %d completed with status: %s", jobID, status)
}

// logUsage records CPU hours used for fair-share tracking
func (e *Executor) logUsage(jobID int) {
	// Calculate CPU hours used
	var groupID int
	var cpuCores int
	var startedAt, completedAt time.Time
	
	err := e.db.QueryRow(`
		SELECT group_id, cpu_cores, started_at, completed_at
		FROM jobs
		WHERE id = $1
	`, jobID).Scan(&groupID, &cpuCores, &startedAt, &completedAt)
	
	if err != nil {
		log.Printf("Error getting job info for usage logging: %v", err)
		return
	}
	
	// Calculate CPU hours
	duration := completedAt.Sub(startedAt).Hours()
	cpuHours := duration * float64(cpuCores)
	
	// Insert usage log
	_, err = e.db.Exec(`
		INSERT INTO usage_logs (group_id, job_id, cpu_hours_used, logged_at)
		VALUES ($1, $2, $3, $4)
	`, groupID, jobID, cpuHours, time.Now())
	
	if err != nil {
		log.Printf("Error logging usage: %v", err)
		return
	}
	
	log.Printf("Logged %.2f CPU hours for group %d", cpuHours, groupID)
}