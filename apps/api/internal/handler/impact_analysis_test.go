package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/domain"
	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/store"
)

type impactAnalysisStoreStub struct {
	services     []domain.Service
	journeys     []domain.BusinessJourney
	healthEvents map[uuid.UUID]domain.HealthEvent
	dependencies map[uuid.UUID][]domain.Service

	listServicesErr error
}

func (s *impactAnalysisStoreStub) ListServicesByWorkspace(
	_ context.Context,
	_ uuid.UUID,
) ([]domain.Service, error) {
	if s.listServicesErr != nil {
		return nil, s.listServicesErr
	}

	return s.services, nil
}

func (s *impactAnalysisStoreStub) GetLatestHealthEvent(
	_ context.Context,
	serviceID uuid.UUID,
) (domain.HealthEvent, error) {
	event, ok := s.healthEvents[serviceID]
	if !ok {
		return domain.HealthEvent{}, store.ErrNoHealthEvents
	}

	return event, nil
}

func (s *impactAnalysisStoreStub) ListBusinessJourneysByWorkspace(
	_ context.Context,
	_ uuid.UUID,
) ([]domain.BusinessJourney, error) {
	return s.journeys, nil
}

func (s *impactAnalysisStoreStub) ListServicesByBusinessJourney(
	_ context.Context,
	journeyID uuid.UUID,
) ([]domain.Service, error) {
	return s.dependencies[journeyID], nil
}

func TestImpactAnalysisHandlerReturnsAffectedJourney(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	workspaceID := uuid.New()
	paymentsID := uuid.New()
	ordersID := uuid.New()
	checkoutID := uuid.New()

	payments := domain.Service{
		ID:          paymentsID,
		WorkspaceID: workspaceID,
		Name:        "Payments API",
		Type:        domain.ServiceTypeAPI,
		Criticality: domain.CriticalityHigh,
	}

	orders := domain.Service{
		ID:          ordersID,
		WorkspaceID: workspaceID,
		Name:        "Orders API",
		Type:        domain.ServiceTypeAPI,
		Criticality: domain.CriticalityHigh,
	}

	checkout := domain.BusinessJourney{
		ID:          checkoutID,
		WorkspaceID: workspaceID,
		Name:        "Checkout",
		Criticality: domain.CriticalityCritical,
	}

	stub := &impactAnalysisStoreStub{
		services: []domain.Service{
			payments,
			orders,
		},
		journeys: []domain.BusinessJourney{
			checkout,
		},
		healthEvents: map[uuid.UUID]domain.HealthEvent{
			paymentsID: {
				ID:         uuid.New(),
				ServiceID:  paymentsID,
				Status:     domain.HealthStatusDegraded,
				ObservedAt: time.Now(),
				CreatedAt:  time.Now(),
			},
			ordersID: {
				ID:         uuid.New(),
				ServiceID:  ordersID,
				Status:     domain.HealthStatusHealthy,
				ObservedAt: time.Now(),
				CreatedAt:  time.Now(),
			},
		},
		dependencies: map[uuid.UUID][]domain.Service{
			checkoutID: {
				payments,
				orders,
			},
		},
	}

	handler := NewImpactAnalysisHandler(stub)

	router := gin.New()
	router.GET(
		"/workspaces/:workspaceID/impact-analysis",
		handler.Analyze,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/workspaces/"+workspaceID.String()+"/impact-analysis",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusOK,
		)
	}

	var response impactAnalysisResponse

	if err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(response.UnhealthyServices) != 1 {
		t.Fatalf(
			"unhealthy services = %d, want 1",
			len(response.UnhealthyServices),
		)
	}

	if response.UnhealthyServices[0].ID != paymentsID {
		t.Fatalf(
			"unhealthy service = %s, want %s",
			response.UnhealthyServices[0].ID,
			paymentsID,
		)
	}

	if len(response.AffectedJourneys) != 1 {
		t.Fatalf(
			"affected journeys = %d, want 1",
			len(response.AffectedJourneys),
		)
	}

	impact := response.AffectedJourneys[0]

	if impact.JourneyID != checkoutID {
		t.Fatalf(
			"journey id = %s, want %s",
			impact.JourneyID,
			checkoutID,
		)
	}

	if impact.Score != 62 {
		t.Fatalf(
			"score = %d, want 62",
			impact.Score,
		)
	}

	if impact.Severity != domain.ImpactSeverityHigh {
		t.Fatalf(
			"severity = %q, want %q",
			impact.Severity,
			domain.ImpactSeverityHigh,
		)
	}

	if len(impact.Factors) != 3 {
		t.Fatalf(
			"factors = %d, want 3",
			len(impact.Factors),
		)
	}
}

