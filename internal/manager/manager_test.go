package manager

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	
	if m.RewriterBinary != "rewriter" {
		t.Errorf("Expected RewriterBinary to be 'rewriter', got '%s'", m.RewriterBinary)
	}
	
	if m.SuspiciousPath != "internal/suspicious/suspicious.go" {
		t.Errorf("Expected SuspiciousPath to be 'internal/suspicious/suspicious.go', got '%s'", m.SuspiciousPath)
	}
	
	if m.OutputPath != "internal/suspicious/suspicious.go.rewritten.go" {
		t.Errorf("Expected OutputPath to be 'internal/suspicious/suspicious.go.rewritten.go', got '%s'", m.OutputPath)
	}
	
	if m.TestTimeout != "30s" {
		t.Errorf("Expected TestTimeout to be '30s', got '%s'", m.TestTimeout)
	}
	
	if m.KeepRewritten != true {
		t.Errorf("Expected KeepRewritten to be true, got %v", m.KeepRewritten)
	}
}

// TestCleanUp tests the cleanup functionality
func TestCleanUp(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "manager-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create test files
	suspDir := filepath.Join(tempDir, "internal", "suspicious")
	if err := os.MkdirAll(suspDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	// Create test files to clean up
	testFiles := []string{
		filepath.Join(suspDir, "suspicious.go.rewritten.go"),
		filepath.Join(suspDir, "suspicious.go.backup"),
		filepath.Join(filepath.Join(tempDir, "cmd", "suspicious"), "suspicious.backup"),
	}
	
	for _, file := range testFiles {
		// Make sure parent directory exists
		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			t.Fatalf("Failed to create directory for test file %s: %v", file, err)
		}
		if err := os.WriteFile(file, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}
	
	// Create a manager with test paths
	m := NewManager()
	m.SuspiciousPath = filepath.Join(suspDir, "suspicious.go")
	m.OutputPath = filepath.Join(suspDir, "suspicious.go.rewritten.go")
	m.TargetBinaryDir = filepath.Join(tempDir, "cmd", "suspicious")
	
	// Run cleanup
	if err := m.CleanUp(); err != nil {
		t.Fatalf("CleanUp failed: %v", err)
	}
	
	// Verify files were deleted or kept as expected
	for _, file := range testFiles {
		if filepath.Base(file) == "suspicious.go.rewritten.go" {
			// This file should be kept with default settings
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Errorf("Expected file %s to be kept, but it was deleted", file)
			}
		} else {
			// Backup files should be deleted
			if _, err := os.Stat(file); !os.IsNotExist(err) {
				t.Errorf("Expected file %s to be deleted, but it still exists", file)
			}
		}
	}
	
	// Test with KeepRewritten = false
	for _, file := range testFiles {
		// Make sure parent directory exists
		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			t.Fatalf("Failed to create directory for test file %s: %v", file, err)
		}
		if err := os.WriteFile(file, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}
	
	m.KeepRewritten = false
	if err := m.CleanUp(); err != nil {
		t.Fatalf("CleanUp failed: %v", err)
	}
	
	// All files should be deleted with KeepRewritten = false
	for _, file := range testFiles {
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			t.Errorf("Expected file %s to be deleted, but it still exists", file)
		}
	}
}