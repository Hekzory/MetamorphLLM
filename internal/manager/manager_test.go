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
	
	if m.SuspiciousPath != "cmd/suspicious/main.go" {
		t.Errorf("Expected SuspiciousPath to be 'cmd/suspicious/main.go', got '%s'", m.SuspiciousPath)
	}
	
	if m.OutputPath != "cmd/suspicious/main.go.rewritten.go" {
		t.Errorf("Expected OutputPath to be 'cmd/suspicious/main.go.rewritten.go', got '%s'", m.OutputPath)
	}
	
	if m.TestTimeout != "30s" {
		t.Errorf("Expected TestTimeout to be '30s', got '%s'", m.TestTimeout)
	}
	
	if m.KeepRewritten != false {
		t.Errorf("Expected KeepRewritten to be false, got %v", m.KeepRewritten)
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
	suspDir := filepath.Join(tempDir, "cmd", "suspicious")
	if err := os.MkdirAll(suspDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	// Create test files to clean up
	testFiles := []string{
		filepath.Join(suspDir, "main.go.rewritten.go"),
		filepath.Join(suspDir, "main.go.backup"),
		filepath.Join(suspDir, "suspicious.backup"),
	}
	
	for _, file := range testFiles {
		if err := os.WriteFile(file, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}
	
	// Create a manager with test paths
	m := NewManager()
	m.SuspiciousPath = filepath.Join(suspDir, "main.go")
	m.OutputPath = filepath.Join(suspDir, "main.go.rewritten.go")
	
	// Run cleanup
	if err := m.CleanUp(); err != nil {
		t.Fatalf("CleanUp failed: %v", err)
	}
	
	// Verify files were deleted
	for _, file := range testFiles {
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			t.Errorf("Expected file %s to be deleted, but it still exists", file)
		}
	}
	
	// Test with KeepRewritten = true
	for _, file := range testFiles {
		if err := os.WriteFile(file, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}
	
	m.KeepRewritten = true
	if err := m.CleanUp(); err != nil {
		t.Fatalf("CleanUp failed: %v", err)
	}
	
	// Verify files were kept
	for _, file := range testFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file %s to be kept, but it was deleted", file)
		}
	}
}