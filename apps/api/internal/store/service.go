package store

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/domain"
)

func (s *Store) CreateService(
	ctx context.Context,
	workspaceID uuid.UUID,
	name string,
	serviceType domain.ServiceType,
	criticality domain.Criticality,
) (domain.Service, error) {
	const query = `
		INSERT INTO services (
			workspace_id,
			name,
			type,
			criticality
		)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id,
			workspace_id,
			name,
			type,
			criticality,
			created_at,
			updated_at
	`

	var service domain.Service

	err := s.db.QueryRow(
		ctx,
		query,
		workspaceID,
		name,
		serviceType,
		criticality,
	).Scan(
		&service.ID,
		&service.WorkspaceID,
		&service.Name,
		&service.Type,
		&service.Criticality,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.ConstraintName {
			case "services_workspace_fk":
				return domain.Service{}, ErrNotFound

			case "services_workspace_name_unique":
				return domain.Service{}, ErrConflict
			}
		}

		return domain.Service{}, err
	}

	return service, nil
}

func (s *Store) GetService(
	ctx context.Context,
	id uuid.UUID,
) (domain.Service, error) {
	const query = `
		SELECT
			id,
			workspace_id,
			name,
			type,
			criticality,
			created_at,
			updated_at
		FROM services
		WHERE id = $1
	`

	var service domain.Service

	err := s.db.QueryRow(ctx, query, id).Scan(
		&service.ID,
		&service.WorkspaceID,
		&service.Name,
		&service.Type,
		&service.Criticality,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Service{}, ErrNotFound
	}

	if err != nil {
		return domain.Service{}, err
	}

	return service, nil
}

func (s *Store) ListServicesByWorkspace(
	ctx context.Context,
	workspaceID uuid.UUID,
) ([]domain.Service, error) {
	const query = `
		SELECT
			id,
			workspace_id,
			name,
			type,
			criticality,
			created_at,
			updated_at
		FROM services
		WHERE workspace_id = $1
		ORDER BY created_at ASC
	`

	rows, err := s.db.Query(ctx, query, workspaceID)
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
