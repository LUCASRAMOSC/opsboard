package domain

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	degradedStatusWeight          = 30
	unavailableStatusWeight       = 55
	mediumCriticalityWeight       = 5
	highCriticalityWeight         = 10
	criticalCriticalityWeight     = 15
	maxAffectedDependenciesWeight = 15

	mediumImpactThreshold   = 40
	highImpactThreshold     = 60
	criticalImpactThreshold = 80
	maxImpactScore          = 100
)

type ImpactSeverity string

const (
	ImpactSeverityLow      ImpactSeverity = "LOW"
	ImpactSeverityMedium   ImpactSeverity = "MEDIUM"
	ImpactSeverityHigh     ImpactSeverity = "HIGH"
	ImpactSeverityCritical ImpactSeverity = "CRITICAL"
)

type ImpactFactorType string

const (
	ImpactFactorAverageServiceContribution ImpactFactorType = "AVERAGE_SERVICE_CONTRIBUTION"
	ImpactFactorJourneyCriticality         ImpactFactorType = "JOURNEY_CRITICALITY"
	ImpactFactorAffectedDependencies       ImpactFactorType = "AFFECTED_DEPENDENCIES"
)

type ServiceImpactInput struct {
	ID          uuid.UUID
	Name        string
	Criticality Criticality
	Status      CurrentServiceStatus
}

type AffectedServiceImpact struct {
	ID                uuid.UUID
	Name              string
	Status            CurrentServiceStatus
	Criticality       Criticality
	StatusWeight      int
	CriticalityWeight int
	Contribution      int
}

type ImpactFactor struct {
	Type   ImpactFactorType
	Value  string
	Weight int
}

type JourneyImpact struct {
	JourneyID            uuid.UUID
	JourneyName          string
	JourneyCriticality   Criticality
	Score                int
	Severity             ImpactSeverity
	AffectedDependencies int
	TotalDependencies    int
	AffectedServices     []AffectedServiceImpact
	Factors              []ImpactFactor
}

func AnalyzeJourneyImpact(
	journey BusinessJourney,
	dependencies []ServiceImpactInput,
) (JourneyImpact, bool) {
	affectedServices := make(
		[]AffectedServiceImpact,
		0,
		len(dependencies),
	)

	totalServiceContribution := 0

	for _, dependency := range dependencies {
		statusWeight := currentStatusImpactWeight(
			dependency.Status,
		)

		if statusWeight == 0 {
			continue
		}

		criticalityWeight := criticalityImpactWeight(
			dependency.Criticality,
		)

		contribution := statusWeight + criticalityWeight

		totalServiceContribution += contribution

		affectedServices = append(
			affectedServices,
			AffectedServiceImpact{
				ID:                dependency.ID,
				Name:              dependency.Name,
				Status:            dependency.Status,
				Criticality:       dependency.Criticality,
				StatusWeight:      statusWeight,
				CriticalityWeight: criticalityWeight,
				Contribution:      contribution,
			},
		)
	}

	if len(affectedServices) == 0 {
		return JourneyImpact{}, false
	}

	averageServiceContribution :=
		totalServiceContribution / len(affectedServices)

	journeyCriticalityWeight := criticalityImpactWeight(
		journey.Criticality,
	)

	affectedDependenciesWeight := 0

	if len(dependencies) > 0 {
		affectedDependenciesWeight =
			len(affectedServices) *
				maxAffectedDependenciesWeight /
				len(dependencies)
	}

	score := clampImpactScore(
		averageServiceContribution +
			journeyCriticalityWeight +
			affectedDependenciesWeight,
	)

	return JourneyImpact{
		JourneyID:            journey.ID,
		JourneyName:          journey.Name,
		JourneyCriticality:   journey.Criticality,
		Score:                score,
		Severity:             classifyImpactSeverity(score),
		AffectedDependencies: len(affectedServices),
		TotalDependencies:    len(dependencies),
		AffectedServices:     affectedServices,
		Factors: []ImpactFactor{
			{
				Type: ImpactFactorAverageServiceContribution,
				Value: fmt.Sprintf(
					"%d affected service(s)",
					len(affectedServices),
				),
				Weight: averageServiceContribution,
			},
			{
				Type:   ImpactFactorJourneyCriticality,
				Value:  string(journey.Criticality),
				Weight: journeyCriticalityWeight,
			},
			{
				Type: ImpactFactorAffectedDependencies,
				Value: fmt.Sprintf(
					"%d/%d",
					len(affectedServices),
					len(dependencies),
				),
				Weight: affectedDependenciesWeight,
			},
		},
	}, true
}

func currentStatusImpactWeight(
	status CurrentServiceStatus,
) int {
	switch status {
	case CurrentServiceStatusDegraded:
		return degradedStatusWeight

	case CurrentServiceStatusUnavailable:
		return unavailableStatusWeight

	default:
		return 0
	}
}

func criticalityImpactWeight(
	criticality Criticality,
) int {
	switch criticality {
	case CriticalityMedium:
		return mediumCriticalityWeight

	case CriticalityHigh:
		return highCriticalityWeight

	case CriticalityCritical:
		return criticalCriticalityWeight

	default:
		return 0
	}
}

func classifyImpactSeverity(score int) ImpactSeverity {
	switch {
	case score >= criticalImpactThreshold:
		return ImpactSeverityCritical

	case score >= highImpactThreshold:
		return ImpactSeverityHigh

	case score >= mediumImpactThreshold:
		return ImpactSeverityMedium

	default:
		return ImpactSeverityLow
	}
}

func clampImpactScore(score int) int {
	switch {
	case score < 0:
		return 0

	case score > maxImpactScore:
		return maxImpactScore

	default:
		return score
	}
}
