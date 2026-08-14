package store

import (
	"context"

	"github.com/google/uuid"

	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/domain"
)

func (s *Store) CreateWorkspace(
	ctx context.Context,
	name string,
) (domain.Workspace, error) {
	const query = `
		INSERT INTO workspaces (name)
		VALUES ($1)
		RETURNING id, name, created_at, updated_at
	`

	var workspace domain.Workspace

	err := s.db.QueryRow(ctx, query, name).Scan(
		&workspace.ID,
		&workspace.Name,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	)
	if err != nil {
		return domain.Workspace{}, err
	}

	return workspace, nil
}

func (s *Store) GetWorkspace(
	ctx context.Context,
	id uuid.UUID,
) (domain.Workspace, error) {
	const query = `
		SELECT id, name, created_at, updated_at
		FROM workspaces
		WHERE id = $1
	`

	var workspace domain.Workspace

	err := s.db.QueryRow(ctx, query, id).Scan(
		&workspace.ID,
		&workspace.Name,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	)
	if err != nil {
		return domain.Workspace{}, err
	}

	return workspace, nil
}

func (s *Store) ListWorkspaces(
	ctx context.Context,
) ([]domain.Workspace, error) {
	const query = `
		SELECT id, name, created_at, updated_at
		FROM workspaces
		ORDER BY created_at ASC
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	workspaces := make([]domain.Workspace, 0)

	for rows.Next() {
		var workspace domain.Workspace

		if err := rows.Scan(
			&workspace.ID,
			&workspace.Name,
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
		); err != nil {
			return nil, err
		}

		workspaces = append(workspaces, workspace)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return workspaces, nil
}
