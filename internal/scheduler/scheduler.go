package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/samik-k21/research-compute-queue/internal/database"
)

// Scheduler manages job scheduling and execution
type Scheduler struct {
	db              *database.DB
	interval        time.Duration
	maxConcurrent   int
	priorityCalc    *PriorityCalculator
	resourceMatcher *ResourceMatcher
	executor        *Executor
	ctx             context.Context
	cancel          context.CancelFunc
}

// JobWithPriority holds job info plus calculated priority
type JobWithPriority struct {
	ID                 int
	UserID             int
	GroupID            int
	Script             string
	CPUCores           int
	MemoryGB           int
	GPUCount           int
	Priority           int
	SubmittedAt        time.Time
	EstimatedHours     float64
	GroupPriority      int
	CalculatedPriority float64
}

// Worker holds worker information
type Worker struct {
	ID       int
	Hostname string
	CPUCores int
	MemoryGB int
	GPUCount int
	Status   string
}

// NewScheduler creates a new scheduler instance
func NewScheduler(db *database.DB, intervalSeconds int, maxConcurrent int) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		db:              db,
		interval:        time.Duration(intervalSeconds) * time.Second,
		maxConcurrent:   maxConcurrent,
		priorityCalc:    NewPriorityCalculator(db),
		resourceMatcher: NewResourceMatcher(db),
		executor:        NewExecutor(db),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Start begins the scheduling loop
func (s *Scheduler) Start() {
	log.Println("Scheduler starting...")

	// Run immediately on start
	s.runSchedulingCycle()

	// Then run on interval
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.runSchedulingCycle()
		case <-s.ctx.Done():
			log.Println("Scheduler stopping...")
			return
		}
	}
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() {
	log.Println("Stopping scheduler...")
	s.cancel()
}

// runSchedulingCycle executes one scheduling cycle
func (s *Scheduler) runSchedulingCycle() {
	log.Println("===== Running scheduling cycle =====")

	// 1. Get pending jobs
	pendingJobs, err := s.getPendingJobs()
	if err != nil {
		log.Printf("Error getting pending jobs: %v", err)
		return
	}

	if len(pendingJobs) == 0 {
		log.Println("No pending jobs")
		return
	}

	log.Printf("Found %d pending jobs", len(pendingJobs))

	// 2. Calculate priorities for all jobs
	jobsWithPriority, err := s.priorityCalc.CalculatePriorities(pendingJobs)
	if err != nil {
		log.Printf("Error calculating priorities: %v", err)
		return
	}

	// 3. Get available workers
	workers, err := s.getAvailableWorkers()
	if err != nil {
		log.Printf("Error getting workers: %v", err)
		return
	}

	if len(workers) == 0 {
		log.Println("No available workers")
		return
	}

	log.Printf("Found %d available workers", len(workers))

	// 4. Check how many jobs are currently running
	runningCount, err := s.getRunningJobCount()
	if err != nil {
		log.Printf("Error getting running job count: %v", err)
		return
	}

	slotsAvailable := s.maxConcurrent - runningCount
	if slotsAvailable <= 0 {
		log.Printf("Max concurrent jobs reached (%d/%d)", runningCount, s.maxConcurrent)
		return
	}

	log.Printf("Can schedule up to %d jobs (%d running, %d max)", slotsAvailable, runningCount, s.maxConcurrent)

	// 5. Try to schedule jobs
	scheduled := 0
	for _, job := range jobsWithPriority {
		if scheduled >= slotsAvailable {
			break
		}

		// Find a suitable worker
		worker, err := s.resourceMatcher.FindWorkerForJob(&job, workers)
		if err != nil || worker == nil {
			log.Printf("No suitable worker for job %d (needs %d CPU, %d GB RAM, %d GPU)",
				job.ID, job.CPUCores, job.MemoryGB, job.GPUCount)
			continue
		}

		// Assign and start job
		err = s.executor.StartJob(&job, worker)
		if err != nil {
			log.Printf("Error starting job %d: %v", job.ID, err)
			continue
		}

		log.Printf("✓ Scheduled job %d on worker %s (priority: %.2f)",
			job.ID, worker.Hostname, job.CalculatedPriority)
		scheduled++

		// Mark worker as busy (remove from available list)
		s.markWorkerBusy(worker.ID)
		workers = removeWorker(workers, worker.ID)
	}

	log.Printf("Scheduled %d jobs in this cycle", scheduled)
	log.Println("====================================")
}

// getPendingJobs retrieves all pending jobs from database
func (s *Scheduler) getPendingJobs() ([]JobWithPriority, error) {
	rows, err := s.db.Query(`
		SELECT j.id, j.user_id, j.group_id, j.script, j.cpu_cores, j.memory_gb,
		       j.gpu_count, j.priority, j.submitted_at, 
		       COALESCE(j.estimated_hours, 0) as estimated_hours,
		       g.priority as group_priority
		FROM jobs j
		JOIN groups g ON j.group_id = g.id
		WHERE j.status = 'pending'
		ORDER BY j.submitted_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []JobWithPriority
	for rows.Next() {
		var job JobWithPriority
		err := rows.Scan(
			&job.ID, &job.UserID, &job.GroupID, &job.Script, &job.CPUCores,
			&job.MemoryGB, &job.GPUCount, &job.Priority, &job.SubmittedAt,
			&job.EstimatedHours, &job.GroupPriority,
		)
		if err != nil {
			log.Printf("Error scanning job: %v", err)
			continue
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// getAvailableWorkers retrieves idle workers
func (s *Scheduler) getAvailableWorkers() ([]Worker, error) {
	rows, err := s.db.Query(`
		SELECT id, hostname, cpu_cores, memory_gb, gpu_count, status
		FROM workers
		WHERE status = 'idle'
		ORDER BY cpu_cores DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workers []Worker
	for rows.Next() {
		var w Worker
		err := rows.Scan(&w.ID, &w.Hostname, &w.CPUCores, &w.MemoryGB, &w.GPUCount, &w.Status)
		if err != nil {
			log.Printf("Error scanning worker: %v", err)
			continue
		}
		workers = append(workers, w)
	}

	return workers, nil
}

// getRunningJobCount returns the number of currently running jobs
func (s *Scheduler) getRunningJobCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM jobs WHERE status = 'running'").Scan(&count)
	return count, err
}

// markWorkerBusy updates worker status to busy
func (s *Scheduler) markWorkerBusy(workerID int) error {
	_, err := s.db.Exec("UPDATE workers SET status = 'busy' WHERE id = $1", workerID)
	return err
}

// removeWorker removes a worker from the list
func removeWorker(workers []Worker, workerID int) []Worker {
	for i, w := range workers {
		if w.ID == workerID {
			return append(workers[:i], workers[i+1:]...)
		}
	}
	return workers
}