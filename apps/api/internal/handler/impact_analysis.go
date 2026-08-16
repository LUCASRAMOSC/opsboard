package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/domain"
	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/store"
)

type ImpactAnalysisStore interface {
	ListServicesByWorkspace(
		ctx context.Context,
		workspaceID uuid.UUID,
	) ([]domain.Service, error)

	GetLatestHealthEvent(
		ctx context.Context,
		serviceID uuid.UUID,
	) (domain.HealthEvent, error)

	ListBusinessJourneysByWorkspace(
		ctx context.Context,
		workspaceID uuid.UUID,
	) ([]domain.BusinessJourney, error)

	ListServicesByBusinessJourney(
		ctx context.Context,
		businessJourneyID uuid.UUID,
	) ([]domain.Service, error)
}

type ImpactAnalysisHandler struct {
	store ImpactAnalysisStore
}

type impactAnalysisResponse struct {
	WorkspaceID       uuid.UUID                  `json:"workspace_id"`
	UnhealthyServices []unhealthyServiceResponse `json:"unhealthy_services"`
	AffectedJourneys  []journeyImpactResponse    `json:"affected_journeys"`
}

type unhealthyServiceResponse struct {
	ID          uuid.UUID                   `json:"id"`
	Name        string                      `json:"name"`
	Status      domain.CurrentServiceStatus `json:"status"`
	Criticality domain.Criticality          `json:"criticality"`
}

type journeyImpactResponse struct {
	JourneyID            uuid.UUID                       `json:"journey_id"`
	JourneyName          string                          `json:"journey_name"`
	JourneyCriticality   domain.Criticality              `json:"journey_criticality"`
	Score                int                             `json:"score"`
	Severity             domain.ImpactSeverity           `json:"severity"`
	AffectedDependencies int                             `json:"affected_dependencies"`
	TotalDependencies    int                             `json:"total_dependencies"`
	AffectedServices     []affectedServiceImpactResponse `json:"affected_services"`
	Factors              []impactFactorResponse          `json:"factors"`
}

type affectedServiceImpactResponse struct {
	ID                uuid.UUID                   `json:"id"`
	Name              string                      `json:"name"`
	Status            domain.CurrentServiceStatus `json:"status"`
	Criticality       domain.Criticality          `json:"criticality"`
	StatusWeight      int                         `json:"status_weight"`
	CriticalityWeight int                         `json:"criticality_weight"`
	Contribution      int                         `json:"contribution"`
}

type impactFactorResponse struct {
	Type   domain.ImpactFactorType `json:"type"`
	Value  string                  `json:"value"`
	Weight int                     `json:"weight"`
}

func NewImpactAnalysisHandler(
	store ImpactAnalysisStore,
) *ImpactAnalysisHandler {
	return &ImpactAnalysisHandler{
		store: store,
	}
}

