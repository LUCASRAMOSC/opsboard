package domain

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestAnalyzeJourneyImpactCheckoutDegradedExample(
	t *testing.T,
) {
	journey := BusinessJourney{
		ID:          uuid.New(),
		Name:        "Checkout",
		Criticality: CriticalityCritical,
	}

	dependencies := []ServiceImpactInput{
		{
			ID:          uuid.New(),
			Name:        "Payments API",
			Criticality: CriticalityHigh,
			Status:      CurrentServiceStatusDegraded,
		},
		{
			ID:          uuid.New(),
			Name:        "Orders API",
			Criticality: CriticalityHigh,
			Status:      CurrentServiceStatusHealthy,
		},
	}

	impact, affected := AnalyzeJourneyImpact(
		journey,
		dependencies,
	)

	if !affected {
		t.Fatal("journey is not affected")
	}

	if impact.Score != 62 {
		t.Fatalf(
			"score = %d, want 62",
			impact.Score,
		)
	}

	if impact.Severity != ImpactSeverityHigh {
		t.Fatalf(
			"severity = %q, want %q",
			impact.Severity,
			ImpactSeverityHigh,
		)
	}

	if impact.AffectedDependencies != 1 {
		t.Fatalf(
			"affected dependencies = %d, want 1",
			impact.AffectedDependencies,
		)
	}

	if impact.TotalDependencies != 2 {
		t.Fatalf(
			"total dependencies = %d, want 2",
			impact.TotalDependencies,
		)
	}
}

func TestAnalyzeJourneyImpactIgnoresHealthyAndUnknownServices(
	t *testing.T,
) {
	journey := BusinessJourney{
		ID:          uuid.New(),
		Name:        "Checkout",
		Criticality: CriticalityCritical,
	}

	dependencies := []ServiceImpactInput{
		{
			ID:          uuid.New(),
			Name:        "Payments API",
			Criticality: CriticalityHigh,
			Status:      CurrentServiceStatusHealthy,
		},
		{
			ID:          uuid.New(),
			Name:        "Orders API",
			Criticality: CriticalityHigh,
			Status:      CurrentServiceStatusUnknown,
		},
	}

	impact, affected := AnalyzeJourneyImpact(
		journey,
		dependencies,
	)

	if affected {
		t.Fatalf(
			"journey unexpectedly affected: %+v",
			impact,
		)
	}
}

func TestUnavailableServiceProducesHigherImpactThanDegraded(
	t *testing.T,
) {
	journey := BusinessJourney{
		ID:          uuid.New(),
		Name:        "Checkout",
		Criticality: CriticalityCritical,
	}

	serviceID := uuid.New()

	degraded, _ := AnalyzeJourneyImpact(
		journey,
		[]ServiceImpactInput{
			{
				ID:          serviceID,
				Name:        "Payments API",
				Criticality: CriticalityHigh,
				Status:      CurrentServiceStatusDegraded,
			},
		},
	)

	unavailable, _ := AnalyzeJourneyImpact(
		journey,
		[]ServiceImpactInput{
			{
				ID:          serviceID,
				Name:        "Payments API",
				Criticality: CriticalityHigh,
				Status:      CurrentServiceStatusUnavailable,
			},
		},
	)

	if unavailable.Score <= degraded.Score {
		t.Fatalf(
			"unavailable score = %d, degraded score = %d",
			unavailable.Score,
			degraded.Score,
		)
	}
}

func TestServiceCriticalityInfluencesImpact(t *testing.T) {
	journey := BusinessJourney{
		ID:          uuid.New(),
		Name:        "Checkout",
		Criticality: CriticalityMedium,
	}

	low, _ := AnalyzeJourneyImpact(
		journey,
		[]ServiceImpactInput{
			{
				ID:          uuid.New(),
				Name:        "Payments API",
				Criticality: CriticalityLow,
				Status:      CurrentServiceStatusDegraded,
			},
		},
	)

	critical, _ := AnalyzeJourneyImpact(
		journey,
		[]ServiceImpactInput{
			{
				ID:          uuid.New(),
				Name:        "Payments API",
				Criticality: CriticalityCritical,
				Status:      CurrentServiceStatusDegraded,
			},
		},
	)

	if critical.Score <= low.Score {
		t.Fatalf(
			"critical service score = %d, low service score = %d",
			critical.Score,
			low.Score,
		)
	}
}

func TestJourneyCriticalityInfluencesImpact(t *testing.T) {
	dependencies := []ServiceImpactInput{
		{
			ID:          uuid.New(),
			Name:        "Payments API",
			Criticality: CriticalityHigh,
			Status:      CurrentServiceStatusDegraded,
		},
	}

	low, _ := AnalyzeJourneyImpact(
		BusinessJourney{
			ID:          uuid.New(),
			Name:        "Low Journey",
			Criticality: CriticalityLow,
		},
		dependencies,
	)

	critical, _ := AnalyzeJourneyImpact(
		BusinessJourney{
			ID:          uuid.New(),
			Name:        "Critical Journey",
			Criticality: CriticalityCritical,
		},
		dependencies,
	)

	if critical.Score <= low.Score {
		t.Fatalf(
			"critical journey score = %d, low journey score = %d",
			critical.Score,
			low.Score,
		)
	}
}

