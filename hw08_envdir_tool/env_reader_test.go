package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadDir(t *testing.T) {
	dir, err := os.MkdirTemp("", "envdir_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	testCases := []struct {
		name     string
		content  string
		expected EnvValue
	}{
		{"FOO", "123\n", EnvValue{Value: "123", NeedRemove: false}},
		{"BAR", "value", EnvValue{Value: "value", NeedRemove: false}},
		{"EMPTY", "", EnvValue{NeedRemove: true}},
		{"WITH_NULL", "first\x00second", EnvValue{Value: "first\nsecond", NeedRemove: false}},
		{"WITH_SPACES", "value   \t\n", EnvValue{Value: "value", NeedRemove: false}},
		{"INVALID_NAME=TEST", "test", EnvValue{}},
	}
	for _, tc := range testCases {
		filePath := filepath.Join(dir, tc.name)
		err := os.WriteFile(filePath, []byte(tc.content), 0o644)
		if err != nil {
			t.Fatal(err)
		}
	}
	env, err := ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	for _, tc := range testCases {
		if strings.Contains(tc.name, "=") {
			if _, exists := env[tc.name]; exists {
				t.Errorf("File with '=' in name should be skipped: %s", tc.name)
			}
			continue
		}
		val, exists := env[tc.name]
		if !exists {
			t.Errorf("Expected variable %s not found", tc.name)
			continue
		}
		if val.Value != tc.expected.Value || val.NeedRemove != tc.expected.NeedRemove {
			t.Errorf("For %s: expected %+v, got %+v", tc.name, tc.expected, val)
		}
	}
}

func TestReadDirWithSubdirectory(t *testing.T) {
	dir, err := os.MkdirTemp("", "envdir_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	subDir := filepath.Join(dir, "subdir")
	err = os.Mkdir(subDir, 0o755)
	if err != nil {
		t.Fatal(err)
	}
	testFiles := []struct {
		path    string
		content string
	}{
		{filepath.Join(dir, "VALID_FILE"), "valid_content"},
		{filepath.Join(subDir, "FILE_IN_SUBDIR"), "subdir_content"},
	}
	for _, tf := range testFiles {
		err := os.WriteFile(tf.path, []byte(tf.content), 0o644)
		if err != nil {
			t.Fatal(err)
		}
	}
	env, err := ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	if val, exists := env["VALID_FILE"]; !exists || val.Value != "valid_content" {
		t.Errorf("VALID_FILE not read correctly, got: %v", val)
	}
	if _, exists := env["FILE_IN_SUBDIR"]; exists {
		t.Error("FILE_IN_SUBDIR from subdirectory should not be included")
	}
	if _, exists := env["subdir"]; exists {
		t.Error("Subdirectory name should not be included")
	}
}
