package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/domain"
	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/store"
)

type HealthEventStore interface {
	CreateHealthEvent(
		ctx context.Context,
		serviceID uuid.UUID,
		status domain.HealthStatus,
		responseTimeMs *int,
		observedAt time.Time,
	) (domain.HealthEvent, error)

	ListHealthEventsByService(
		ctx context.Context,
		serviceID uuid.UUID,
	) ([]domain.HealthEvent, error)

	GetLatestHealthEvent(
		ctx context.Context,
		serviceID uuid.UUID,
	) (domain.HealthEvent, error)
}

type HealthEventHandler struct {
	store HealthEventStore
}

func NewHealthEventHandler(
	store HealthEventStore,
) *HealthEventHandler {
	return &HealthEventHandler{
		store: store,
	}
}

type createHealthEventRequest struct {
	Status         domain.HealthStatus `json:"status"`
	ResponseTimeMs *int                `json:"response_time_ms"`
}

type healthEventResponse struct {
	ID             uuid.UUID           `json:"id"`
	ServiceID      uuid.UUID           `json:"service_id"`
	Status         domain.HealthStatus `json:"status"`
	ResponseTimeMs *int                `json:"response_time_ms"`
	ObservedAt     time.Time           `json:"observed_at"`
	CreatedAt      time.Time           `json:"created_at"`
}

func (h *HealthEventHandler) Create(c *gin.Context) {
	serviceID, err := uuid.Parse(c.Param("serviceID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid service ID",
		})

		return
	}

	var request createHealthEventRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})

		return
	}

	if !request.Status.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid health status",
		})

		return
	}

	if !domain.ValidResponseTimeMs(request.ResponseTimeMs) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid response time",
		})

		return
	}

	healthEvent, err := h.store.CreateHealthEvent(
		c.Request.Context(),
		serviceID,
		request.Status,
		request.ResponseTimeMs,
		time.Now().UTC(),
	)

	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "service not found",
		})

		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create health event",
		})

		return
	}

	c.JSON(
		http.StatusCreated,
		newHealthEventResponse(healthEvent),
	)
}

func (h *HealthEventHandler) ListByService(
	c *gin.Context,
) {
	serviceID, err := uuid.Parse(c.Param("serviceID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid service ID",
		})

		return
	}

	healthEvents, err := h.store.ListHealthEventsByService(
		c.Request.Context(),
		serviceID,
	)

	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "service not found",
		})

		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list health events",
		})

		return
	}

	response := make(
		[]healthEventResponse,
		0,
		len(healthEvents),
	)

	for _, healthEvent := range healthEvents {
		response = append(
			response,
			newHealthEventResponse(healthEvent),
		)
	}

	c.JSON(http.StatusOK, response)
}

func (h *HealthEventHandler) Latest(c *gin.Context) {
	serviceID, err := uuid.Parse(c.Param("serviceID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid service ID",
		})

		return
	}

	healthEvent, err := h.store.GetLatestHealthEvent(
		c.Request.Context(),
		serviceID,
	)

	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "service not found",
		})

		return
	}

	if errors.Is(err, store.ErrNoHealthEvents) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "no health events found for service",
		})

		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get latest health event",
		})

		return
	}

	c.JSON(
		http.StatusOK,
		newHealthEventResponse(healthEvent),
	)
}

func newHealthEventResponse(
	healthEvent domain.HealthEvent,
) healthEventResponse {
	return healthEventResponse{
		ID:             healthEvent.ID,
		ServiceID:      healthEvent.ServiceID,
		Status:         healthEvent.Status,
		ResponseTimeMs: healthEvent.ResponseTimeMs,
		ObservedAt:     healthEvent.ObservedAt,
		CreatedAt:      healthEvent.CreatedAt,
	}
}
