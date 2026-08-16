package domain

import (
	"time"

	"github.com/google/uuid"
)

type HealthStatus string

const (
	HealthStatusHealthy     HealthStatus = "HEALTHY"
	HealthStatusDegraded    HealthStatus = "DEGRADED"
	HealthStatusUnavailable HealthStatus = "UNAVAILABLE"
)

func (s HealthStatus) Valid() bool {
	switch s {
	case HealthStatusHealthy,
		HealthStatusDegraded,
		HealthStatusUnavailable:
		return true
	default:
		return false
	}
}

func ValidResponseTimeMs(responseTimeMs *int) bool {
	return responseTimeMs == nil || *responseTimeMs >= 0
}

type HealthEvent struct {
	ID             uuid.UUID
	ServiceID      uuid.UUID
	Status         HealthStatus
	ResponseTimeMs *int
	ObservedAt     time.Time
	CreatedAt      time.Time
}
