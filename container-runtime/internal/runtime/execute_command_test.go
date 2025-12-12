package runtime

import (
	"testing"
)

func TestPrepareExec(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		args        []string
		env         []string
		expectError bool
	}{
		{
			name:        "absolute path",
			command:     "/bin/sh",
			args:        []string{"/bin/sh", "-c", "echo test"},
			env:         []string{"PATH=/usr/bin:/bin"},
			expectError: false,
		},
		{
			name:        "command in PATH",
			command:     "sh",
			args:        []string{"sh", "-c", "echo test"},
			env:         []string{"PATH=/usr/bin:/bin"},
			expectError: false,
		},
		{
			name:        "command not found falls back to shell",
			command:     "nonexistent",
			args:        []string{"nonexistent"},
			env:         []string{"PATH=/usr/bin:/bin"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execPath, execArgs, err := PrepareExec(tt.command, tt.args, tt.env)

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				if execPath == "" {
					t.Error("Expected non-empty exec path")
				}
				if len(execArgs) == 0 {
					t.Error("Expected non-empty args")
				}
			}
		})
	}
}
