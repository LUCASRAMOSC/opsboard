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

func newBusinessJourneyTestStore(
	t *testing.T,
) (context.Context, *Store) {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	t.Cleanup(cancel)

	db, err := database.NewPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	t.Cleanup(db.Close)

	return ctx, New(db)
}

func createBusinessJourneyTestWorkspace(
	t *testing.T,
	ctx context.Context,
	dataStore *Store,
) domain.Workspace {
	t.Helper()

	workspace, err := dataStore.CreateWorkspace(
		ctx,
		"journey-test-"+uuid.NewString(),
	)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if _, err := dataStore.db.Exec(
			cleanupCtx,
			"DELETE FROM workspaces WHERE id = $1",
			workspace.ID,
		); err != nil {
			t.Errorf("cleanup workspace: %v", err)
		}
	})

	return workspace
}

func createBusinessJourneyTestService(
	t *testing.T,
	ctx context.Context,
	dataStore *Store,
	workspaceID uuid.UUID,
	name string,
) domain.Service {
	t.Helper()

	service, err := dataStore.CreateService(
		ctx,
		workspaceID,
		name,
		domain.ServiceTypeAPI,
		domain.CriticalityHigh,
	)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	return service
}

func createBusinessJourneyTestJourney(
	t *testing.T,
	ctx context.Context,
	dataStore *Store,
	workspaceID uuid.UUID,
	name string,
) domain.BusinessJourney {
	t.Helper()

	journey, err := dataStore.CreateBusinessJourney(
		ctx,
		workspaceID,
		name,
		domain.CriticalityCritical,
	)
	if err != nil {
		t.Fatalf("create business journey: %v", err)
	}

	return journey
}

func TestBusinessJourneyStoreIntegration(t *testing.T) {
	ctx, dataStore := newBusinessJourneyTestStore(t)

	workspace := createBusinessJourneyTestWorkspace(
		t,
		ctx,
		dataStore,
	)

	created := createBusinessJourneyTestJourney(
		t,
		ctx,
		dataStore,
		workspace.ID,
		"Checkout",
	)

	if created.ID == uuid.Nil {
		t.Fatal("business journey ID is nil")
	}

	if created.WorkspaceID != workspace.ID {
		t.Fatalf(
			"workspace ID = %s, want %s",
			created.WorkspaceID,
			workspace.ID,
		)
	}

	if created.Name != "Checkout" {
		t.Fatalf(
			"name = %q, want %q",
			created.Name,
			"Checkout",
		)
	}

	if created.Criticality != domain.CriticalityCritical {
		t.Fatalf(
			"criticality = %q, want %q",
			created.Criticality,
			domain.CriticalityCritical,
		)
	}

	retrieved, err := dataStore.GetBusinessJourney(
		ctx,
		created.ID,
	)
	if err != nil {
		t.Fatalf("get business journey: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Fatalf(
			"retrieved ID = %s, want %s",
			retrieved.ID,
			created.ID,
		)
	}

	journeys, err := dataStore.ListBusinessJourneysByWorkspace(
		ctx,
		workspace.ID,
	)
	if err != nil {
		t.Fatalf("list business journeys: %v", err)
	}

	if len(journeys) != 1 {
		t.Fatalf(
			"journeys length = %d, want 1",
			len(journeys),
		)
	}
}

func TestBusinessJourneySupportsManyToManyServices(
	t *testing.T,
) {
	ctx, dataStore := newBusinessJourneyTestStore(t)

	workspace := createBusinessJourneyTestWorkspace(
		t,
		ctx,
		dataStore,
	)

	payments := createBusinessJourneyTestService(
		t,
		ctx,
		dataStore,
		workspace.ID,
		"Payments API",
	)

	orders := createBusinessJourneyTestService(
		t,
		ctx,
		dataStore,
		workspace.ID,
		"Orders API",
	)

	checkout := createBusinessJourneyTestJourney(
		t,
		ctx,
		dataStore,
		workspace.ID,
		"Checkout",
	)

	orderTracking := createBusinessJourneyTestJourney(
		t,
		ctx,
		dataStore,
		workspace.ID,
		"Order Tracking",
	)

	if err := dataStore.AddServiceToBusinessJourney(
		ctx,
		checkout.ID,
		payments.ID,
	); err != nil {
		t.Fatalf("add Payments API to Checkout: %v", err)
	}

	if err := dataStore.AddServiceToBusinessJourney(
		ctx,
		checkout.ID,
		orders.ID,
	); err != nil {
		t.Fatalf("add Orders API to Checkout: %v", err)
	}

	if err := dataStore.AddServiceToBusinessJourney(
		ctx,
		orderTracking.ID,
		orders.ID,
	); err != nil {
		t.Fatalf(
			"add Orders API to Order Tracking: %v",
			err,
		)
	}

	checkoutServices, err := dataStore.ListServicesByBusinessJourney(
		ctx,
		checkout.ID,
	)
	if err != nil {
		t.Fatalf("list Checkout services: %v", err)
	}

	if len(checkoutServices) != 2 {
		t.Fatalf(
			"Checkout services length = %d, want 2",
			len(checkoutServices),
		)
	}

	orderTrackingServices, err := dataStore.ListServicesByBusinessJourney(
		ctx,
		orderTracking.ID,
	)
	if err != nil {
		t.Fatalf(
			"list Order Tracking services: %v",
			err,
		)
	}

	if len(orderTrackingServices) != 1 {
		t.Fatalf(
			"Order Tracking services length = %d, want 1",
			len(orderTrackingServices),
		)
	}

	if orderTrackingServices[0].ID != orders.ID {
		t.Fatalf(
			"service ID = %s, want %s",
			orderTrackingServices[0].ID,
			orders.ID,
		)
	}
}

