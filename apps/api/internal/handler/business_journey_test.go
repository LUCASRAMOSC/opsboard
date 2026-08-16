package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/domain"
	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/store"
)

type businessJourneyStoreStub struct {
	createFunc func(
		context.Context,
		uuid.UUID,
		string,
		domain.Criticality,
	) (domain.BusinessJourney, error)

	getFunc func(
		context.Context,
		uuid.UUID,
	) (domain.BusinessJourney, error)

	listFunc func(
		context.Context,
		uuid.UUID,
	) ([]domain.BusinessJourney, error)

	addServiceFunc func(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) error

	removeServiceFunc func(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) error

	listServicesFunc func(
		context.Context,
		uuid.UUID,
	) ([]domain.Service, error)
}

func (s *businessJourneyStoreStub) CreateBusinessJourney(
	ctx context.Context,
	workspaceID uuid.UUID,
	name string,
	criticality domain.Criticality,
) (domain.BusinessJourney, error) {
	if s.createFunc == nil {
		panic("unexpected CreateBusinessJourney call")
	}

	return s.createFunc(
		ctx,
		workspaceID,
		name,
		criticality,
	)
}

func (s *businessJourneyStoreStub) GetBusinessJourney(
	ctx context.Context,
	id uuid.UUID,
) (domain.BusinessJourney, error) {
	if s.getFunc == nil {
		panic("unexpected GetBusinessJourney call")
	}

	return s.getFunc(ctx, id)
}

func (s *businessJourneyStoreStub) ListBusinessJourneysByWorkspace(
	ctx context.Context,
	workspaceID uuid.UUID,
) ([]domain.BusinessJourney, error) {
	if s.listFunc == nil {
		panic("unexpected ListBusinessJourneysByWorkspace call")
	}

	return s.listFunc(ctx, workspaceID)
}

func (s *businessJourneyStoreStub) AddServiceToBusinessJourney(
	ctx context.Context,
	journeyID uuid.UUID,
	serviceID uuid.UUID,
) error {
	if s.addServiceFunc == nil {
		panic("unexpected AddServiceToBusinessJourney call")
	}

	return s.addServiceFunc(
		ctx,
		journeyID,
		serviceID,
	)
}

func (s *businessJourneyStoreStub) RemoveServiceFromBusinessJourney(
	ctx context.Context,
	journeyID uuid.UUID,
	serviceID uuid.UUID,
) error {
	if s.removeServiceFunc == nil {
		panic("unexpected RemoveServiceFromBusinessJourney call")
	}

	return s.removeServiceFunc(
		ctx,
		journeyID,
		serviceID,
	)
}

func (s *businessJourneyStoreStub) ListServicesByBusinessJourney(
	ctx context.Context,
	journeyID uuid.UUID,
) ([]domain.Service, error) {
	if s.listServicesFunc == nil {
		panic("unexpected ListServicesByBusinessJourney call")
	}

	return s.listServicesFunc(ctx, journeyID)
}

func TestBusinessJourneyHandlerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	workspaceID := uuid.New()
	journeyID := uuid.New()

	dataStore := &businessJourneyStoreStub{
		createFunc: func(
			ctx context.Context,
			gotWorkspaceID uuid.UUID,
			name string,
			criticality domain.Criticality,
		) (domain.BusinessJourney, error) {
			if gotWorkspaceID != workspaceID {
				t.Fatalf(
					"workspace ID = %s, want %s",
					gotWorkspaceID,
					workspaceID,
				)
			}

			if name != "Checkout" {
				t.Fatalf(
					"name = %q, want Checkout",
					name,
				)
			}

			if criticality != domain.CriticalityCritical {
				t.Fatalf(
					"criticality = %q, want %q",
					criticality,
					domain.CriticalityCritical,
				)
			}

			return domain.BusinessJourney{
				ID:          journeyID,
				WorkspaceID: workspaceID,
				Name:        name,
				Criticality: criticality,
			}, nil
		},
	}

	handler := NewBusinessJourneyHandler(dataStore)

	router := gin.New()
	router.POST(
		"/workspaces/:workspaceID/journeys",
		handler.Create,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/workspaces/"+workspaceID.String()+"/journeys",
		strings.NewReader(
			`{
				"name":"  Checkout  ",
				"criticality":"CRITICAL"
			}`,
		),
	)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf(
			"status code = %d, want %d",
			response.Code,
			http.StatusCreated,
		)
	}

	if !strings.Contains(
		response.Body.String(),
		`"name":"Checkout"`,
	) {
		t.Fatalf(
			"response body = %q, want Checkout",
			response.Body.String(),
		)
	}
}

func TestBusinessJourneyHandlerCreateRejectsInvalidCriticality(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	workspaceID := uuid.New()

	handler := NewBusinessJourneyHandler(
		&businessJourneyStoreStub{},
	)

	router := gin.New()
	router.POST(
		"/workspaces/:workspaceID/journeys",
		handler.Create,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/workspaces/"+workspaceID.String()+"/journeys",
		strings.NewReader(
			`{
				"name":"Checkout",
				"criticality":"ABSURDA"
			}`,
		),
	)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"status code = %d, want %d",
			response.Code,
			http.StatusBadRequest,
		)
	}
}

