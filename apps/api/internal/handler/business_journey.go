package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/domain"
	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/store"
)

type BusinessJourneyStore interface {
	CreateBusinessJourney(
		ctx context.Context,
		workspaceID uuid.UUID,
		name string,
		criticality domain.Criticality,
	) (domain.BusinessJourney, error)

	GetBusinessJourney(
		ctx context.Context,
		id uuid.UUID,
	) (domain.BusinessJourney, error)

	ListBusinessJourneysByWorkspace(
		ctx context.Context,
		workspaceID uuid.UUID,
	) ([]domain.BusinessJourney, error)

	AddServiceToBusinessJourney(
		ctx context.Context,
		journeyID uuid.UUID,
		serviceID uuid.UUID,
	) error

	RemoveServiceFromBusinessJourney(
		ctx context.Context,
		journeyID uuid.UUID,
		serviceID uuid.UUID,
	) error

	ListServicesByBusinessJourney(
		ctx context.Context,
		journeyID uuid.UUID,
	) ([]domain.Service, error)
}

type BusinessJourneyHandler struct {
	store BusinessJourneyStore
}

func NewBusinessJourneyHandler(
	store BusinessJourneyStore,
) *BusinessJourneyHandler {
	return &BusinessJourneyHandler{
		store: store,
	}
}

type createBusinessJourneyRequest struct {
	Name        string             `json:"name"`
	Criticality domain.Criticality `json:"criticality"`
}

type businessJourneyResponse struct {
	ID          uuid.UUID          `json:"id"`
	WorkspaceID uuid.UUID          `json:"workspace_id"`
	Name        string             `json:"name"`
	Criticality domain.Criticality `json:"criticality"`
}

func (h *BusinessJourneyHandler) Create(c *gin.Context) {
	workspaceID, err := uuid.Parse(c.Param("workspaceID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid workspace ID",
		})

		return
	}

	var request createBusinessJourneyRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})

		return
	}

	name := strings.TrimSpace(request.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "business journey name is required",
		})

		return
	}

	if !request.Criticality.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid business journey criticality",
		})

		return
	}

	journey, err := h.store.CreateBusinessJourney(
		c.Request.Context(),
		workspaceID,
		name,
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
			"error": "business journey name already exists in workspace",
		})

		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create business journey",
		})

		return
	}

	c.JSON(
		http.StatusCreated,
		newBusinessJourneyResponse(journey),
	)
}

func (h *BusinessJourneyHandler) Get(c *gin.Context) {
	journeyID, err := uuid.Parse(c.Param("journeyID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid business journey ID",
		})

		return
	}

	journey, err := h.store.GetBusinessJourney(
		c.Request.Context(),
		journeyID,
	)

	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "business journey not found",
		})

		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get business journey",
		})

		return
	}

	c.JSON(
		http.StatusOK,
		newBusinessJourneyResponse(journey),
	)
}

func (h *BusinessJourneyHandler) ListByWorkspace(
	c *gin.Context,
) {
	workspaceID, err := uuid.Parse(c.Param("workspaceID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid workspace ID",
		})

		return
	}

	journeys, err := h.store.ListBusinessJourneysByWorkspace(
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
			"error": "failed to list business journeys",
		})

		return
	}

	response := make(
		[]businessJourneyResponse,
		0,
		len(journeys),
	)

	for _, journey := range journeys {
		response = append(
			response,
			newBusinessJourneyResponse(journey),
		)
	}

	c.JSON(http.StatusOK, response)
}

func (h *BusinessJourneyHandler) AddService(
	c *gin.Context,
) {
	journeyID, err := uuid.Parse(c.Param("journeyID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid business journey ID",
		})

		return
	}

	serviceID, err := uuid.Parse(c.Param("serviceID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid service ID",
		})

		return
	}

	err = h.store.AddServiceToBusinessJourney(
		c.Request.Context(),
		journeyID,
		serviceID,
	)

	if errors.Is(err, store.ErrWorkspaceMismatch) {
		c.JSON(http.StatusConflict, gin.H{
			"error": "service and business journey must belong to the same workspace",
		})

		return
	}

	if errors.Is(err, store.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{
			"error": "service already associated with business journey",
		})

		return
	}

	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "business journey or service not found",
		})

		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to associate service with business journey",
		})

		return
	}

	c.Status(http.StatusNoContent)
}

func (h *BusinessJourneyHandler) RemoveService(
	c *gin.Context,
) {
	journeyID, err := uuid.Parse(c.Param("journeyID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid business journey ID",
		})

		return
	}

	serviceID, err := uuid.Parse(c.Param("serviceID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid service ID",
		})

		return
	}

	err = h.store.RemoveServiceFromBusinessJourney(
		c.Request.Context(),
		journeyID,
		serviceID,
	)

	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "business journey service association not found",
		})

		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to remove service from business journey",
		})

		return
	}

	c.Status(http.StatusNoContent)
}

func (h *BusinessJourneyHandler) ListServices(
	c *gin.Context,
) {
	journeyID, err := uuid.Parse(c.Param("journeyID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid business journey ID",
		})

		return
	}

	services, err := h.store.ListServicesByBusinessJourney(
		c.Request.Context(),
		journeyID,
	)

	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "business journey not found",
		})

		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list business journey services",
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

func newBusinessJourneyResponse(
	journey domain.BusinessJourney,
) businessJourneyResponse {
	return businessJourneyResponse{
		ID:          journey.ID,
		WorkspaceID: journey.WorkspaceID,
		Name:        journey.Name,
		Criticality: journey.Criticality,
	}
}