func TestAddServiceToBusinessJourneyRejectsCrossWorkspace(
	t *testing.T,
) {
	ctx, dataStore := newBusinessJourneyTestStore(t)

	journeyWorkspace := createBusinessJourneyTestWorkspace(
		t,
		ctx,
		dataStore,
	)

	serviceWorkspace := createBusinessJourneyTestWorkspace(
		t,
		ctx,
		dataStore,
	)

	journey := createBusinessJourneyTestJourney(
		t,
		ctx,
		dataStore,
		journeyWorkspace.ID,
		"Checkout",
	)

	service := createBusinessJourneyTestService(
		t,
		ctx,
		dataStore,
		serviceWorkspace.ID,
		"Payments API",
	)

	err := dataStore.AddServiceToBusinessJourney(
		ctx,
		journey.ID,
		service.ID,
	)

	if !errors.Is(err, ErrWorkspaceMismatch) {
		t.Fatalf(
			"error = %v, want ErrWorkspaceMismatch",
			err,
		)
	}
}

func TestAddServiceToBusinessJourneyReturnsConflictForDuplicateAssociation(
	t *testing.T,
) {
	ctx, dataStore := newBusinessJourneyTestStore(t)

	workspace := createBusinessJourneyTestWorkspace(
		t,
		ctx,
		dataStore,
	)

	journey := createBusinessJourneyTestJourney(
		t,
		ctx,
		dataStore,
		workspace.ID,
		"Checkout",
	)

	service := createBusinessJourneyTestService(
		t,
		ctx,
		dataStore,
		workspace.ID,
		"Payments API",
	)

	if err := dataStore.AddServiceToBusinessJourney(
		ctx,
		journey.ID,
		service.ID,
	); err != nil {
		t.Fatalf("add service: %v", err)
	}

	err := dataStore.AddServiceToBusinessJourney(
		ctx,
		journey.ID,
		service.ID,
	)

	if !errors.Is(err, ErrConflict) {
		t.Fatalf(
			"error = %v, want ErrConflict",
			err,
		)
	}
}

func TestRemoveServiceFromBusinessJourneyKeepsService(
	t *testing.T,
) {
	ctx, dataStore := newBusinessJourneyTestStore(t)

	workspace := createBusinessJourneyTestWorkspace(
		t,
		ctx,
		dataStore,
	)

	journey := createBusinessJourneyTestJourney(
		t,
		ctx,
		dataStore,
		workspace.ID,
		"Checkout",
	)

	service := createBusinessJourneyTestService(
		t,
		ctx,
		dataStore,
		workspace.ID,
		"Payments API",
	)

	if err := dataStore.AddServiceToBusinessJourney(
		ctx,
		journey.ID,
		service.ID,
	); err != nil {
		t.Fatalf("add service: %v", err)
	}

	if err := dataStore.RemoveServiceFromBusinessJourney(
		ctx,
		journey.ID,
		service.ID,
	); err != nil {
		t.Fatalf("remove service: %v", err)
	}

	services, err := dataStore.ListServicesByBusinessJourney(
		ctx,
		journey.ID,
	)
	if err != nil {
		t.Fatalf("list journey services: %v", err)
	}

	if len(services) != 0 {
		t.Fatalf(
			"services length = %d, want 0",
			len(services),
		)
	}

	retrievedService, err := dataStore.GetService(
		ctx,
		service.ID,
	)
	if err != nil {
		t.Fatalf(
			"get service after association removal: %v",
			err,
		)
	}

	if retrievedService.ID != service.ID {
		t.Fatalf(
			"service ID = %s, want %s",
			retrievedService.ID,
			service.ID,
		)
	}
}
