package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/domain"
	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/store"
)

type serviceStoreStub struct {
	createFunc func(
		context.Context,
		uuid.UUID,
		string,
		domain.ServiceType,
		domain.Criticality,
	) (domain.Service, error)

	getFunc func(
		context.Context,
		uuid.UUID,
	) (domain.Service, error)

	listFunc func(
		context.Context,
		uuid.UUID,
	) ([]domain.Service, error)
}

func (s *serviceStoreStub) CreateService(
	ctx context.Context,
	workspaceID uuid.UUID,
	name string,
	serviceType domain.ServiceType,
	criticality domain.Criticality,
) (domain.Service, error) {
	if s.createFunc == nil {
		panic("unexpected CreateService call")
	}

	return s.createFunc(
		ctx,
		workspaceID,
		name,
		serviceType,
		criticality,
	)
}

func (s *serviceStoreStub) GetService(
	ctx context.Context,
	id uuid.UUID,
) (domain.Service, error) {
	if s.getFunc == nil {
		panic("unexpected GetService call")
	}

	return s.getFunc(ctx, id)
}

func (s *serviceStoreStub) ListServicesByWorkspace(
	ctx context.Context,
	workspaceID uuid.UUID,
) ([]domain.Service, error) {
	if s.listFunc == nil {
		panic("unexpected ListServicesByWorkspace call")
	}

	return s.listFunc(ctx, workspaceID)
}

func TestServiceHandlerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	workspaceID := uuid.New()
	serviceID := uuid.New()
	now := time.Now()

	dataStore := &serviceStoreStub{
		createFunc: func(
			ctx context.Context,
			gotWorkspaceID uuid.UUID,
			name string,
			serviceType domain.ServiceType,
			criticality domain.Criticality,
		) (domain.Service, error) {
			if gotWorkspaceID != workspaceID {
				t.Fatalf(
					"workspace ID = %s, want %s",
					gotWorkspaceID,
					workspaceID,
				)
			}

			if name != "Payments API" {
				t.Fatalf(
					"name = %q, want %q",
					name,
					"Payments API",
				)
			}

			return domain.Service{
				ID:          serviceID,
				WorkspaceID: workspaceID,
				Name:        name,
				Type:        serviceType,
				Criticality: criticality,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
	}

	handler := NewServiceHandler(dataStore)

	router := gin.New()
	router.POST(
		"/workspaces/:workspaceID/services",
		handler.Create,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/workspaces/"+workspaceID.String()+"/services",
		strings.NewReader(
			`{
				"name":"  Payments API  ",
				"type":"API",
				"criticality":"HIGH"
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
		`"name":"Payments API"`,
	) {
		t.Errorf(
			"response body = %q, want service name",
			response.Body.String(),
		)
	}
}

func TestServiceHandlerCreateRejectsInvalidWorkspaceUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewServiceHandler(&serviceStoreStub{})

	router := gin.New()
	router.POST(
		"/workspaces/:workspaceID/services",
		handler.Create,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/workspaces/abc/services",
		strings.NewReader(
			`{
				"name":"Payments API",
				"type":"API",
				"criticality":"HIGH"
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

func TestServiceHandlerCreateRejectsInvalidType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	workspaceID := uuid.New()

	handler := NewServiceHandler(&serviceStoreStub{})

	router := gin.New()
	router.POST(
		"/workspaces/:workspaceID/services",
		handler.Create,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/workspaces/"+workspaceID.String()+"/services",
		strings.NewReader(
			`{
				"name":"Payments API",
				"type":"INVALID",
				"criticality":"HIGH"
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

func TestServiceHandlerCreateRejectsInvalidCriticality(t *testing.T) {
	gin.SetMode(gin.TestMode)

	workspaceID := uuid.New()

	handler := NewServiceHandler(&serviceStoreStub{})

	router := gin.New()
	router.POST(
		"/workspaces/:workspaceID/services",
		handler.Create,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/workspaces/"+workspaceID.String()+"/services",
		strings.NewReader(
			`{
				"name":"Payments API",
				"type":"API",
				"criticality":"INVALID"
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

func TestServiceHandlerCreateReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	workspaceID := uuid.New()

	dataStore := &serviceStoreStub{
		createFunc: func(
			ctx context.Context,
			workspaceID uuid.UUID,
			name string,
			serviceType domain.ServiceType,
			criticality domain.Criticality,
		) (domain.Service, error) {
			return domain.Service{}, store.ErrNotFound
		},
	}

	handler := NewServiceHandler(dataStore)

	router := gin.New()
	router.POST(
		"/workspaces/:workspaceID/services",
		handler.Create,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/workspaces/"+workspaceID.String()+"/services",
		strings.NewReader(
			`{
				"name":"Payments API",
				"type":"API",
				"criticality":"HIGH"
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

func TestServiceHandlerCreateReturnsConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)

	workspaceID := uuid.New()

	dataStore := &serviceStoreStub{
		createFunc: func(
			ctx context.Context,
			workspaceID uuid.UUID,
			name string,
			serviceType domain.ServiceType,
			criticality domain.Criticality,
		) (domain.Service, error) {
			return domain.Service{}, store.ErrConflict
		},
	}

	handler := NewServiceHandler(dataStore)

	router := gin.New()
	router.POST(
		"/workspaces/:workspaceID/services",
		handler.Create,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/workspaces/"+workspaceID.String()+"/services",
		strings.NewReader(
			`{
				"name":"Payments API",
				"type":"API",
				"criticality":"HIGH"
			}`,
		),
	)
	request.Header.Set("Content-Type", "application/json")

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

func TestServiceHandlerGetReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceID := uuid.New()

	dataStore := &serviceStoreStub{
		getFunc: func(
			ctx context.Context,
			id uuid.UUID,
		) (domain.Service, error) {
			return domain.Service{}, store.ErrNotFound
		},
	}

	handler := NewServiceHandler(dataStore)

	router := gin.New()
	router.GET("/services/:serviceID", handler.Get)

	request := httptest.NewRequest(
		http.MethodGet,
		"/services/"+serviceID.String(),
		nil,
	)

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

func TestServiceHandlerListByWorkspace(t *testing.T) {
	gin.SetMode(gin.TestMode)

	workspaceID := uuid.New()
	serviceID := uuid.New()

	dataStore := &serviceStoreStub{
		listFunc: func(
			ctx context.Context,
			gotWorkspaceID uuid.UUID,
		) ([]domain.Service, error) {
			if gotWorkspaceID != workspaceID {
				t.Fatalf(
					"workspace ID = %s, want %s",
					gotWorkspaceID,
					workspaceID,
				)
			}

			return []domain.Service{
				{
					ID:          serviceID,
					WorkspaceID: workspaceID,
					Name:        "Payments API",
					Type:        domain.ServiceTypeAPI,
					Criticality: domain.CriticalityHigh,
				},
			}, nil
		},
	}

	handler := NewServiceHandler(dataStore)

	router := gin.New()
	router.GET(
		"/workspaces/:workspaceID/services",
		handler.ListByWorkspace,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/workspaces/"+workspaceID.String()+"/services",
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
		t.Errorf(
			"response body = %q, want service",
			response.Body.String(),
		)
	}
}

func TestServiceHandlerListByWorkspaceReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	workspaceID := uuid.New()

	dataStore := &serviceStoreStub{
		listFunc: func(
			ctx context.Context,
			workspaceID uuid.UUID,
		) ([]domain.Service, error) {
			return nil, store.ErrNotFound
		},
	}

	handler := NewServiceHandler(dataStore)

	router := gin.New()
	router.GET(
		"/workspaces/:workspaceID/services",
		handler.ListByWorkspace,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/workspaces/"+workspaceID.String()+"/services",
		nil,
	)

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

func TestServiceHandlerCreateRejectsBlankName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	workspaceID := uuid.New()

	handler := NewServiceHandler(&serviceStoreStub{})

	router := gin.New()
	router.POST(
		"/workspaces/:workspaceID/services",
		handler.Create,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/workspaces/"+workspaceID.String()+"/services",
		strings.NewReader(
			`{
				"name":"   ",
				"type":"API",
				"criticality":"HIGH"
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

func TestServiceHandlerGetRejectsInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewServiceHandler(&serviceStoreStub{})

	router := gin.New()
	router.GET("/services/:serviceID", handler.Get)

	request := httptest.NewRequest(
		http.MethodGet,
		"/services/abc",
		nil,
	)

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