func (h *ImpactAnalysisHandler) Analyze(c *gin.Context) {
	workspaceID, err := uuid.Parse(c.Param("workspaceID"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid workspace id"},
		)
		return
	}

	services, err := h.store.ListServicesByWorkspace(
		c.Request.Context(),
		workspaceID,
	)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(
				http.StatusNotFound,
				gin.H{"error": "workspace not found"},
			)
			return
		}

		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "failed to list services"},
		)
		return
	}

	statusByServiceID := make(
		map[uuid.UUID]domain.CurrentServiceStatus,
		len(services),
	)

	unhealthyServices := make(
		[]unhealthyServiceResponse,
		0,
	)

	for _, service := range services {
		status, err := h.currentServiceStatus(
			c.Request.Context(),
			service.ID,
		)
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"error": "failed to resolve service status"},
			)
			return
		}

		statusByServiceID[service.ID] = status

		if !isUnhealthyStatus(status) {
			continue
		}

		unhealthyServices = append(
			unhealthyServices,
			unhealthyServiceResponse{
				ID:          service.ID,
				Name:        service.Name,
				Status:      status,
				Criticality: service.Criticality,
			},
		)
	}

	journeys, err := h.store.ListBusinessJourneysByWorkspace(
		c.Request.Context(),
		workspaceID,
	)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(
				http.StatusNotFound,
				gin.H{"error": "workspace not found"},
			)
			return
		}

		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "failed to list business journeys"},
		)
		return
	}

	affectedJourneys := make(
		[]journeyImpactResponse,
		0,
	)

	for _, journey := range journeys {
		dependencies, err :=
			h.store.ListServicesByBusinessJourney(
				c.Request.Context(),
				journey.ID,
			)
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": "failed to list journey services",
				},
			)
			return
		}

		inputs := make(
			[]domain.ServiceImpactInput,
			0,
			len(dependencies),
		)

		for _, dependency := range dependencies {
			status, ok := statusByServiceID[dependency.ID]
			if !ok {
				c.JSON(
					http.StatusInternalServerError,
					gin.H{
						"error": "journey dependency is outside workspace",
					},
				)
				return
			}

			inputs = append(
				inputs,
				domain.ServiceImpactInput{
					ID:          dependency.ID,
					Name:        dependency.Name,
					Criticality: dependency.Criticality,
					Status:      status,
				},
			)
		}

		impact, affected :=
			domain.AnalyzeJourneyImpact(
				journey,
				inputs,
			)

		if !affected {
			continue
		}

		affectedJourneys = append(
			affectedJourneys,
			newJourneyImpactResponse(impact),
		)
	}

	c.JSON(
		http.StatusOK,
		impactAnalysisResponse{
			WorkspaceID:       workspaceID,
			UnhealthyServices: unhealthyServices,
			AffectedJourneys:  affectedJourneys,
		},
	)
}

func (h *ImpactAnalysisHandler) currentServiceStatus(
	ctx context.Context,
	serviceID uuid.UUID,
) (domain.CurrentServiceStatus, error) {
	event, err := h.store.GetLatestHealthEvent(
		ctx,
		serviceID,
	)
	if err != nil {
		if errors.Is(err, store.ErrNoHealthEvents) {
			return domain.CurrentServiceStatusUnknown, nil
		}

		return domain.CurrentServiceStatusUnknown, err
	}

	return domain.DeriveCurrentServiceStatus(&event), nil
}

func isUnhealthyStatus(
	status domain.CurrentServiceStatus,
) bool {
	return status == domain.CurrentServiceStatusDegraded ||
		status == domain.CurrentServiceStatusUnavailable
}

func newJourneyImpactResponse(
	impact domain.JourneyImpact,
) journeyImpactResponse {
	affectedServices := make(
		[]affectedServiceImpactResponse,
		0,
		len(impact.AffectedServices),
	)

	for _, service := range impact.AffectedServices {
		affectedServices = append(
			affectedServices,
			affectedServiceImpactResponse{
				ID:                service.ID,
				Name:              service.Name,
				Status:            service.Status,
				Criticality:       service.Criticality,
				StatusWeight:      service.StatusWeight,
				CriticalityWeight: service.CriticalityWeight,
				Contribution:      service.Contribution,
			},
		)
	}

	factors := make(
		[]impactFactorResponse,
		0,
		len(impact.Factors),
	)

	for _, factor := range impact.Factors {
		factors = append(
			factors,
			impactFactorResponse{
				Type:   factor.Type,
				Value:  factor.Value,
				Weight: factor.Weight,
			},
		)
	}

	return journeyImpactResponse{
		JourneyID:            impact.JourneyID,
		JourneyName:          impact.JourneyName,
		JourneyCriticality:   impact.JourneyCriticality,
		Score:                impact.Score,
		Severity:             impact.Severity,
		AffectedDependencies: impact.AffectedDependencies,
		TotalDependencies:    impact.TotalDependencies,
		AffectedServices:     affectedServices,
		Factors:              factors,
	}
}
