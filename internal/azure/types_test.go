package azure

import (
	"testing"
	"time"
)

// timePtr is a helper to create a pointer to a time.Time value
func timePtr(t time.Time) *time.Time {
	return &t
}

func TestStatusFromExpiry(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		expiry   *time.Time
		expected ActivationStatus
	}{
		{
			name:     "nil expiry returns StatusInactive",
			expiry:   nil,
			expected: StatusInactive,
		},
		{
			name:     "expiry in far future (>30min) returns StatusActive",
			expiry:   timePtr(now.Add(2 * time.Hour)),
			expected: StatusActive,
		},
		{
			name:     "expiry in <30min returns StatusExpiringSoon",
			expiry:   timePtr(now.Add(15 * time.Minute)),
			expected: StatusExpiringSoon,
		},
		{
			name:     "expiry exactly 30min returns StatusExpiringSoon (< not <=)",
			expiry:   timePtr(now.Add(30 * time.Minute)),
			expected: StatusExpiringSoon,
		},
		{
			name:     "expiry just over 30min returns StatusActive",
			expiry:   timePtr(now.Add(30*time.Minute + 1*time.Second)),
			expected: StatusActive,
		},
		{
			name:     "expiry just under 30min returns StatusExpiringSoon",
			expiry:   timePtr(now.Add(29*time.Minute + 59*time.Second)),
			expected: StatusExpiringSoon,
		},
		{
			name:     "expiry in past returns StatusExpiringSoon",
			expiry:   timePtr(now.Add(-5 * time.Minute)),
			expected: StatusExpiringSoon,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StatusFromExpiry(tt.expiry)
			if got != tt.expected {
				t.Errorf("StatusFromExpiry() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestActivationStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		status   ActivationStatus
		expected string
	}{
		{
			name:     "StatusActive returns Active",
			status:   StatusActive,
			expected: "Active",
		},
		{
			name:     "StatusExpiringSoon returns Expiring Soon",
			status:   StatusExpiringSoon,
			expected: "Expiring Soon",
		},
		{
			name:     "StatusPending returns Pending",
			status:   StatusPending,
			expected: "Pending",
		},
		{
			name:     "StatusInactive returns Inactive",
			status:   StatusInactive,
			expected: "Inactive",
		},
		{
			name:     "unknown status returns Inactive (default)",
			status:   ActivationStatus(99),
			expected: "Inactive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.String()
			if got != tt.expected {
				t.Errorf("ActivationStatus.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestActivationStatus_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   ActivationStatus
		expected bool
	}{
		{
			name:     "StatusActive returns true",
			status:   StatusActive,
			expected: true,
		},
		{
			name:     "StatusExpiringSoon returns true",
			status:   StatusExpiringSoon,
			expected: true,
		},
		{
			name:     "StatusPending returns false",
			status:   StatusPending,
			expected: false,
		},
		{
			name:     "StatusInactive returns false",
			status:   StatusInactive,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsActive()
			if got != tt.expected {
				t.Errorf("ActivationStatus.IsActive() = %v, want %v", got, tt.expected)
			}
		})
	}
}
