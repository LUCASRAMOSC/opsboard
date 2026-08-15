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

type workspaceStoreStub struct {
	createFunc func(
		context.Context,
		string,
	) (domain.Workspace, error)

	getFunc func(
		context.Context,
		uuid.UUID,
	) (domain.Workspace, error)

	listFunc func(
		context.Context,
	) ([]domain.Workspace, error)
}

func (s *workspaceStoreStub) CreateWorkspace(
	ctx context.Context,
	name string,
) (domain.Workspace, error) {
	if s.createFunc == nil {
		panic("unexpected CreateWorkspace call")
	}

	return s.createFunc(ctx, name)
}

func (s *workspaceStoreStub) GetWorkspace(
	ctx context.Context,
	id uuid.UUID,
) (domain.Workspace, error) {
	if s.getFunc == nil {
		panic("unexpected GetWorkspace call")
	}

	return s.getFunc(ctx, id)
}

func (s *workspaceStoreStub) ListWorkspaces(
	ctx context.Context,
) ([]domain.Workspace, error) {
	if s.listFunc == nil {
		panic("unexpected ListWorkspaces call")
	}

	return s.listFunc(ctx)
}

func TestWorkspaceHandlerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()
	now := time.Now()

	dataStore := &workspaceStoreStub{
		createFunc: func(
			ctx context.Context,
			name string,
		) (domain.Workspace, error) {
			if name != "ShopFlow" {
				t.Fatalf("name = %q, want %q", name, "ShopFlow")
			}

			return domain.Workspace{
				ID:        id,
				Name:      name,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}

	handler := NewWorkspaceHandler(dataStore)

	router := gin.New()
	router.POST("/workspaces", handler.Create)

	request := httptest.NewRequest(
		http.MethodPost,
		"/workspaces",
		strings.NewReader(`{"name":"  ShopFlow  "}`),
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

	if !strings.Contains(response.Body.String(), `"name":"ShopFlow"`) {
		t.Errorf(
			"response body = %q, want workspace name",
			response.Body.String(),
		)
	}
}

func TestWorkspaceHandlerCreateRejectsBlankName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewWorkspaceHandler(&workspaceStoreStub{})

	router := gin.New()
	router.POST("/workspaces", handler.Create)

	request := httptest.NewRequest(
		http.MethodPost,
		"/workspaces",
		strings.NewReader(`{"name":"   "}`),
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

func TestWorkspaceHandlerGetRejectsInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewWorkspaceHandler(&workspaceStoreStub{})

	router := gin.New()
	router.GET("/workspaces/:workspaceID", handler.Get)

	request := httptest.NewRequest(
		http.MethodGet,
		"/workspaces/abc",
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

func TestWorkspaceHandlerGetReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()

	dataStore := &workspaceStoreStub{
		getFunc: func(
			ctx context.Context,
			workspaceID uuid.UUID,
		) (domain.Workspace, error) {
			return domain.Workspace{}, store.ErrNotFound
		},
	}

	handler := NewWorkspaceHandler(dataStore)

	router := gin.New()
	router.GET("/workspaces/:workspaceID", handler.Get)

	request := httptest.NewRequest(
		http.MethodGet,
		"/workspaces/"+id.String(),
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

func TestWorkspaceHandlerList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dataStore := &workspaceStoreStub{
		listFunc: func(
			ctx context.Context,
		) ([]domain.Workspace, error) {
			return []domain.Workspace{
				{
					ID:   uuid.New(),
					Name: "ShopFlow",
				},
			}, nil
		},
	}

	handler := NewWorkspaceHandler(dataStore)

	router := gin.New()
	router.GET("/workspaces", handler.List)

	request := httptest.NewRequest(
		http.MethodGet,
		"/workspaces",
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

	if !strings.Contains(response.Body.String(), `"name":"ShopFlow"`) {
		t.Errorf(
			"response body = %q, want workspace",
			response.Body.String(),
		)
	}
}
