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

type CurrentServiceStatus string

const (
	CurrentServiceStatusUnknown     CurrentServiceStatus = "UNKNOWN"
	CurrentServiceStatusHealthy     CurrentServiceStatus = "HEALTHY"
	CurrentServiceStatusDegraded    CurrentServiceStatus = "DEGRADED"
	CurrentServiceStatusUnavailable CurrentServiceStatus = "UNAVAILABLE"
)

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

func DeriveCurrentServiceStatus(
	latestHealthEvent *HealthEvent,
) CurrentServiceStatus {
	if latestHealthEvent == nil {
		return CurrentServiceStatusUnknown
	}

	switch latestHealthEvent.Status {
	case HealthStatusHealthy:
		return CurrentServiceStatusHealthy

	case HealthStatusDegraded:
		return CurrentServiceStatusDegraded

	case HealthStatusUnavailable:
		return CurrentServiceStatusUnavailable

	default:
		return CurrentServiceStatusUnknown
	}
}
