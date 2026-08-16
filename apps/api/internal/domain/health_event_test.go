package domain

import "testing"

func TestHealthStatusValid(t *testing.T) {
	tests := []struct {
		name   string
		status HealthStatus
		want   bool
	}{
		{
			name:   "healthy",
			status: HealthStatusHealthy,
			want:   true,
		},
		{
			name:   "degraded",
			status: HealthStatusDegraded,
			want:   true,
		},
		{
			name:   "unavailable",
			status: HealthStatusUnavailable,
			want:   true,
		},
		{
			name:   "invalid",
			status: HealthStatus("INVALID"),
			want:   false,
		},
		{
			name:   "empty",
			status: HealthStatus(""),
			want:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.status.Valid(); got != test.want {
				t.Fatalf(
					"HealthStatus.Valid() = %v, want %v",
					got,
					test.want,
				)
			}
		})
	}
}

func TestValidResponseTimeMs(t *testing.T) {
	zero := 0
	positive := 125
	negative := -1

	tests := []struct {
		name           string
		responseTimeMs *int
		want           bool
	}{
		{
			name:           "nil",
			responseTimeMs: nil,
			want:           true,
		},
		{
			name:           "zero",
			responseTimeMs: &zero,
			want:           true,
		},
		{
			name:           "positive",
			responseTimeMs: &positive,
			want:           true,
		},
		{
			name:           "negative",
			responseTimeMs: &negative,
			want:           false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ValidResponseTimeMs(test.responseTimeMs); got != test.want {
				t.Fatalf(
					"ValidResponseTimeMs() = %v, want %v",
					got,
					test.want,
				)
			}
		})
	}
}