func TestBusinessJourneyHandlerCreateReturnsNotFound(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	workspaceID := uuid.New()

	dataStore := &businessJourneyStoreStub{
		createFunc: func(
			context.Context,
			uuid.UUID,
			string,
			domain.Criticality,
		) (domain.BusinessJourney, error) {
			return domain.BusinessJourney{}, store.ErrNotFound
		},
	}

	handler := NewBusinessJourneyHandler(dataStore)

	router := gin.New()
	router.POST(
		"/workspaces/:workspaceID/journeys",
		handler.Create,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/workspaces/"+workspaceID.String()+"/journeys",
		strings.NewReader(
			`{
				"name":"Checkout",
				"criticality":"CRITICAL"
			}`,
		),
	)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"status code = %d, want %d",
			response.Code,
			http.StatusNotFound,
		)
	}
}

func TestBusinessJourneyHandlerAddService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	journeyID := uuid.New()
	serviceID := uuid.New()

	dataStore := &businessJourneyStoreStub{
		addServiceFunc: func(
			ctx context.Context,
			gotJourneyID uuid.UUID,
			gotServiceID uuid.UUID,
		) error {
			if gotJourneyID != journeyID {
				t.Fatalf(
					"journey ID = %s, want %s",
					gotJourneyID,
					journeyID,
				)
			}

			if gotServiceID != serviceID {
				t.Fatalf(
					"service ID = %s, want %s",
					gotServiceID,
					serviceID,
				)
			}

			return nil
		},
	}

	handler := NewBusinessJourneyHandler(dataStore)

	router := gin.New()
	router.POST(
		"/journeys/:journeyID/services/:serviceID",
		handler.AddService,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/journeys/"+journeyID.String()+
			"/services/"+serviceID.String(),
		nil,
	)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf(
			"status code = %d, want %d",
			response.Code,
			http.StatusNoContent,
		)
	}
}

func TestBusinessJourneyHandlerAddServiceRejectsWorkspaceMismatch(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	journeyID := uuid.New()
	serviceID := uuid.New()

	dataStore := &businessJourneyStoreStub{
		addServiceFunc: func(
			context.Context,
			uuid.UUID,
			uuid.UUID,
		) error {
			return store.ErrWorkspaceMismatch
		},
	}

	handler := NewBusinessJourneyHandler(dataStore)

	router := gin.New()
	router.POST(
		"/journeys/:journeyID/services/:serviceID",
		handler.AddService,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/journeys/"+journeyID.String()+
			"/services/"+serviceID.String(),
		nil,
	)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf(
			"status code = %d, want %d",
			response.Code,
			http.StatusConflict,
		)
	}
}

func TestBusinessJourneyHandlerAddServiceReturnsConflictForDuplicate(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	journeyID := uuid.New()
	serviceID := uuid.New()

	dataStore := &businessJourneyStoreStub{
		addServiceFunc: func(
			context.Context,
			uuid.UUID,
			uuid.UUID,
		) error {
			return store.ErrConflict
		},
	}

	handler := NewBusinessJourneyHandler(dataStore)

	router := gin.New()
	router.POST(
		"/journeys/:journeyID/services/:serviceID",
		handler.AddService,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/journeys/"+journeyID.String()+
			"/services/"+serviceID.String(),
		nil,
	)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf(
			"status code = %d, want %d",
			response.Code,
			http.StatusConflict,
		)
	}
}

func TestBusinessJourneyHandlerListServices(t *testing.T) {
	gin.SetMode(gin.TestMode)

	journeyID := uuid.New()
	serviceID := uuid.New()

	dataStore := &businessJourneyStoreStub{
		listServicesFunc: func(
			ctx context.Context,
			gotJourneyID uuid.UUID,
		) ([]domain.Service, error) {
			if gotJourneyID != journeyID {
				t.Fatalf(
					"journey ID = %s, want %s",
					gotJourneyID,
					journeyID,
				)
			}

			return []domain.Service{
				{
					ID:          serviceID,
					Name:        "Payments API",
					Type:        domain.ServiceTypeAPI,
					Criticality: domain.CriticalityHigh,
				},
			}, nil
		},
	}

	handler := NewBusinessJourneyHandler(dataStore)

	router := gin.New()
	router.GET(
		"/journeys/:journeyID/services",
		handler.ListServices,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/journeys/"+journeyID.String()+"/services",
		nil,
	)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"status code = %d, want %d",
			response.Code,
			http.StatusOK,
		)
	}

	if !strings.Contains(
		response.Body.String(),
		`"name":"Payments API"`,
	) {
		t.Fatalf(
			"response body = %q, want Payments API",
			response.Body.String(),
		)
	}
}

func TestBusinessJourneyHandlerRemoveService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	journeyID := uuid.New()
	serviceID := uuid.New()

	dataStore := &businessJourneyStoreStub{
		removeServiceFunc: func(
			ctx context.Context,
			gotJourneyID uuid.UUID,
			gotServiceID uuid.UUID,
		) error {
			if gotJourneyID != journeyID {
				t.Fatalf(
					"journey ID = %s, want %s",
					gotJourneyID,
					journeyID,
				)
			}

			if gotServiceID != serviceID {
				t.Fatalf(
					"service ID = %s, want %s",
					gotServiceID,
					serviceID,
				)
			}

			return nil
		},
	}

	handler := NewBusinessJourneyHandler(dataStore)

	router := gin.New()
	router.DELETE(
		"/journeys/:journeyID/services/:serviceID",
		handler.RemoveService,
	)

	request := httptest.NewRequest(
		http.MethodDelete,
		"/journeys/"+journeyID.String()+
			"/services/"+serviceID.String(),
		nil,
	)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf(
			"status code = %d, want %d",
			response.Code,
			http.StatusNoContent,
		)
	}
}
