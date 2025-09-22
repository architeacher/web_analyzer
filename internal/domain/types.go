package domain

import (
	"time"

	"github.com/architeacher/svc-web-analyzer/internal/handlers"
)

// DependencyStatus represents the health status of a dependency
type DependencyStatus struct {
	Status       handlers.DependencyCheckStatus
	ResponseTime float32
	LastChecked  time.Time
	Error        string
}

// HealthResult contains comprehensive health check results
type HealthResult struct {
	OverallStatus handlers.HealthResponseStatus
	Storage       DependencyStatus
	Cache         DependencyStatus
	Queue         DependencyStatus
	Uptime        float32
}

// ReadinessResult contains readiness check results
type ReadinessResult struct {
	OverallStatus handlers.ReadinessResponseStatus
	Storage       DependencyStatus
	Cache         DependencyStatus
	Queue         DependencyStatus
}

// LivenessResult contains liveness check results
type LivenessResult struct {
	OverallStatus handlers.LivenessResponseStatus
	Storage       DependencyStatus
	Cache         DependencyStatus
	Queue         DependencyStatus
}
