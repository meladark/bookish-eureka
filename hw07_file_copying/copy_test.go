package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCopy(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filecopy_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	content := []byte("Hello, this is a test file.")
	srcFilePath := filepath.Join(tmpDir, "input.txt")
	err = os.WriteFile(srcFilePath, content, 0o644)
	if err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	t.Run("Full copy", func(t *testing.T) {
		dst := filepath.Join(tmpDir, "copy_full.txt")
		err := Copy(srcFilePath, dst, 0, 0)
		if err != nil {
			t.Fatalf("copy failed: %v", err)
		}

		orig, _ := os.ReadFile(srcFilePath)
		copied, _ := os.ReadFile(dst)

		if string(orig) != string(copied) {
			t.Errorf("full copy: content mismatch")
		}
	})

	t.Run("Copy with offset", func(t *testing.T) {
		dst := filepath.Join(tmpDir, "copy_offset.txt")
		offset := int64(7)
		err := Copy(srcFilePath, dst, offset, 0)
		if err != nil {
			t.Fatalf("copy with offset failed: %v", err)
		}

		expected := content[offset:]
		result, _ := os.ReadFile(dst)

		if string(expected) != string(result) {
			t.Errorf("copy with offset: content mismatch")
		}
	})

	t.Run("Copy with limit", func(t *testing.T) {
		dst := filepath.Join(tmpDir, "copy_limit.txt")
		limit := int64(5)
		err := Copy(srcFilePath, dst, 0, limit)
		if err != nil {
			t.Fatalf("copy with limit failed: %v", err)
		}

		expected := content[:limit]
		result, _ := os.ReadFile(dst)

		if string(expected) != string(result) {
			t.Errorf("copy with limit: content mismatch")
		}
	})

	t.Run("Copy with offset and limit", func(t *testing.T) {
		dst := filepath.Join(tmpDir, "copy_offset_limit.txt")
		offset := int64(7)
		limit := int64(4)
		err := Copy(srcFilePath, dst, offset, limit)
		if err != nil {
			t.Fatalf("copy with offset+limit failed: %v", err)
		}

		expected := content[offset : offset+limit]
		result, _ := os.ReadFile(dst)

		if string(expected) != string(result) {
			t.Errorf("copy with offset+limit: content mismatch")
		}
	})

	t.Run("Offset exceeds file size", func(t *testing.T) {
		dst := filepath.Join(tmpDir, "copy_bad_offset.txt")
		offset := int64(len(content) + 10)
		err := Copy(srcFilePath, dst, offset, 0)

		if !errors.Is(err, ErrOffsetExceedsFileSize) {
			t.Fatalf("expected ErrOffsetExceedsFileSize, got: %v", err)
		}
	})

	t.Run("Unsupported file type (directory)", func(t *testing.T) {
		dst := filepath.Join(tmpDir, "copy_dir.txt")
		err := Copy(tmpDir, dst, 0, 0)
		if !errors.Is(err, ErrUnsupportedFile) {
			t.Fatalf("expected ErrUnsupportedFile, got: %v", err)
		}
	})

	t.Run("Copy empty file", func(t *testing.T) {
		emptyFile := filepath.Join(tmpDir, "empty.txt")
		os.WriteFile(emptyFile, []byte{}, 0o644)

		dst := filepath.Join(tmpDir, "copy_empty.txt")
		err := Copy(emptyFile, dst, 0, 0)
		if err != nil {
			t.Fatalf("copy empty file failed: %v", err)
		}

		stat, _ := os.Stat(dst)
		if stat.Size() != 0 {
			t.Errorf("copy empty file: expected size 0, got %d", stat.Size())
		}
	})

	t.Run("Stat error", func(t *testing.T) {
		nonExistent := filepath.Join(tmpDir, "no_such_file.txt")
		dst := filepath.Join(tmpDir, "copy_nowhere.txt")

		err := Copy(nonExistent, dst, 0, 0)
		if err == nil {
			t.Fatal("expected error for missing file, got nil")
		}
	})

	t.Run("Unsupported file type (/dev/urandom)", func(t *testing.T) {
		dst := filepath.Join(tmpDir, "copy_urandom.txt")
		err := Copy("/dev/urandom", dst, 0, 0)
		if !errors.Is(err, ErrUnsupportedFile) {
			t.Fatalf("expected ErrUnsupportedFile, got: %v", err)
		}
	})
}
