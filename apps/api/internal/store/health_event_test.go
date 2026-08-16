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

func newHealthEventTestStore(
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

func createHealthEventTestService(
	t *testing.T,
	ctx context.Context,
	dataStore *Store,
) domain.Service {
	t.Helper()

	workspace, err := dataStore.CreateWorkspace(
		ctx,
		"health-event-test-"+uuid.NewString(),
	)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	service, err := dataStore.CreateService(
		ctx,
		workspace.ID,
		"Payments API",
		domain.ServiceTypeAPI,
		domain.CriticalityHigh,
	)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cleanupCancel()

		if _, err := dataStore.db.Exec(
			cleanupCtx,
			"DELETE FROM health_events WHERE service_id = $1",
			service.ID,
		); err != nil {
			t.Errorf("cleanup health events: %v", err)
		}

		if _, err := dataStore.db.Exec(
			cleanupCtx,
			"DELETE FROM workspaces WHERE id = $1",
			workspace.ID,
		); err != nil {
			t.Errorf("cleanup workspace: %v", err)
		}
	})

	return service
}

func TestHealthEventStoreIntegration(t *testing.T) {
	ctx, dataStore := newHealthEventTestStore(t)

	service := createHealthEventTestService(
		t,
		ctx,
		dataStore,
	)

	responseTimeMs := 87
	observedAt := time.Now().
		UTC().
		Truncate(time.Microsecond)

	created, err := dataStore.CreateHealthEvent(
		ctx,
		service.ID,
		domain.HealthStatusHealthy,
		&responseTimeMs,
		observedAt,
	)
	if err != nil {
		t.Fatalf("create health event: %v", err)
	}

	if created.ID == uuid.Nil {
		t.Fatal("health event ID is nil")
	}

	if created.ServiceID != service.ID {
		t.Fatalf(
			"service ID = %s, want %s",
			created.ServiceID,
			service.ID,
		)
	}

	if created.Status != domain.HealthStatusHealthy {
		t.Fatalf(
			"status = %q, want %q",
			created.Status,
			domain.HealthStatusHealthy,
		)
	}

	if created.ResponseTimeMs == nil {
		t.Fatal("response time is nil")
	}

	if *created.ResponseTimeMs != responseTimeMs {
		t.Fatalf(
			"response time = %d, want %d",
			*created.ResponseTimeMs,
			responseTimeMs,
		)
	}

	if !created.ObservedAt.Equal(observedAt) {
		t.Fatalf(
			"observed at = %s, want %s",
			created.ObservedAt,
			observedAt,
		)
	}

	if created.CreatedAt.IsZero() {
		t.Fatal("created at is zero")
	}

	history, err := dataStore.ListHealthEventsByService(
		ctx,
		service.ID,
	)
	if err != nil {
		t.Fatalf("list health events: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf(
			"history length = %d, want 1",
			len(history),
		)
	}

	if history[0].ID != created.ID {
		t.Fatalf(
			"persisted health event ID = %s, want %s",
			history[0].ID,
			created.ID,
		)
	}
}

func TestHealthEventHistoryPreservesEvents(t *testing.T) {
	ctx, dataStore := newHealthEventTestStore(t)

	service := createHealthEventTestService(
		t,
		ctx,
		dataStore,
	)

	baseTime := time.Now().
		UTC().
		Truncate(time.Microsecond)

	events := []struct {
		status     domain.HealthStatus
		observedAt time.Time
	}{
		{
			status:     domain.HealthStatusHealthy,
			observedAt: baseTime,
		},
		{
			status:     domain.HealthStatusDegraded,
			observedAt: baseTime.Add(time.Minute),
		},
		{
			status:     domain.HealthStatusUnavailable,
			observedAt: baseTime.Add(2 * time.Minute),
		},
	}

	createdIDs := make([]uuid.UUID, 0, len(events))

	for _, event := range events {
		created, err := dataStore.CreateHealthEvent(
			ctx,
			service.ID,
			event.status,
			nil,
			event.observedAt,
		)
		if err != nil {
			t.Fatalf(
				"create health event %q: %v",
				event.status,
				err,
			)
		}

		createdIDs = append(createdIDs, created.ID)
	}

	history, err := dataStore.ListHealthEventsByService(
		ctx,
		service.ID,
	)
	if err != nil {
		t.Fatalf("list health events: %v", err)
	}

	if len(history) != len(events) {
		t.Fatalf(
			"history length = %d, want %d",
			len(history),
			len(events),
		)
	}

	for i, expected := range events {
		if history[i].ID != createdIDs[i] {
			t.Errorf(
				"event %d ID = %s, want %s",
				i,
				history[i].ID,
				createdIDs[i],
			)
		}

		if history[i].Status != expected.status {
			t.Errorf(
				"event %d status = %q, want %q",
				i,
				history[i].Status,
				expected.status,
			)
		}

		if !history[i].ObservedAt.Equal(expected.observedAt) {
			t.Errorf(
				"event %d observed at = %s, want %s",
				i,
				history[i].ObservedAt,
				expected.observedAt,
			)
		}
	}
}

