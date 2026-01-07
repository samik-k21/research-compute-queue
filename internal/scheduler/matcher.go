package scheduler

import (
	"errors"

	"github.com/samik-k21/research-compute-queue/internal/database"
)

// ResourceMatcher finds suitable workers for jobs
type ResourceMatcher struct {
	db *database.DB
}

// NewResourceMatcher creates a new resource matcher
func NewResourceMatcher(db *database.DB) *ResourceMatcher {
	return &ResourceMatcher{db: db}
}

// FindWorkerForJob finds a worker that can run the job
func (rm *ResourceMatcher) FindWorkerForJob(job *JobWithPriority, workers []Worker) (*Worker, error) {
	for i := range workers {
		if rm.workerCanRunJob(&workers[i], job) {
			return &workers[i], nil
		}
	}
	return nil, errors.New("no suitable worker found")
}

// workerCanRunJob checks if worker has enough resources
func (rm *ResourceMatcher) workerCanRunJob(worker *Worker, job *JobWithPriority) bool {
	return worker.CPUCores >= job.CPUCores &&
		worker.MemoryGB >= job.MemoryGB &&
		worker.GPUCount >= job.GPUCount
}