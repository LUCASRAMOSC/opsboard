package store

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/domain"
)

func (s *Store) CreateBusinessJourney(
	ctx context.Context,
	workspaceID uuid.UUID,
	name string,
	criticality domain.Criticality,
) (domain.BusinessJourney, error) {
	const query = `
		INSERT INTO business_journeys (
			workspace_id,
			name,
			criticality
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			workspace_id,
			name,
			criticality,
			created_at,
			updated_at
	`

	var journey domain.BusinessJourney

	err := s.db.QueryRow(
		ctx,
		query,
		workspaceID,
		name,
		criticality,
	).Scan(
		&journey.ID,
		&journey.WorkspaceID,
		&journey.Name,
		&journey.Criticality,
		&journey.CreatedAt,
		&journey.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.ConstraintName {
			case "business_journeys_workspace_fk":
				return domain.BusinessJourney{}, ErrNotFound

			case "business_journeys_workspace_name_unique":
				return domain.BusinessJourney{}, ErrConflict
			}
		}

		return domain.BusinessJourney{}, err
	}

	return journey, nil
}

func (s *Store) GetBusinessJourney(
	ctx context.Context,
	id uuid.UUID,
) (domain.BusinessJourney, error) {
	const query = `
		SELECT
			id,
			workspace_id,
			name,
			criticality,
			created_at,
			updated_at
		FROM business_journeys
		WHERE id = $1
	`

	var journey domain.BusinessJourney

	err := s.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&journey.ID,
		&journey.WorkspaceID,
		&journey.Name,
		&journey.Criticality,
		&journey.CreatedAt,
		&journey.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.BusinessJourney{}, ErrNotFound
	}

	if err != nil {
		return domain.BusinessJourney{}, err
	}

	return journey, nil
}

func (s *Store) ListBusinessJourneysByWorkspace(
	ctx context.Context,
	workspaceID uuid.UUID,
) ([]domain.BusinessJourney, error) {
	if _, err := s.GetWorkspace(ctx, workspaceID); err != nil {
		return nil, err
	}

	const query = `
		SELECT
			id,
			workspace_id,
			name,
			criticality,
			created_at,
			updated_at
		FROM business_journeys
		WHERE workspace_id = $1
		ORDER BY created_at ASC, id ASC
	`

	rows, err := s.db.Query(
		ctx,
		query,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	journeys := make([]domain.BusinessJourney, 0)

	for rows.Next() {
		var journey domain.BusinessJourney

		if err := rows.Scan(
			&journey.ID,
			&journey.WorkspaceID,
			&journey.Name,
			&journey.Criticality,
			&journey.CreatedAt,
			&journey.UpdatedAt,
		); err != nil {
			return nil, err
		}

		journeys = append(journeys, journey)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return journeys, nil
}

func (s *Store) AddServiceToBusinessJourney(
	ctx context.Context,
	journeyID uuid.UUID,
	serviceID uuid.UUID,
) error {
	journey, err := s.GetBusinessJourney(
		ctx,
		journeyID,
	)
	if err != nil {
		return err
	}

	service, err := s.GetService(
		ctx,
		serviceID,
	)
	if err != nil {
		return err
	}

	if journey.WorkspaceID != service.WorkspaceID {
		return ErrWorkspaceMismatch
	}

	const query = `
		INSERT INTO business_journey_services (
			business_journey_id,
			service_id
		)
		VALUES ($1, $2)
	`

	_, err = s.db.Exec(
		ctx,
		query,
		journeyID,
		serviceID,
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.ConstraintName {
			case "business_journey_services_pk":
				return ErrConflict

			case "business_journey_services_journey_fk",
				"business_journey_services_service_fk":
				return ErrNotFound
			}
		}

		return err
	}

	return nil
}

func (s *Store) RemoveServiceFromBusinessJourney(
	ctx context.Context,
	journeyID uuid.UUID,
	serviceID uuid.UUID,
) error {
	const query = `
		DELETE FROM business_journey_services
		WHERE business_journey_id = $1
			AND service_id = $2
	`

	result, err := s.db.Exec(
		ctx,
		query,
		journeyID,
		serviceID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *Store) ListServicesByBusinessJourney(
	ctx context.Context,
	journeyID uuid.UUID,
) ([]domain.Service, error) {
	if _, err := s.GetBusinessJourney(ctx, journeyID); err != nil {
		return nil, err
	}

	const query = `
		SELECT
			s.id,
			s.workspace_id,
			s.name,
			s.type,
			s.criticality,
			s.created_at,
			s.updated_at
		FROM services s
		INNER JOIN business_journey_services bjs
			ON bjs.service_id = s.id
		WHERE bjs.business_journey_id = $1
		ORDER BY bjs.created_at ASC, s.id ASC
	`

	rows, err := s.db.Query(
		ctx,
		query,
		journeyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := make([]domain.Service, 0)

	for rows.Next() {
		var service domain.Service

		if err := rows.Scan(
			&service.ID,
			&service.WorkspaceID,
			&service.Name,
			&service.Type,
			&service.Criticality,
			&service.CreatedAt,
			&service.UpdatedAt,
		); err != nil {
			return nil, err
		}

		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}
