package config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"testing"
)

// This is just my helper debug line, leaving it here for future use.
// t.Fatalf("Failed reading config file %s\n%T\n%+v", tmpFile, err, err)

func createTempFile(t *testing.T, content string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	_ = tmpFile.Close()
	return tmpFile.Name()
}

func TestReadConfig_FileNotFound(t *testing.T) {
	tmpFile := "nonexistent_file.json"
	_, err := readConfig(tmpFile)
	if err == nil {
		t.Fatalf("Expected error for missing file, got nil")
	}
	et := fs.ErrNotExist
	if !errors.Is(err, et) {
		t.Fatalf("Expected error type %v, %T", et, err)
	}
}

func TestReadConfig_InvalidJSON(t *testing.T) {
	tmpFile := createTempFile(t, `{"name": "test", "size": }`)
	defer func() { _ = os.Remove(tmpFile) }() // Yes, we are ignoring any errors

	_, err := readConfig(tmpFile)
	if err == nil {
		t.Fatalf("Expected JSON parse error, got nil")
	}
	var serr *json.SyntaxError
	if !errors.As(err, &serr) {
		t.Fatalf("Expected error type json.SyntaxError, %T", err)
	}
}

func TestReadConfig_EmptyFile(t *testing.T) {
	tmpFile := createTempFile(t, ``)
	defer func() { _ = os.Remove(tmpFile) }() // Yes, we are ignoring any errors

	_, err := readConfig(tmpFile)
	if err == nil {
		t.Fatalf("Expected error for empty file, got nil")
	}
	var serr *json.SyntaxError
	if !errors.As(err, &serr) {
		t.Fatalf("Expected error type json.SyntaxError, %T", err)
	}
}

func TestReadConfig_InvalidConfig(t *testing.T) {
	tmpFile := createTempFile(t,
		`{"file-path": "test",
	"ref-path": "y",
	"old-path": "z" }`)
	defer func() { _ = os.Remove(tmpFile) }() // Yes, we are ignoring any errors

	_, err := readConfig(tmpFile)
	if err == nil {
		t.Fatalf("Expected configuration to be invalid, blob-path is missing.")
	}
}

func TestReadConfig_Success(t *testing.T) {
	tmpFile := createTempFile(t,
		`{"file-path": "test",
	"blob-path": "y",
	"ref-path": "y",
	"old-path": "z" }`)
	defer func() { _ = os.Remove(tmpFile) }() // Yes, we are ignoring any errors

	_, err := readConfig(tmpFile)
	if err != nil {
		t.Fatalf("Configuration should be valid: %v", err)
	}
}
