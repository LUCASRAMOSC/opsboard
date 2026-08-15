package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/domain"
	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/store"
)

type ServiceStore interface {
	CreateService(
		ctx context.Context,
		workspaceID uuid.UUID,
		name string,
		serviceType domain.ServiceType,
		criticality domain.Criticality,
	) (domain.Service, error)

	GetService(
		ctx context.Context,
		id uuid.UUID,
	) (domain.Service, error)

	ListServicesByWorkspace(
		ctx context.Context,
		workspaceID uuid.UUID,
	) ([]domain.Service, error)
}

type ServiceHandler struct {
	store ServiceStore
}

func NewServiceHandler(store ServiceStore) *ServiceHandler {
	return &ServiceHandler{
		store: store,
	}
}

type createServiceRequest struct {
	Name        string             `json:"name"`
	Type        domain.ServiceType `json:"type"`
	Criticality domain.Criticality `json:"criticality"`
}

type serviceResponse struct {
	ID          uuid.UUID          `json:"id"`
	WorkspaceID uuid.UUID          `json:"workspace_id"`
	Name        string             `json:"name"`
	Type        domain.ServiceType `json:"type"`
	Criticality domain.Criticality `json:"criticality"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

func (h *ServiceHandler) Create(c *gin.Context) {
	workspaceID, err := uuid.Parse(c.Param("workspaceID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid workspace ID",
		})

		return
	}

	var request createServiceRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})

		return
	}

	name := strings.TrimSpace(request.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "service name is required",
		})

		return
	}

	if !request.Type.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid service type",
		})

		return
	}

	if !request.Criticality.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid service criticality",
		})

		return
	}

	service, err := h.store.CreateService(
		c.Request.Context(),
		workspaceID,
		name,
		request.Type,
		request.Criticality,
	)

	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "workspace not found",
		})

		return
	}

	if errors.Is(err, store.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{
			"error": "service name already exists in workspace",
		})

		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create service",
		})

		return
	}

	c.JSON(
		http.StatusCreated,
		newServiceResponse(service),
	)
}

func (h *ServiceHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("serviceID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid service ID",
		})

		return
	}

	service, err := h.store.GetService(
		c.Request.Context(),
		id,
	)

	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "service not found",
		})

		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get service",
		})

		return
	}

	c.JSON(
		http.StatusOK,
		newServiceResponse(service),
	)
}

func (h *ServiceHandler) ListByWorkspace(c *gin.Context) {
	workspaceID, err := uuid.Parse(c.Param("workspaceID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid workspace ID",
		})

		return
	}

	services, err := h.store.ListServicesByWorkspace(
		c.Request.Context(),
		workspaceID,
	)

	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "workspace not found",
		})

		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list services",
		})

		return
	}

	response := make(
		[]serviceResponse,
		0,
		len(services),
	)

	for _, service := range services {
		response = append(
			response,
			newServiceResponse(service),
		)
	}

	c.JSON(http.StatusOK, response)
}

func newServiceResponse(service domain.Service) serviceResponse {
	return serviceResponse{
		ID:          service.ID,
		WorkspaceID: service.WorkspaceID,
		Name:        service.Name,
		Type:        service.Type,
		Criticality: service.Criticality,
		CreatedAt:   service.CreatedAt,
		UpdatedAt:   service.UpdatedAt,
	}
}
