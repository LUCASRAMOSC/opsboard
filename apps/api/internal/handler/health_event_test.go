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

type healthEventStoreStub struct {
	createFunc func(
		context.Context,
		uuid.UUID,
		domain.HealthStatus,
		*int,
		time.Time,
	) (domain.HealthEvent, error)

	listFunc func(
		context.Context,
		uuid.UUID,
	) ([]domain.HealthEvent, error)

	latestFunc func(
		context.Context,
		uuid.UUID,
	) (domain.HealthEvent, error)
}

func (s *healthEventStoreStub) CreateHealthEvent(
	ctx context.Context,
	serviceID uuid.UUID,
	status domain.HealthStatus,
	responseTimeMs *int,
	observedAt time.Time,
) (domain.HealthEvent, error) {
	if s.createFunc == nil {
		panic("unexpected CreateHealthEvent call")
	}

	return s.createFunc(
		ctx,
		serviceID,
		status,
		responseTimeMs,
		observedAt,
	)
}

func (s *healthEventStoreStub) ListHealthEventsByService(
	ctx context.Context,
	serviceID uuid.UUID,
) ([]domain.HealthEvent, error) {
	if s.listFunc == nil {
		panic("unexpected ListHealthEventsByService call")
	}

	return s.listFunc(ctx, serviceID)
}

func (s *healthEventStoreStub) GetLatestHealthEvent(
	ctx context.Context,
	serviceID uuid.UUID,
) (domain.HealthEvent, error) {
	if s.latestFunc == nil {
		panic("unexpected GetLatestHealthEvent call")
	}

	return s.latestFunc(ctx, serviceID)
}

func TestHealthEventHandlerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceID := uuid.New()
	healthEventID := uuid.New()
	responseTimeMs := 1250

	dataStore := &healthEventStoreStub{
		createFunc: func(
			ctx context.Context,
			gotServiceID uuid.UUID,
			status domain.HealthStatus,
			gotResponseTimeMs *int,
			observedAt time.Time,
		) (domain.HealthEvent, error) {
			if gotServiceID != serviceID {
				t.Fatalf(
					"service ID = %s, want %s",
					gotServiceID,
					serviceID,
				)
			}

			if status != domain.HealthStatusDegraded {
				t.Fatalf(
					"status = %q, want %q",
					status,
					domain.HealthStatusDegraded,
				)
			}

			if gotResponseTimeMs == nil {
				t.Fatal("response time is nil")
			}

			if *gotResponseTimeMs != responseTimeMs {
				t.Fatalf(
					"response time = %d, want %d",
					*gotResponseTimeMs,
					responseTimeMs,
				)
			}

			if observedAt.IsZero() {
				t.Fatal("observed at is zero")
			}

			return domain.HealthEvent{
				ID:             healthEventID,
				ServiceID:      serviceID,
				Status:         status,
				ResponseTimeMs: gotResponseTimeMs,
				ObservedAt:     observedAt,
				CreatedAt:      observedAt,
			}, nil
		},
	}

	handler := NewHealthEventHandler(dataStore)

	router := gin.New()
	router.POST(
		"/services/:serviceID/health-events",
		handler.Create,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/services/"+serviceID.String()+"/health-events",
		strings.NewReader(
			`{
				"status":"DEGRADED",
				"response_time_ms":1250
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
		`"status":"DEGRADED"`,
	) {
		t.Errorf(
			"response body = %q, want DEGRADED status",
			response.Body.String(),
		)
	}

	if !strings.Contains(
		response.Body.String(),
		`"response_time_ms":1250`,
	) {
		t.Errorf(
			"response body = %q, want response time",
			response.Body.String(),
		)
	}
}

func TestHealthEventHandlerCreateRejectsInvalidServiceUUID(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	handler := NewHealthEventHandler(
		&healthEventStoreStub{},
	)

	router := gin.New()
	router.POST(
		"/services/:serviceID/health-events",
		handler.Create,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/services/abc/health-events",
		strings.NewReader(
			`{
				"status":"HEALTHY"
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

func TestHealthEventHandlerCreateRejectsInvalidStatus(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	serviceID := uuid.New()

	handler := NewHealthEventHandler(
		&healthEventStoreStub{},
	)

	router := gin.New()
	router.POST(
		"/services/:serviceID/health-events",
		handler.Create,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/services/"+serviceID.String()+"/health-events",
		strings.NewReader(
			`{
				"status":"BROKEN"
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

func TestHealthEventHandlerCreateRejectsNegativeResponseTime(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	serviceID := uuid.New()

	handler := NewHealthEventHandler(
		&healthEventStoreStub{},
	)

	router := gin.New()
	router.POST(
		"/services/:serviceID/health-events",
		handler.Create,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/services/"+serviceID.String()+"/health-events",
		strings.NewReader(
			`{
				"status":"DEGRADED",
				"response_time_ms":-1
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

func TestHealthEventHandlerCreateReturnsNotFound(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	serviceID := uuid.New()

	dataStore := &healthEventStoreStub{
		createFunc: func(
			ctx context.Context,
			serviceID uuid.UUID,
			status domain.HealthStatus,
			responseTimeMs *int,
			observedAt time.Time,
		) (domain.HealthEvent, error) {
			return domain.HealthEvent{}, store.ErrNotFound
		},
	}

	handler := NewHealthEventHandler(dataStore)

	router := gin.New()
	router.POST(
		"/services/:serviceID/health-events",
		handler.Create,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/services/"+serviceID.String()+"/health-events",
		strings.NewReader(
			`{
				"status":"HEALTHY"
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

func TestHealthEventHandlerListByService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceID := uuid.New()

	dataStore := &healthEventStoreStub{
		listFunc: func(
			ctx context.Context,
			gotServiceID uuid.UUID,
		) ([]domain.HealthEvent, error) {
			if gotServiceID != serviceID {
				t.Fatalf(
					"service ID = %s, want %s",
					gotServiceID,
					serviceID,
				)
			}

			return []domain.HealthEvent{
				{
					ID:        uuid.New(),
					ServiceID: serviceID,
					Status:    domain.HealthStatusHealthy,
				},
				{
					ID:        uuid.New(),
					ServiceID: serviceID,
					Status:    domain.HealthStatusDegraded,
				},
			}, nil
		},
	}

	handler := NewHealthEventHandler(dataStore)

	router := gin.New()
	router.GET(
		"/services/:serviceID/health-events",
		handler.ListByService,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/services/"+serviceID.String()+"/health-events",
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

	body := response.Body.String()

	if !strings.Contains(body, `"status":"HEALTHY"`) {
		t.Errorf(
			"response body = %q, want HEALTHY event",
			body,
		)
	}

	if !strings.Contains(body, `"status":"DEGRADED"`) {
		t.Errorf(
			"response body = %q, want DEGRADED event",
			body,
		)
	}
}

func TestHealthEventHandlerListByServiceReturnsEmptyHistory(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	serviceID := uuid.New()

	dataStore := &healthEventStoreStub{
		listFunc: func(
			ctx context.Context,
			serviceID uuid.UUID,
		) ([]domain.HealthEvent, error) {
			return []domain.HealthEvent{}, nil
		},
	}

	handler := NewHealthEventHandler(dataStore)

	router := gin.New()
	router.GET(
		"/services/:serviceID/health-events",
		handler.ListByService,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/services/"+serviceID.String()+"/health-events",
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

	if strings.TrimSpace(response.Body.String()) != "[]" {
		t.Fatalf(
			"response body = %q, want []",
			response.Body.String(),
		)
	}
}

func TestHealthEventHandlerListByServiceReturnsNotFound(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	serviceID := uuid.New()

	dataStore := &healthEventStoreStub{
		listFunc: func(
			ctx context.Context,
			serviceID uuid.UUID,
		) ([]domain.HealthEvent, error) {
			return nil, store.ErrNotFound
		},
	}

	handler := NewHealthEventHandler(dataStore)

	router := gin.New()
	router.GET(
		"/services/:serviceID/health-events",
		handler.ListByService,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/services/"+serviceID.String()+"/health-events",
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

func TestHealthEventHandlerLatest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceID := uuid.New()

	dataStore := &healthEventStoreStub{
		latestFunc: func(
			ctx context.Context,
			gotServiceID uuid.UUID,
		) (domain.HealthEvent, error) {
			if gotServiceID != serviceID {
				t.Fatalf(
					"service ID = %s, want %s",
					gotServiceID,
					serviceID,
				)
			}

			return domain.HealthEvent{
				ID:        uuid.New(),
				ServiceID: serviceID,
				Status:    domain.HealthStatusUnavailable,
			}, nil
		},
	}

	handler := NewHealthEventHandler(dataStore)

	router := gin.New()
	router.GET(
		"/services/:serviceID/health-events/latest",
		handler.Latest,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/services/"+serviceID.String()+"/health-events/latest",
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
		`"status":"UNAVAILABLE"`,
	) {
		t.Errorf(
			"response body = %q, want UNAVAILABLE status",
			response.Body.String(),
		)
	}
}

func TestHealthEventHandlerLatestReturnsNoHealthEvents(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	serviceID := uuid.New()

	dataStore := &healthEventStoreStub{
		latestFunc: func(
			ctx context.Context,
			serviceID uuid.UUID,
		) (domain.HealthEvent, error) {
			return domain.HealthEvent{}, store.ErrNoHealthEvents
		},
	}

	handler := NewHealthEventHandler(dataStore)

	router := gin.New()
	router.GET(
		"/services/:serviceID/health-events/latest",
		handler.Latest,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/services/"+serviceID.String()+"/health-events/latest",
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

	if !strings.Contains(
		response.Body.String(),
		`"error":"no health events found for service"`,
	) {
		t.Errorf(
			"response body = %q, want no health events error",
			response.Body.String(),
		)
	}
}
