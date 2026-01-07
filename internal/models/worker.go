package models

import "time"

// Worker represents a compute node
type Worker struct {
	ID            int       `json:"id"`
	Hostname      string    `json:"hostname"`
	CPUCores      int       `json:"cpu_cores"`
	MemoryGB      int       `json:"memory_gb"`
	GPUCount      int       `json:"gpu_count"`
	Status        string    `json:"status"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	CreatedAt     time.Time `json:"created_at"`
}

// Worker status constants
const (
	WorkerIdle    = "idle"
	WorkerBusy    = "busy"
	WorkerOffline = "offline"
)