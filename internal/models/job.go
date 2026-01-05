package models

import "time"

// Job represents a compute job
type Job struct {
	ID             int        `json:"id"`
	UserID         int        `json:"user_id"`
	GroupID        int        `json:"group_id"`
	Script         string     `json:"script"`
	CPUCores       int        `json:"cpu_cores"`
	MemoryGB       int        `json:"memory_gb"`
	GPUCount       int        `json:"gpu_count"`
	EstimatedHours float64    `json:"estimated_hours,omitempty"`
	Status         string     `json:"status"`
	Priority       int        `json:"priority"`
	SubmittedAt    time.Time  `json:"submitted_at"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	ExitCode       *int       `json:"exit_code,omitempty"`
	OutputPath     string     `json:"output_path,omitempty"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	WorkerID       *int       `json:"worker_id,omitempty"`
}

// JobStatus constants
const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"
)

// CreateJobRequest represents a job submission request
type CreateJobRequest struct {
	Script         string   `json:"script" binding:"required"`
	CPUCores       int      `json:"cpu_cores" binding:"required,min=1"`
	MemoryGB       int      `json:"memory_gb" binding:"required,min=1"`
	GPUCount       int      `json:"gpu_count"`
	EstimatedHours float64  `json:"estimated_hours"`
	Priority       int      `json:"priority" binding:"min=1,max=10"`
	Dependencies   []string `json:"dependencies"` // Job IDs this job depends on
}