func TestGetLatestHealthEventUsesObservedAt(t *testing.T) {
	ctx, dataStore := newHealthEventTestStore(t)

	service := createHealthEventTestService(
		t,
		ctx,
		dataStore,
	)

	baseTime := time.Now().
		UTC().
		Truncate(time.Microsecond)

	latestObservedAt := baseTime.Add(10 * time.Minute)

	latestCreated, err := dataStore.CreateHealthEvent(
		ctx,
		service.ID,
		domain.HealthStatusDegraded,
		nil,
		latestObservedAt,
	)
	if err != nil {
		t.Fatalf("create latest health event: %v", err)
	}

	_, err = dataStore.CreateHealthEvent(
		ctx,
		service.ID,
		domain.HealthStatusHealthy,
		nil,
		baseTime,
	)
	if err != nil {
		t.Fatalf("create older health event: %v", err)
	}

	latest, err := dataStore.GetLatestHealthEvent(
		ctx,
		service.ID,
	)
	if err != nil {
		t.Fatalf("get latest health event: %v", err)
	}

	if latest.ID != latestCreated.ID {
		t.Fatalf(
			"latest ID = %s, want %s",
			latest.ID,
			latestCreated.ID,
		)
	}

	if latest.Status != domain.HealthStatusDegraded {
		t.Fatalf(
			"latest status = %q, want %q",
			latest.Status,
			domain.HealthStatusDegraded,
		)
	}

	if !latest.ObservedAt.Equal(latestObservedAt) {
		t.Fatalf(
			"latest observed at = %s, want %s",
			latest.ObservedAt,
			latestObservedAt,
		)
	}
}

func TestGetLatestHealthEventReturnsNoHealthEvents(
	t *testing.T,
) {
	ctx, dataStore := newHealthEventTestStore(t)

	service := createHealthEventTestService(
		t,
		ctx,
		dataStore,
	)

	_, err := dataStore.GetLatestHealthEvent(
		ctx,
		service.ID,
	)

	if !errors.Is(err, ErrNoHealthEvents) {
		t.Fatalf(
			"error = %v, want ErrNoHealthEvents",
			err,
		)
	}
}

func TestListHealthEventsReturnsEmptyForServiceWithoutEvents(
	t *testing.T,
) {
	ctx, dataStore := newHealthEventTestStore(t)

	service := createHealthEventTestService(
		t,
		ctx,
		dataStore,
	)

	history, err := dataStore.ListHealthEventsByService(
		ctx,
		service.ID,
	)
	if err != nil {
		t.Fatalf("list health events: %v", err)
	}

	if history == nil {
		t.Fatal("history is nil, want empty slice")
	}

	if len(history) != 0 {
		t.Fatalf(
			"history length = %d, want 0",
			len(history),
		)
	}
}

func TestHealthEventStoreReturnsNotFoundForMissingService(
	t *testing.T,
) {
	ctx, dataStore := newHealthEventTestStore(t)

	serviceID := uuid.New()
	observedAt := time.Now().
		UTC().
		Truncate(time.Microsecond)

	t.Run("create", func(t *testing.T) {
		_, err := dataStore.CreateHealthEvent(
			ctx,
			serviceID,
			domain.HealthStatusHealthy,
			nil,
			observedAt,
		)

		if !errors.Is(err, ErrNotFound) {
			t.Fatalf(
				"error = %v, want ErrNotFound",
				err,
			)
		}
	})

	t.Run("list", func(t *testing.T) {
		_, err := dataStore.ListHealthEventsByService(
			ctx,
			serviceID,
		)

		if !errors.Is(err, ErrNotFound) {
			t.Fatalf(
				"error = %v, want ErrNotFound",
				err,
			)
		}
	})

	t.Run("latest", func(t *testing.T) {
		_, err := dataStore.GetLatestHealthEvent(
			ctx,
			serviceID,
		)

		if !errors.Is(err, ErrNotFound) {
			t.Fatalf(
				"error = %v, want ErrNotFound",
				err,
			)
		}
	})
}
