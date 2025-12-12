package runtime

import (
	"testing"

	"mrunc/pkg/specs"
)

func TestValidateUser(t *testing.T) {
	tests := []struct {
		name        string
		user        *specs.User
		expectError bool
	}{
		{
			name:        "nil user (optional)",
			user:        nil,
			expectError: false,
		},
		{
			name:        "valid root user",
			user:        &specs.User{UID: 0, GID: 0},
			expectError: false,
		},
		{
			name:        "valid non-root user",
			user:        &specs.User{UID: 1000, GID: 1000},
			expectError: false,
		},
		{
			name:        "valid with additional groups",
			user:        &specs.User{UID: 1000, GID: 1000, AdditionalGids: []int{100, 200}},
			expectError: false,
		},
		{
			name:        "invalid negative UID",
			user:        &specs.User{UID: -1, GID: 0},
			expectError: true,
		},
		{
			name:        "invalid UID too large",
			user:        &specs.User{UID: 70000, GID: 0},
			expectError: true,
		},
		{
			name:        "invalid negative GID",
			user:        &specs.User{UID: 0, GID: -1},
			expectError: true,
		},
		{
			name:        "invalid GID too large",
			user:        &specs.User{UID: 0, GID: 70000},
			expectError: true,
		},
		{
			name:        "invalid additional GID",
			user:        &specs.User{UID: 1000, GID: 1000, AdditionalGids: []int{100, -1}},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUser(tt.user)
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
