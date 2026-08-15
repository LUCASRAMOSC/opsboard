package store

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/database"
	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/domain"
)

func TestServiceStoreIntegration(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.NewPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	t.Cleanup(db.Close)

	store := New(db)

	workspace, err := store.CreateWorkspace(
		ctx,
		"service-integration-"+uuid.NewString(),
	)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cleanupCancel()

		if _, err := db.Exec(
			cleanupCtx,
			"DELETE FROM workspaces WHERE id = $1",
			workspace.ID,
		); err != nil {
			t.Errorf("cleanup workspace: %v", err)
		}
	})

	name := "Payments API"

	service, err := store.CreateService(
		ctx,
		workspace.ID,
		name,
		domain.ServiceTypeAPI,
		domain.CriticalityHigh,
	)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	if service.ID == uuid.Nil {
		t.Error("expected generated service UUID")
	}

	if service.WorkspaceID != workspace.ID {
		t.Errorf(
			"workspace ID = %s, want %s",
			service.WorkspaceID,
			workspace.ID,
		)
	}

	if service.Name != name {
		t.Errorf("service name = %q, want %q", service.Name, name)
	}

	if service.Type != domain.ServiceTypeAPI {
		t.Errorf(
			"service type = %q, want %q",
			service.Type,
			domain.ServiceTypeAPI,
		)
	}

	if service.Criticality != domain.CriticalityHigh {
		t.Errorf(
			"service criticality = %q, want %q",
			service.Criticality,
			domain.CriticalityHigh,
		)
	}

	found, err := store.GetService(ctx, service.ID)
	if err != nil {
		t.Fatalf("get service: %v", err)
	}

	if found.ID != service.ID {
		t.Errorf("service ID = %s, want %s", found.ID, service.ID)
	}

	services, err := store.ListServicesByWorkspace(ctx, workspace.ID)
	if err != nil {
		t.Fatalf("list services: %v", err)
	}

	var listed bool

	for _, item := range services {
		if item.ID == service.ID {
			listed = true

			break
		}
	}

	if !listed {
		t.Error("created service was not returned by ListServicesByWorkspace")
	}
}

func TestCreateServiceReturnsNotFoundForMissingWorkspace(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.NewPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	t.Cleanup(db.Close)

	store := New(db)

	_, err = store.CreateService(
		ctx,
		uuid.New(),
		"Payments API",
		domain.ServiceTypeAPI,
		domain.CriticalityHigh,
	)

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestCreateServiceReturnsConflictForDuplicateName(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.NewPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	t.Cleanup(db.Close)

	store := New(db)

	workspace, err := store.CreateWorkspace(
		ctx,
		"duplicate-service-"+uuid.NewString(),
	)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cleanupCancel()

		if _, err := db.Exec(
			cleanupCtx,
			"DELETE FROM workspaces WHERE id = $1",
			workspace.ID,
		); err != nil {
			t.Errorf("cleanup workspace: %v", err)
		}
	})

	_, err = store.CreateService(
		ctx,
		workspace.ID,
		"Payments API",
		domain.ServiceTypeAPI,
		domain.CriticalityHigh,
	)
	if err != nil {
		t.Fatalf("create first service: %v", err)
	}

	_, err = store.CreateService(
		ctx,
		workspace.ID,
		"Payments API",
		domain.ServiceTypeAPI,
		domain.CriticalityHigh,
	)

	if !errors.Is(err, ErrConflict) {
		t.Fatalf("error = %v, want ErrConflict", err)
	}
}

func TestGetServiceReturnsNotFound(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.NewPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	t.Cleanup(db.Close)

	store := New(db)

	_, err = store.GetService(ctx, uuid.New())

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestListServicesByWorkspaceReturnsNotFoundForMissingWorkspace(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.NewPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	t.Cleanup(db.Close)

	store := New(db)

	_, err = store.ListServicesByWorkspace(ctx, uuid.New())

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}
