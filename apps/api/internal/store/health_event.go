package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/domain"
)

func (s *Store) CreateHealthEvent(
	ctx context.Context,
	serviceID uuid.UUID,
	status domain.HealthStatus,
	responseTimeMs *int,
	observedAt time.Time,
) (domain.HealthEvent, error) {
	const query = `
		INSERT INTO health_events (
			service_id,
			status,
			response_time_ms,
			observed_at
		)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id,
			service_id,
			status,
			response_time_ms,
			observed_at,
			created_at
	`

	var healthEvent domain.HealthEvent

	err := s.db.QueryRow(
		ctx,
		query,
		serviceID,
		status,
		responseTimeMs,
		observedAt,
	).Scan(
		&healthEvent.ID,
		&healthEvent.ServiceID,
		&healthEvent.Status,
		&healthEvent.ResponseTimeMs,
		&healthEvent.ObservedAt,
		&healthEvent.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.ConstraintName == "health_events_service_fk" {
			return domain.HealthEvent{}, ErrNotFound
		}

		return domain.HealthEvent{}, err
	}

	return healthEvent, nil
}

func (s *Store) ListHealthEventsByService(
	ctx context.Context,
	serviceID uuid.UUID,
) ([]domain.HealthEvent, error) {
	if _, err := s.GetService(ctx, serviceID); err != nil {
		return nil, err
	}

	const query = `
		SELECT
			id,
			service_id,
			status,
			response_time_ms,
			observed_at,
			created_at
		FROM health_events
		WHERE service_id = $1
		ORDER BY observed_at ASC, created_at ASC
	`

	rows, err := s.db.Query(ctx, query, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	healthEvents := make([]domain.HealthEvent, 0)

	for rows.Next() {
		var healthEvent domain.HealthEvent

		if err := rows.Scan(
			&healthEvent.ID,
			&healthEvent.ServiceID,
			&healthEvent.Status,
			&healthEvent.ResponseTimeMs,
			&healthEvent.ObservedAt,
			&healthEvent.CreatedAt,
		); err != nil {
			return nil, err
		}

		healthEvents = append(healthEvents, healthEvent)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return healthEvents, nil
}

func (s *Store) GetLatestHealthEvent(
	ctx context.Context,
	serviceID uuid.UUID,
) (domain.HealthEvent, error) {
	if _, err := s.GetService(ctx, serviceID); err != nil {
		return domain.HealthEvent{}, err
	}

	const query = `
		SELECT
			id,
			service_id,
			status,
			response_time_ms,
			observed_at,
			created_at
		FROM health_events
		WHERE service_id = $1
		ORDER BY observed_at DESC, created_at DESC
		LIMIT 1
	`

	var healthEvent domain.HealthEvent

	err := s.db.QueryRow(
		ctx,
		query,
		serviceID,
	).Scan(
		&healthEvent.ID,
		&healthEvent.ServiceID,
		&healthEvent.Status,
		&healthEvent.ResponseTimeMs,
		&healthEvent.ObservedAt,
		&healthEvent.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.HealthEvent{}, ErrNoHealthEvents
	}

	if err != nil {
		return domain.HealthEvent{}, err
	}

	return healthEvent, nil
}
