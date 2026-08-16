package domain

import (
	"time"

	"github.com/google/uuid"
)

type BusinessJourney struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	Name        string
	Criticality Criticality
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
