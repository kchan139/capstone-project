package runtime

import (
	"testing"

	"github.com/moby/sys/capability"
)

func TestParseCapability(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "valid cap with prefix",
			input:       "CAP_NET_BIND_SERVICE",
			expectError: false,
		},
		{
			name:        "valid cap without prefix",
			input:       "NET_BIND_SERVICE",
			expectError: false,
		},
		{
			name:        "CAP_KILL",
			input:       "CAP_KILL",
			expectError: false,
		},
		{
			name:        "CAP_AUDIT_WRITE",
			input:       "CAP_AUDIT_WRITE",
			expectError: false,
		},
		{
			name:        "invalid capability",
			input:       "CAP_INVALID_NOT_REAL",
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cap, err := parseCapability(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %q but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %q: %v", tt.input, err)
				}
				if cap == capability.Cap(0) && tt.input != "" {
					t.Errorf("Got zero capability for valid input %q", tt.input)
				}
			}
		})
	}
}
