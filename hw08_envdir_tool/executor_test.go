package main

import (
	"testing"
)

func TestRunCmd(t *testing.T) {
	tests := []struct {
		name     string
		cmd      []string
		env      Environment
		expected []int
	}{
		{
			name:     "successful command",
			cmd:      []string{"echo", "hello"},
			env:      Environment{},
			expected: []int{0},
		},
		{
			name:     "command with env",
			cmd:      []string{"sh", "-c", "echo $TEST_VAR"},
			env:      Environment{"TEST_VAR": EnvValue{Value: "test_value", NeedRemove: false}},
			expected: []int{0},
		},
		{
			name:     "command with removed env",
			cmd:      []string{"sh", "-c", "echo $PATH"},
			env:      Environment{"PATH": EnvValue{NeedRemove: true}},
			expected: []int{0},
		},
		{
			name:     "nonexistent command",
			cmd:      []string{"nonexistent_command"},
			env:      Environment{},
			expected: []int{1},
		},
		{
			name:     "command with empty input",
			cmd:      []string{},
			env:      Environment{},
			expected: []int{1},
		},
		{
			name:     "return error from command",
			cmd:      []string{"ls", "-l", "/nonexistent/path"},
			env:      Environment{},
			expected: []int{1, 127, 2},
		},
		{
			name: "return error AND args",
			cmd:  []string{"ls", "-l", "$TEST_VAR"},
			env:  Environment{"TEST_VAR": EnvValue{Value: "/nonexistent/path", NeedRemove: false}},
			// как показал тест, для команды с ошибкой коды ошибок могут быть разные в зависимости от ОС
			expected: []int{1, 127, 2},
		},
		{
			name:     "bad_command",
			cmd:      []string{"bad_command"},
			env:      Environment{},
			expected: []int{1, 127, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RunCmd(tt.cmd, tt.env)
			matched := false
			for _, expectedCode := range tt.expected {
				if got == expectedCode {
					matched = true
					break
				}
			}
			if !matched {
				t.Errorf("RunCmd() = %v, want %v", got, tt.expected)
			}
		})
	}
}
