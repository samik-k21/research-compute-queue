package scheduler

import (
	"log"
	"time"

	"github.com/samik-k21/research-compute-queue/internal/database"
)

// PriorityCalculator calculates job priorities using fair-share algorithm
type PriorityCalculator struct {
	db *database.DB
}

// NewPriorityCalculator creates a new priority calculator
func NewPriorityCalculator(db *database.DB) *PriorityCalculator {
	return &PriorityCalculator{db: db}
}

// CalculatePriorities computes final priority for each job
func (pc *PriorityCalculator) CalculatePriorities(jobs []JobWithPriority) ([]JobWithPriority, error) {
	// Get usage data for fair-share calculation
	groupUsage, err := pc.getGroupUsage()
	if err != nil {
		log.Printf("Error getting group usage: %v", err)
		return jobs, nil // Continue with base priorities
	}
	
	// Calculate priority for each job
	for i := range jobs {
		jobs[i].CalculatedPriority = pc.calculateJobPriority(&jobs[i], groupUsage)
	}
	
	// Sort jobs by calculated priority (highest first)
	sortJobsByPriority(jobs)
	
	return jobs, nil
}

// calculateJobPriority computes the final priority score
func (pc *PriorityCalculator) calculateJobPriority(job *JobWithPriority, groupUsage map[int]UsageData) float64 {
	// Base priority (1-10, from job and group)
	basePriority := float64(job.Priority + job.GroupPriority)
	
	// Fair-share multiplier
	fairShareMultiplier := pc.calculateFairShare(job.GroupID, groupUsage)
	
	// Wait time boost (jobs waiting longer get priority boost)
	waitTimeBoost := pc.calculateWaitTimeBoost(job.SubmittedAt)
	
	// Final priority formula
	finalPriority := basePriority * fairShareMultiplier * waitTimeBoost
	
	return finalPriority
}

// calculateFairShare computes fair-share multiplier for a group
func (pc *PriorityCalculator) calculateFairShare(groupID int, groupUsage map[int]UsageData) float64 {
	usage, exists := groupUsage[groupID]
	if !exists {
		return 1.0 // No usage data, use neutral multiplier
	}
	
	// If group hasn't used anything, boost their priority
	if usage.CPUHoursUsed == 0 {
		return 2.0
	}
	
	// Calculate fair-share ratio
	// If group uses less than quota → ratio > 1 (boost)
	// If group uses more than quota → ratio < 1 (penalty)
	quota := float64(usage.CPUQuota)
	used := usage.CPUHoursUsed
	
	if quota <= 0 {
		return 1.0
	}
	
	ratio := quota / used
	
	// Cap the multiplier between 0.5 and 2.0
	if ratio > 2.0 {
		ratio = 2.0
	}
	if ratio < 0.5 {
		ratio = 0.5
	}
	
	return ratio
}

// calculateWaitTimeBoost increases priority for jobs waiting longer
func (pc *PriorityCalculator) calculateWaitTimeBoost(submittedAt time.Time) float64 {
	waitMinutes := time.Since(submittedAt).Minutes()
	
	// 1% boost per hour of waiting
	// Job waiting 1 hour: 1.01
	// Job waiting 10 hours: 1.10
	boost := 1.0 + (waitMinutes / 60.0 * 0.01)
	
	// Cap at 1.5x (50% boost after 50 hours)
	if boost > 1.5 {
		boost = 1.5
	}
	
	return boost
}

// UsageData holds group resource usage information
type UsageData struct {
	GroupID       int
	CPUQuota      int
	CPUHoursUsed  float64
}

// getGroupUsage retrieves recent usage for all groups
func (pc *PriorityCalculator) getGroupUsage() (map[int]UsageData, error) {
	// Get usage from last 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	
	rows, err := pc.db.Query(`
		SELECT g.id, g.cpu_quota, COALESCE(SUM(ul.cpu_hours_used), 0) as total_used
		FROM groups g
		LEFT JOIN usage_logs ul ON g.id = ul.group_id AND ul.logged_at > $1
		GROUP BY g.id, g.cpu_quota
	`, thirtyDaysAgo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	usage := make(map[int]UsageData)
	for rows.Next() {
		var data UsageData
		err := rows.Scan(&data.GroupID, &data.CPUQuota, &data.CPUHoursUsed)
		if err != nil {
			continue
		}
		usage[data.GroupID] = data
	}
	
	return usage, nil
}

// sortJobsByPriority sorts jobs by calculated priority (descending)
func sortJobsByPriority(jobs []JobWithPriority) {
	// Simple bubble sort (fine for small job counts)
	n := len(jobs)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if jobs[j].CalculatedPriority < jobs[j+1].CalculatedPriority {
				jobs[j], jobs[j+1] = jobs[j+1], jobs[j]
			}
		}
	}
}