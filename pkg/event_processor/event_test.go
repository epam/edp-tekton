package event_processor

import "testing"

func TestContainsPipelineRecheckPrefix(t *testing.T) {
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
			result := ContainsPipelineRecheckPrefix(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

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
			name:     "Valid recheck in the middle",
			input:    "Some text /recheck more text",
			expected: true,
		},
		{
			name:     "Valid ok-to-test in the middle",
			input:    "Some text /ok-to-test more text",
			expected: true,
		},
		{
			name:     "Valid recheck at the end",
			input:    "Some text /recheck",
			expected: true,
		},
		{
			name:     "Valid ok-to-test at the end",
			input:    "Some text /ok-to-test",
			expected: true,
		},
		{
			name:     "Multiple recheck commands",
			input:    "/recheck /recheck",
			expected: true,
		},
		{
			name:     "Multiple ok-to-test commands",
			input:    "/ok-to-test /ok-to-test",
			expected: true,
		},
		{
			name:     "Mixed recheck and ok-to-test",
			input:    "/recheck /ok-to-test",
			expected: true,
		},
		{
			name:     "Case insensitive should not match",
			input:    "/Recheck",
			expected: false,
		},
		{
			name:     "Case insensitive ok-to-test should not match",
			input:    "/Ok-To-Test",
			expected: false,
		},
		{
			name:     "Partial match should not match",
			input:    "/rechec",
			expected: false,
		},
		{
			name:     "Partial ok-to-test should not match",
			input:    "/ok-to-tes",
			expected: false,
		},
		{
			name:     "Text without commands",
			input:    "This is a regular comment",
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "Whitespace only",
			input:    "   ",
			expected: false,
		},
		{
			name:     "Command with extra spaces",
			input:    "  /recheck  ",
			expected: true,
		},
		{
			name:     "Command with extra spaces ok-to-test",
			input:    "  /ok-to-test  ",
			expected: true,
		},
		{
			name:     "gerrit comment",
			input:    "Patch Set 2:\n\n/recheck",
			expected: true,
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
