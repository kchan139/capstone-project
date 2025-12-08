package utils

import "testing"

func TestParseEnvKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple key-value",
			input:    "PATH=/usr/bin",
			expected: "PATH",
		},
		{
			name:     "key with no value",
			input:    "EMPTY=",
			expected: "EMPTY",
		},
		{
			name:     "no equals sign",
			input:    "NOEQUALS",
			expected: "NOEQUALS",
		},
		{
			name:     "multiple equals",
			input:    "KEY=VALUE=EXTRA",
			expected: "KEY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseEnvKey(tt.input)
			if result != tt.expected {
				t.Errorf("ParseEnvKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseEnvValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple key-value",
			input:    "PATH=/usr/bin",
			expected: "/usr/bin",
		},
		{
			name:     "empty value",
			input:    "EMPTY=",
			expected: "",
		},
		{
			name:     "no equals sign",
			input:    "NOEQUALS",
			expected: "",
		},
		{
			name:     "value with equals",
			input:    "KEY=VALUE=EXTRA",
			expected: "VALUE=EXTRA",
		},
		{
			name:     "complex path",
			input:    "PATH=/usr/local/bin:/usr/bin:/bin",
			expected: "/usr/local/bin:/usr/bin:/bin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseEnvValue(tt.input)
			if result != tt.expected {
				t.Errorf("ParseEnvValue(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