func TestImpactAnalysisHandlerIgnoresUnknownService(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	workspaceID := uuid.New()
	serviceID := uuid.New()
	journeyID := uuid.New()

	service := domain.Service{
		ID:          serviceID,
		WorkspaceID: workspaceID,
		Name:        "Payments API",
		Type:        domain.ServiceTypeAPI,
		Criticality: domain.CriticalityHigh,
	}

	journey := domain.BusinessJourney{
		ID:          journeyID,
		WorkspaceID: workspaceID,
		Name:        "Checkout",
		Criticality: domain.CriticalityCritical,
	}

	stub := &impactAnalysisStoreStub{
		services: []domain.Service{
			service,
		},
		journeys: []domain.BusinessJourney{
			journey,
		},
		healthEvents: map[uuid.UUID]domain.HealthEvent{},
		dependencies: map[uuid.UUID][]domain.Service{
			journeyID: {
				service,
			},
		},
	}

	handler := NewImpactAnalysisHandler(stub)

	router := gin.New()
	router.GET(
		"/workspaces/:workspaceID/impact-analysis",
		handler.Analyze,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/workspaces/"+workspaceID.String()+"/impact-analysis",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusOK,
		)
	}

	var response impactAnalysisResponse

	if err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(response.UnhealthyServices) != 0 {
		t.Fatalf(
			"unhealthy services = %d, want 0",
			len(response.UnhealthyServices),
		)
	}

	if len(response.AffectedJourneys) != 0 {
		t.Fatalf(
			"affected journeys = %d, want 0",
			len(response.AffectedJourneys),
		)
	}
}

func TestImpactAnalysisHandlerReturnsNotFoundForMissingWorkspace(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	workspaceID := uuid.New()

	stub := &impactAnalysisStoreStub{
		listServicesErr: store.ErrNotFound,
	}

	handler := NewImpactAnalysisHandler(stub)

	router := gin.New()
	router.GET(
		"/workspaces/:workspaceID/impact-analysis",
		handler.Analyze,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/workspaces/"+workspaceID.String()+"/impact-analysis",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusNotFound,
		)
	}
}

func TestImpactAnalysisHandlerReturnsInternalErrorForStatusFailure(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	workspaceID := uuid.New()
	serviceID := uuid.New()

	stub := &impactAnalysisStoreWithHealthError{
		service: domain.Service{
			ID:          serviceID,
			WorkspaceID: workspaceID,
			Name:        "Payments API",
			Type:        domain.ServiceTypeAPI,
			Criticality: domain.CriticalityHigh,
		},
	}

	handler := NewImpactAnalysisHandler(stub)

	router := gin.New()
	router.GET(
		"/workspaces/:workspaceID/impact-analysis",
		handler.Analyze,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/workspaces/"+workspaceID.String()+"/impact-analysis",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusInternalServerError,
		)
	}
}

type impactAnalysisStoreWithHealthError struct {
	service domain.Service
}

func (s *impactAnalysisStoreWithHealthError) ListServicesByWorkspace(
	_ context.Context,
	_ uuid.UUID,
) ([]domain.Service, error) {
	return []domain.Service{s.service}, nil
}

func (s *impactAnalysisStoreWithHealthError) GetLatestHealthEvent(
	_ context.Context,
	_ uuid.UUID,
) (domain.HealthEvent, error) {
	return domain.HealthEvent{}, errors.New("database error")
}

func (s *impactAnalysisStoreWithHealthError) ListBusinessJourneysByWorkspace(
	_ context.Context,
	_ uuid.UUID,
) ([]domain.BusinessJourney, error) {
	return nil, nil
}

func (s *impactAnalysisStoreWithHealthError) ListServicesByBusinessJourney(
	_ context.Context,
	_ uuid.UUID,
) ([]domain.Service, error) {
	return nil, nil
}