func TestAffectedDependencyCountInfluencesImpact(
	t *testing.T,
) {
	journey := BusinessJourney{
		ID:          uuid.New(),
		Name:        "Checkout",
		Criticality: CriticalityMedium,
	}

	oneAffected, _ := AnalyzeJourneyImpact(
		journey,
		[]ServiceImpactInput{
			{
				ID:          uuid.New(),
				Name:        "Payments API",
				Criticality: CriticalityHigh,
				Status:      CurrentServiceStatusDegraded,
			},
			{
				ID:          uuid.New(),
				Name:        "Orders API",
				Criticality: CriticalityHigh,
				Status:      CurrentServiceStatusHealthy,
			},
		},
	)

	twoAffected, _ := AnalyzeJourneyImpact(
		journey,
		[]ServiceImpactInput{
			{
				ID:          uuid.New(),
				Name:        "Payments API",
				Criticality: CriticalityHigh,
				Status:      CurrentServiceStatusDegraded,
			},
			{
				ID:          uuid.New(),
				Name:        "Orders API",
				Criticality: CriticalityHigh,
				Status:      CurrentServiceStatusDegraded,
			},
		},
	)

	if twoAffected.Score <= oneAffected.Score {
		t.Fatalf(
			"two affected score = %d, one affected score = %d",
			twoAffected.Score,
			oneAffected.Score,
		)
	}
}

func TestImpactScoreDoesNotExceedMaximum(t *testing.T) {
	impact, affected := AnalyzeJourneyImpact(
		BusinessJourney{
			ID:          uuid.New(),
			Name:        "Checkout",
			Criticality: CriticalityCritical,
		},
		[]ServiceImpactInput{
			{
				ID:          uuid.New(),
				Name:        "Payments API",
				Criticality: CriticalityCritical,
				Status:      CurrentServiceStatusUnavailable,
			},
		},
	)

	if !affected {
		t.Fatal("journey is not affected")
	}

	if impact.Score != 100 {
		t.Fatalf(
			"score = %d, want 100",
			impact.Score,
		)
	}

	if impact.Severity != ImpactSeverityCritical {
		t.Fatalf(
			"severity = %q, want %q",
			impact.Severity,
			ImpactSeverityCritical,
		)
	}
}

func TestImpactFactorsExplainScore(t *testing.T) {
	impact, affected := AnalyzeJourneyImpact(
		BusinessJourney{
			ID:          uuid.New(),
			Name:        "Checkout",
			Criticality: CriticalityCritical,
		},
		[]ServiceImpactInput{
			{
				ID:          uuid.New(),
				Name:        "Payments API",
				Criticality: CriticalityHigh,
				Status:      CurrentServiceStatusDegraded,
			},
			{
				ID:          uuid.New(),
				Name:        "Orders API",
				Criticality: CriticalityHigh,
				Status:      CurrentServiceStatusHealthy,
			},
		},
	)

	if !affected {
		t.Fatal("journey is not affected")
	}

	totalFactorWeight := 0

	for _, factor := range impact.Factors {
		totalFactorWeight += factor.Weight
	}

	if totalFactorWeight != impact.Score {
		t.Fatalf(
			"factor weights = %d, score = %d",
			totalFactorWeight,
			impact.Score,
		)
	}

	if len(impact.AffectedServices) != 1 {
		t.Fatalf(
			"affected services = %d, want 1",
			len(impact.AffectedServices),
		)
	}

	service := impact.AffectedServices[0]

	if service.Contribution !=
		service.StatusWeight+service.CriticalityWeight {
		t.Fatalf(
			"service contribution = %d, status + criticality = %d",
			service.Contribution,
			service.StatusWeight+
				service.CriticalityWeight,
		)
	}
}

func TestImpactAnalysisIsDeterministic(t *testing.T) {
	journey := BusinessJourney{
		ID:          uuid.New(),
		Name:        "Checkout",
		Criticality: CriticalityCritical,
	}

	dependencies := []ServiceImpactInput{
		{
			ID:          uuid.New(),
			Name:        "Payments API",
			Criticality: CriticalityHigh,
			Status:      CurrentServiceStatusDegraded,
		},
		{
			ID:          uuid.New(),
			Name:        "Orders API",
			Criticality: CriticalityHigh,
			Status:      CurrentServiceStatusHealthy,
		},
	}

	first, firstAffected := AnalyzeJourneyImpact(
		journey,
		dependencies,
	)

	second, secondAffected := AnalyzeJourneyImpact(
		journey,
		dependencies,
	)

	if firstAffected != secondAffected {
		t.Fatal("affected result changed between executions")
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatalf(
			"impact analysis is not deterministic:\nfirst: %+v\nsecond: %+v",
			first,
			second,
		)
	}
}

func TestClassifyImpactSeverityThresholds(t *testing.T) {
	tests := []struct {
		name  string
		score int
		want  ImpactSeverity
	}{
		{
			name:  "zero is low",
			score: 0,
			want:  ImpactSeverityLow,
		},
		{
			name:  "below medium threshold is low",
			score: 39,
			want:  ImpactSeverityLow,
		},
		{
			name:  "medium threshold",
			score: 40,
			want:  ImpactSeverityMedium,
		},
		{
			name:  "below high threshold is medium",
			score: 59,
			want:  ImpactSeverityMedium,
		},
		{
			name:  "high threshold",
			score: 60,
			want:  ImpactSeverityHigh,
		},
		{
			name:  "below critical threshold is high",
			score: 79,
			want:  ImpactSeverityHigh,
		},
		{
			name:  "critical threshold",
			score: 80,
			want:  ImpactSeverityCritical,
		},
		{
			name:  "maximum score is critical",
			score: 100,
			want:  ImpactSeverityCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyImpactSeverity(tt.score)

			if got != tt.want {
				t.Fatalf(
					"severity for score %d = %q, want %q",
					tt.score,
					got,
					tt.want,
				)
			}
		})
	}
}
