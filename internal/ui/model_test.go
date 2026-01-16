package ui

import (
	"strings"
	"testing"
)

func TestClampCursor(t *testing.T) {
	tests := []struct {
		name     string
		cursor   int
		delta    int
		length   int
		expected int
	}{
		{
			name:     "can't go below 0",
			cursor:   0,
			delta:    -1,
			length:   5,
			expected: 0,
		},
		{
			name:     "can't exceed length-1",
			cursor:   4,
			delta:    1,
			length:   5,
			expected: 4,
		},
		{
			name:     "normal movement up",
			cursor:   2,
			delta:    -1,
			length:   5,
			expected: 1,
		},
		{
			name:     "normal movement down",
			cursor:   2,
			delta:    1,
			length:   5,
			expected: 3,
		},
		{
			name:     "empty list edge case",
			cursor:   0,
			delta:    0,
			length:   0,
			expected: 0,
		},
		{
			name:     "empty list with positive delta",
			cursor:   0,
			delta:    1,
			length:   0,
			expected: 0,
		},
		{
			name:     "empty list with negative delta",
			cursor:   0,
			delta:    -1,
			length:   0,
			expected: 0,
		},
		{
			name:     "large positive delta clamped",
			cursor:   0,
			delta:    100,
			length:   5,
			expected: 4,
		},
		{
			name:     "large negative delta clamped",
			cursor:   4,
			delta:    -100,
			length:   5,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampCursor(tt.cursor, tt.delta, tt.length)
			if got != tt.expected {
				t.Errorf("clampCursor(%d, %d, %d) = %d, want %d",
					tt.cursor, tt.delta, tt.length, got, tt.expected)
			}
		})
	}
}

func TestIndexOf(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		val      int
		expected int
	}{
		{
			name:     "find 4 in middle",
			slice:    []int{1, 2, 4, 8},
			val:      4,
			expected: 2,
		},
		{
			name:     "find 1 at start",
			slice:    []int{1, 2, 4, 8},
			val:      1,
			expected: 0,
		},
		{
			name:     "find 8 at end",
			slice:    []int{1, 2, 4, 8},
			val:      8,
			expected: 3,
		},
		{
			name:     "not found returns 0",
			slice:    []int{1, 2, 4, 8},
			val:      5,
			expected: 0,
		},
		{
			name:     "empty slice returns 0",
			slice:    []int{},
			val:      1,
			expected: 0,
		},
		{
			name:     "find 2 in middle",
			slice:    []int{1, 2, 4, 8},
			val:      2,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := indexOf(tt.slice, tt.val)
			if got != tt.expected {
				t.Errorf("indexOf(%v, %d) = %d, want %d",
					tt.slice, tt.val, got, tt.expected)
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected LogLevel
	}{
		{
			name:     "debug returns LogDebug",
			input:    "debug",
			expected: LogDebug,
		},
		{
			name:     "error returns LogError",
			input:    "error",
			expected: LogError,
		},
		{
			name:     "info returns LogInfo",
			input:    "info",
			expected: LogInfo,
		},
		{
			name:     "empty string returns LogInfo (default)",
			input:    "",
			expected: LogInfo,
		},
		{
			name:     "unknown returns LogInfo (default)",
			input:    "unknown",
			expected: LogInfo,
		},
		{
			name:     "uppercase DEBUG returns LogInfo (case-sensitive)",
			input:    "DEBUG",
			expected: LogInfo,
		},
		{
			name:     "warning returns LogInfo (not recognized)",
			input:    "warning",
			expected: LogInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLogLevel(tt.input)
			if got != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v",
					tt.input, got, tt.expected)
			}
		})
	}
}

func TestValidateJustification(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
		errorSubstr string
	}{
		{
			name:        "valid reason passes",
			input:       "valid reason",
			expected:    "valid reason",
			expectError: false,
		},
		{
			name:        "empty string returns error",
			input:       "",
			expected:    "",
			expectError: true,
			errorSubstr: "required",
		},
		{
			name:        "whitespace only returns error",
			input:       "   ",
			expected:    "",
			expectError: true,
			errorSubstr: "required",
		},
		{
			name:        "tabs only returns error",
			input:       "\t\t",
			expected:    "",
			expectError: true,
			errorSubstr: "required",
		},
		{
			name:        "string with 501 chars returns error",
			input:       strings.Repeat("a", 501),
			expected:    "",
			expectError: true,
			errorSubstr: "exceeds 500",
		},
		{
			name:        "string with exactly 500 chars passes",
			input:       strings.Repeat("a", 500),
			expected:    strings.Repeat("a", 500),
			expectError: false,
		},
		{
			name:        "string with NUL char returns error",
			input:       "test\x00string",
			expected:    "",
			expectError: true,
			errorSubstr: "control",
		},
		{
			name:        "string with DEL char returns error",
			input:       "test\x7fstring",
			expected:    "",
			expectError: true,
			errorSubstr: "control",
		},
		{
			name:        "string with BEL char returns error",
			input:       "test\x07string",
			expected:    "",
			expectError: true,
			errorSubstr: "control",
		},
		{
			name:        "string with tabs is allowed",
			input:       "reason\twith\ttabs",
			expected:    "reason\twith\ttabs",
			expectError: false,
		},
		{
			name:        "string with newlines is allowed",
			input:       "reason\nwith\nnewlines",
			expected:    "reason\nwith\nnewlines",
			expectError: false,
		},
		{
			name:        "string with carriage returns is allowed",
			input:       "reason\rwith\rCR",
			expected:    "reason\rwith\rCR",
			expectError: false,
		},
		{
			name:        "leading and trailing whitespace trimmed",
			input:       "  trimmed reason  ",
			expected:    "trimmed reason",
			expectError: false,
		},
		{
			name:        "string with ESC char returns error",
			input:       "test\x1bstring",
			expected:    "",
			expectError: true,
			errorSubstr: "control",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateJustification(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("validateJustification(%q) error = nil, want error containing %q",
						tt.input, tt.errorSubstr)
					return
				}
				if !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("validateJustification(%q) error = %q, want error containing %q",
						tt.input, err.Error(), tt.errorSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("validateJustification(%q) error = %v, want nil", tt.input, err)
				return
			}

			if got != tt.expected {
				t.Errorf("validateJustification(%q) = %q, want %q",
					tt.input, got, tt.expected)
			}
		})
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected string
	}{
		{
			name:     "LogError returns ERROR",
			level:    LogError,
			expected: "ERROR",
		},
		{
			name:     "LogInfo returns INFO",
			level:    LogInfo,
			expected: "INFO",
		},
		{
			name:     "LogDebug returns DEBUG",
			level:    LogDebug,
			expected: "DEBUG",
		},
		{
			name:     "unknown level returns ERROR (default)",
			level:    LogLevel(99),
			expected: "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.level.String()
			if got != tt.expected {
				t.Errorf("LogLevel.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}
