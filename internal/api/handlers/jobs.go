package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/samik-k21/research-compute-queue/internal/database"
	"github.com/samik-k21/research-compute-queue/internal/models"
)

type JobHandler struct {
	db *database.DB
}

func NewJobHandler(db *database.DB) *JobHandler {
	return &JobHandler{db: db}
}

// SubmitJob creates a new job
func (h *JobHandler) SubmitJob(c *gin.Context) {
	var req models.CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get user ID from JWT token
	// For now, hardcode user ID 1
	userID := 1

	// Get user's group ID
	var groupID int
	err := h.db.QueryRow("SELECT group_id FROM users WHERE id=$1", userID).Scan(&groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}

	// Insert job
	var jobID int
	err = h.db.QueryRow(`
		INSERT INTO jobs (user_id, group_id, script, cpu_cores, memory_gb, gpu_count, 
		                  estimated_hours, priority, status, submitted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`, userID, groupID, req.Script, req.CPUCores, req.MemoryGB, req.GPUCount,
		req.EstimatedHours, req.Priority, models.StatusPending, time.Now()).Scan(&jobID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job"})
		return
	}

	// TODO: Handle job dependencies if provided

	c.JSON(http.StatusCreated, gin.H{
		"message": "Job submitted successfully",
		"job_id":  jobID,
		"status":  models.StatusPending,
	})
}

// GetJob retrieves a job by ID
func (h *JobHandler) GetJob(c *gin.Context) {
	jobIDStr := c.Param("id")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	var job models.Job
	err = h.db.QueryRow(`
		SELECT id, user_id, group_id, script, cpu_cores, memory_gb, gpu_count,
		       estimated_hours, status, priority, submitted_at, started_at, 
		       completed_at, exit_code, output_path, error_message, worker_id
		FROM jobs WHERE id=$1
	`, jobID).Scan(
		&job.ID, &job.UserID, &job.GroupID, &job.Script, &job.CPUCores,
		&job.MemoryGB, &job.GPUCount, &job.EstimatedHours, &job.Status,
		&job.Priority, &job.SubmittedAt, &job.StartedAt, &job.CompletedAt,
		&job.ExitCode, &job.OutputPath, &job.ErrorMessage, &job.WorkerID,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// ListJobs retrieves all jobs for a user
func (h *JobHandler) ListJobs(c *gin.Context) {
	// TODO: Get user ID from JWT
	userID := 1

	// Optional filters
	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)

	// Build query
	query := `
		SELECT id, user_id, group_id, script, cpu_cores, memory_gb, gpu_count,
		       status, priority, submitted_at, started_at, completed_at
		FROM jobs WHERE user_id=$1
	`
	args := []interface{}{userID}

	if status != "" {
		query += " AND status=$2"
		args = append(args, status)
	}

	query += " ORDER BY submitted_at DESC LIMIT $" + strconv.Itoa(len(args)+1)
	args = append(args, limit)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	jobs := []models.Job{}
	for rows.Next() {
		var job models.Job
		err := rows.Scan(
			&job.ID, &job.UserID, &job.GroupID, &job.Script, &job.CPUCores,
			&job.MemoryGB, &job.GPUCount, &job.Status, &job.Priority,
			&job.SubmittedAt, &job.StartedAt, &job.CompletedAt,
		)
		if err != nil {
			continue
		}
		jobs = append(jobs, job)
	}

	c.JSON(http.StatusOK, gin.H{
		"jobs":  jobs,
		"count": len(jobs),
	})
}

// CancelJob cancels a pending or running job
func (h *JobHandler) CancelJob(c *gin.Context) {
	jobIDStr := c.Param("id")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	// TODO: Verify user owns this job

	// Update job status
	result, err := h.db.Exec(`
		UPDATE jobs 
		SET status=$1, completed_at=$2 
		WHERE id=$3 AND status IN ($4, $5)
	`, models.StatusCancelled, time.Now(), jobID, models.StatusPending, models.StatusRunning)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel job"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found or already completed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Job cancelled successfully",
		"job_id":  jobID,
	})
}