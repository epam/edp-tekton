package event_processor

import "testing"

func TestContainsPipelineRecheck(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid recheck at the beginning",
			input:    "/recheck",
			expected: true,
		},
		{
			name:     "Valid ok-to-test at the beginning",
			input:    "/ok-to-test",
			expected: true,
		},
		{
			name:     "Invalid recheck not at the beginning",
			input:    "Some text /recheck",
			expected: false,
		},
		{
			name:     "Invalid ok-to-test not at the beginning",
			input:    "Some text /ok-to-test",
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsPipelineRecheck(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
