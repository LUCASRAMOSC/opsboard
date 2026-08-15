package domain

import "testing"

func TestServiceTypeValid(t *testing.T) {
	tests := []struct {
		name        string
		serviceType ServiceType
		want        bool
	}{
		{
			name:        "frontend",
			serviceType: ServiceTypeFrontend,
			want:        true,
		},
		{
			name:        "api",
			serviceType: ServiceTypeAPI,
			want:        true,
		},
		{
			name:        "database",
			serviceType: ServiceTypeDatabase,
			want:        true,
		},
		{
			name:        "invalid",
			serviceType: ServiceType("INVALID"),
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.serviceType.Valid(); got != tt.want {
				t.Errorf("ServiceType.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCriticalityValid(t *testing.T) {
	tests := []struct {
		name        string
		criticality Criticality
		want        bool
	}{
		{
			name:        "low",
			criticality: CriticalityLow,
			want:        true,
		},
		{
			name:        "medium",
			criticality: CriticalityMedium,
			want:        true,
		},
		{
			name:        "high",
			criticality: CriticalityHigh,
			want:        true,
		},
		{
			name:        "critical",
			criticality: CriticalityCritical,
			want:        true,
		},
		{
			name:        "invalid",
			criticality: Criticality("INVALID"),
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.criticality.Valid(); got != tt.want {
				t.Errorf("Criticality.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}
