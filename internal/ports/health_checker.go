package ports

import (
	"context"

	"github.com/architeacher/svc-web-analyzer/internal/domain"
)

// HealthChecker defines the interface for checking system health
type HealthChecker interface {
	// CheckReadiness performs readiness check and returns detailed results
	CheckReadiness(ctx context.Context) *domain.ReadinessResult

	// CheckLiveness performs liveness check and returns detailed results
	CheckLiveness(ctx context.Context) *domain.LivenessResult

	// CheckHealth performs a comprehensive health check and returns detailed results
	CheckHealth(ctx context.Context) *domain.HealthResult
}
