package domain

import (
	"time"

	"github.com/google/uuid"
)

type ServiceType string

const (
	ServiceTypeFrontend ServiceType = "FRONTEND"
	ServiceTypeAPI      ServiceType = "API"
	ServiceTypeDatabase ServiceType = "DATABASE"
)

func (t ServiceType) Valid() bool {
	switch t {
	case ServiceTypeFrontend, ServiceTypeAPI, ServiceTypeDatabase:
		return true
	default:
		return false
	}
}

type Criticality string

const (
	CriticalityLow      Criticality = "LOW"
	CriticalityMedium   Criticality = "MEDIUM"
	CriticalityHigh     Criticality = "HIGH"
	CriticalityCritical Criticality = "CRITICAL"
)

func (c Criticality) Valid() bool {
	switch c {
	case CriticalityLow, CriticalityMedium, CriticalityHigh, CriticalityCritical:
		return true
	default:
		return false
	}
}

type Service struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	Name        string
	Type        ServiceType
	Criticality Criticality
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
