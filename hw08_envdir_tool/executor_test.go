package main

import (
	"testing"
)

func TestRunCmd(t *testing.T) {
	tests := []struct {
		name     string
		cmd      []string
		env      Environment
		expected int
	}{
		{
			name:     "successful command",
			cmd:      []string{"echo", "hello"},
			env:      Environment{},
			expected: 0,
		},
		{
			name:     "command with env",
			cmd:      []string{"sh", "-c", "echo $TEST_VAR"},
			env:      Environment{"TEST_VAR": EnvValue{Value: "test_value", NeedRemove: false}},
			expected: 0,
		},
		{
			name:     "command with removed env",
			cmd:      []string{"sh", "-c", "echo $PATH"},
			env:      Environment{"PATH": EnvValue{NeedRemove: true}},
			expected: 0,
		},
		{
			name:     "nonexistent command",
			cmd:      []string{"nonexistent_command"},
			env:      Environment{},
			expected: 1,
		},
		{
			name:     "command with empty input",
			cmd:      []string{},
			env:      Environment{},
			expected: 1,
		},
		{
			name:     "return error from command",
			cmd:      []string{"ls", "-l", "/nonexistent/path"},
			env:      Environment{},
			expected: 1,
		},
		{
			name:     "return error AND args",
			cmd:      []string{"ls", "-l", "$TEST_VAR"},
			env:      Environment{"TEST_VAR": EnvValue{Value: "/nonexistent/path", NeedRemove: false}},
			expected: 1,
		},
		{
			name:     "bad_command",
			cmd:      []string{"bad_command"},
			env:      Environment{},
			expected: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RunCmd(tt.cmd, tt.env)
			if got != tt.expected {
				t.Errorf("RunCmd() = %v, want %v", got, tt.expected)
			}
		})
	}
}
