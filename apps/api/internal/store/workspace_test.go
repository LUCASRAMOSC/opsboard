package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/database"
)

func TestWorkspaceStoreIntegration(t *testing.T) {
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

	name := "integration-test-" + uuid.NewString()

	workspace, err := store.CreateWorkspace(ctx, name)
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

	if workspace.ID == uuid.Nil {
		t.Error("expected generated workspace UUID")
	}

	if workspace.Name != name {
		t.Errorf("workspace name = %q, want %q", workspace.Name, name)
	}

	found, err := store.GetWorkspace(ctx, workspace.ID)
	if err != nil {
		t.Fatalf("get workspace: %v", err)
	}

	if found.ID != workspace.ID {
		t.Errorf("workspace ID = %s, want %s", found.ID, workspace.ID)
	}

	if found.Name != workspace.Name {
		t.Errorf("workspace name = %q, want %q", found.Name, workspace.Name)
	}

	workspaces, err := store.ListWorkspaces(ctx)
	if err != nil {
		t.Fatalf("list workspaces: %v", err)
	}

	var listed bool

	for _, item := range workspaces {
		if item.ID == workspace.ID {
			listed = true

			break
		}
	}

	if !listed {
		t.Error("created workspace was not returned by ListWorkspaces")
	}
}
