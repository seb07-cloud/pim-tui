package ui

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "30 minutes",
			duration: 30 * time.Minute,
			expected: "30m",
		},
		{
			name:     "1 hour",
			duration: 1 * time.Hour,
			expected: "1h",
		},
		{
			name:     "1.5 hours",
			duration: 90 * time.Minute,
			expected: "1h30m",
		},
		{
			name:     "2 hours",
			duration: 2 * time.Hour,
			expected: "2h",
		},
		{
			name:     "2 hours 15 minutes",
			duration: 2*time.Hour + 15*time.Minute,
			expected: "2h15m",
		},
		{
			name:     "8 hours",
			duration: 8 * time.Hour,
			expected: "8h",
		},
		{
			name:     "45 minutes",
			duration: 45 * time.Minute,
			expected: "45m",
		},
		{
			name:     "5 minutes",
			duration: 5 * time.Minute,
			expected: "5m",
		},
		{
			name:     "0 minutes",
			duration: 0,
			expected: "0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.expected {
				t.Errorf("formatDuration(%v) = %q, want %q",
					tt.duration, got, tt.expected)
			}
		})
	}
}

func TestFormatCompactDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "30 seconds returns <1m",
			duration: 30 * time.Second,
			expected: "<1m",
		},
		{
			name:     "45 seconds returns <1m",
			duration: 45 * time.Second,
			expected: "<1m",
		},
		{
			name:     "59 seconds returns <1m",
			duration: 59 * time.Second,
			expected: "<1m",
		},
		{
			name:     "1 minute returns 1m",
			duration: 1 * time.Minute,
			expected: "1m",
		},
		{
			name:     "45 minutes returns 45m",
			duration: 45 * time.Minute,
			expected: "45m",
		},
		{
			name:     "59 minutes returns 59m",
			duration: 59 * time.Minute,
			expected: "59m",
		},
		{
			name:     "1 hour returns 1h",
			duration: 1 * time.Hour,
			expected: "1h",
		},
		{
			name:     "2h30m returns 2h30m",
			duration: 2*time.Hour + 30*time.Minute,
			expected: "2h30m",
		},
		{
			name:     "2h returns 2h",
			duration: 2 * time.Hour,
			expected: "2h",
		},
		{
			name:     "8h returns 8h",
			duration: 8 * time.Hour,
			expected: "8h",
		},
		{
			name:     "1h1m returns 1h1m",
			duration: 1*time.Hour + 1*time.Minute,
			expected: "1h1m",
		},
		{
			name:     "0 duration returns <1m",
			duration: 0,
			expected: "<1m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatCompactDuration(tt.duration)
			if got != tt.expected {
				t.Errorf("formatCompactDuration(%v) = %q, want %q",
					tt.duration, got, tt.expected)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		max      int
		expected string
	}{
		{
			name:     "short string unchanged",
			input:    "short",
			max:      10,
			expected: "short",
		},
		{
			name:     "long string truncated with ellipsis",
			input:    "this is a long string",
			max:      10,
			expected: "this is...",
		},
		{
			name:     "exactly at max length unchanged",
			input:    "exactly10!",
			max:      10,
			expected: "exactly10!",
		},
		{
			name:     "one over max truncated",
			input:    "exactly11!x",
			max:      10,
			expected: "exactly...",
		},
		{
			name:     "empty string unchanged",
			input:    "",
			max:      10,
			expected: "",
		},
		{
			name:     "very long string truncated",
			input:    "this is a very long string that should be truncated",
			max:      20,
			expected: "this is a very lo...",
		},
		{
			name:     "string at boundary minus 1",
			input:    "123456789",
			max:      10,
			expected: "123456789",
		},
		{
			name:     "max of 3 shows just ellipsis for long string",
			input:    "hello",
			max:      3,
			expected: "...",
		},
		{
			name:     "max of 4 shows one char plus ellipsis",
			input:    "hello",
			max:      4,
			expected: "h...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.max)
			if got != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q",
					tt.input, tt.max, got, tt.expected)
			}
		})
	}
}
